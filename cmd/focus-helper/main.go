// cmd/focus-helper/main.go
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"focus-helper/pkg/actions"
	"focus-helper/pkg/activity"
	"focus-helper/pkg/audio"
	"focus-helper/pkg/config"
	"focus-helper/pkg/database"
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	logging "focus-helper/pkg/loggin"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/server"
	"focus-helper/pkg/sheduler"
	"focus-helper/pkg/state"
	"focus-helper/pkg/utils"
	"focus-helper/pkg/variables"
	"focus-helper/pkg/voice"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gordonklaus/portaudio"
	_ "github.com/mattn/go-sqlite3"
)

type appComponents struct {
	actionExecutor     *actions.Executor
	activityMonitor    *activity.Monitor
	appState           *state.AppState
	db                 *sql.DB
	notifier           notifications.Notifier
	variablesProcessor *variables.Processor
}

func main() {
	defer utils.ClearTempAudioOnExit()

	appConfig, err := loadConfiguration()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	logging.SetupLogger(appConfig)

	log.Println("--- Starting focus helper ---")
	log.Printf("PERSONA ACTIVE: %s", appConfig.PersonaName)
	if appConfig.DEBUG {
		log.Println("!!!!!!!!!! RUNNING IN DEBUG MODE !!!!!!!!!!")
	}

	components, err := initComponents(appConfig)
	if err != nil {
		log.Fatalf("Error initializing components: %v", err)
	}
	defer components.db.Close()

	setupCustomVariables(components.variablesProcessor, components.appState, appConfig)

	startServices(appConfig, components)

	waitForShutdownSignal()
	// portaudio.Terminate()
	log.Println("Interrupt signal received, stopping all.")
}

func loadConfiguration() (*models.Config, error) {
	debugFlag := flag.Bool("debug", false, "Enable debug mode for faster testing.")
	profileFlag := flag.String("profile", "default", "Profile name to load from profiles.json.")
	flag.Parse()

	if profileFlag == nil || *profileFlag == "" {
		return nil, errors.New("profile flag cannot be empty")
	}

	return config.LoadConfig(*profileFlag, *debugFlag)
}

func initComponents(appConfig *models.Config) (*appComponents, error) {
	db, err := database.Init(appConfig.DatabaseFile)
	if err != nil {
		return nil, err
	}

	notifier := notifications.NewDesktopNotifier()
	llmAdapter, err := llm.NewAdapter(appConfig.IAModel)
	if err != nil {
		return nil, err
	}

	variablesProcessor := variables.NewProcessor()

	currentPersona, err := persona.GetPersona(appConfig.PersonaName, variablesProcessor)
	if err != nil {
		return nil, err
	}

	langsPath := filepath.Join(config.GetUserConfigPath(), "langs")

	lm, err := language.NewManager(langsPath, appConfig.PersonaName, appConfig.Language)
	if err != nil {
		log.Print("faild to load language manager")
		return nil, err
	}

	appStateDependencies := state.AppStateDependencies{
		Persona:       currentPersona,
		Language:      lm,
		LLMAdapter:    &llmAdapter,
		TextProcessor: variablesProcessor,
	}
	appState := state.NewAppState(appStateDependencies)
	go appState.EventLoop()
	executorDeps := actions.ExecutorDependencies{
		AppConfig:    appConfig,
		AppState:     appState,
		VarProcessor: variablesProcessor,
	}
	actionExecutor := actions.NewExecutor(executorDeps)

	activityMonitorDeps := activity.MonitorDependencies{
		DB:             db,
		ActionExecutor: actionExecutor,
		AppState:       appState,
		LLMAdapter:     llmAdapter,
		AppConfig:      appConfig,
	}
	activityMonitor := activity.NewMonitor(activityMonitorDeps)

	return &appComponents{
		actionExecutor:     actionExecutor,
		activityMonitor:    activityMonitor,
		appState:           appState,
		db:                 db,
		notifier:           notifier,
		variablesProcessor: variablesProcessor,
	}, nil
}

func startServices(appConfig *models.Config, c *appComponents) {
	go server.StartServer()
	go c.activityMonitor.MonitorActivityLoop()

	if appConfig.WellbeingQuestionsEnabled {
		go sheduler.SchedulerLoop(appConfig, c.db, c.actionExecutor, c.notifier)
	} else {
		log.Println("Wellbeing questions disabled.")
	}

	if appConfig.ListenerEnabled {
		portaudio.Initialize()
		listener, err := voice.NewListener(appConfig, c.appState)
		if err != nil {
			log.Fatalf("Failed to initialize voice listener: %v", err)
		}
		c.appState.IsListening
		registerVoiceCommands(listener, c)
		go listener.ListenContinuously()
	} else {
		log.Println("Voice command listener is disabled in the config.")
	}

	welcomeAction := models.ActionConfig{
		Type: models.ActionSpeak,
		Text: c.appState.Language.Get("hello_prompt"),
	}
	go c.actionExecutor.Execute(welcomeAction)
}

func waitForShutdownSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}

func setupCustomVariables(processor *variables.Processor, appState *state.AppState, appConfig *models.Config) {
	processor.RegisterHandler("level", func(context ...string) string {
		if appState.Hyperfocus != nil {
			return appState.Language.Get(appState.Hyperfocus.Level)
		}
		return appState.Language.Get("no_hyperfocus")
	})
	processor.RegisterHandler("activity_duration", func(context ...string) string {
		usageDuration := time.Since(appState.ContinuousUsageStartTime)
		return utils.FormatDuration(usageDuration)
	})
	processor.RegisterHandler("mode", func(context ...string) string {
		if appConfig.DEBUG {
			return appState.Language.Get("debug_on")
		}
		return appState.Language.Get("debug_off")
	})
	processor.RegisterHandler("username", func(context ...string) string {
		return appConfig.Username
	})
	processor.RegisterHandler("person", func(context ...string) string {
		return appState.Persona.GetName()
	})
	processor.RegisterHandler("date", func(context ...string) string {
		now := time.Now()
		monthName := appState.Language.Get(fmt.Sprintf("months.%d", now.Month()))
		dateFormat := appState.Language.Get("date_format")
		result := strings.ReplaceAll(dateFormat, "{day}", fmt.Sprintf("%d", now.Day()))
		result = strings.ReplaceAll(result, "{month}", monthName)
		result = strings.ReplaceAll(result, "{year}", fmt.Sprintf("%d", now.Year()))
		return result
	})
	processor.RegisterHandler("time", func(context ...string) string {
		loc, _ := time.LoadLocation(appConfig.TimeLocation)
		now := time.Now().In(loc)
		return now.Format(appState.Language.Get("time_format"))
	})
}
func registerVoiceCommands(listener *voice.Listener, appComponent *appComponents) {

	wakeWord := listener.AppConfig().ActivationWord
	if wakeWord != "" {
		listener.RegisterCommand(func(text string) {
			wakeAction := models.ActionConfig{
				Type: models.ActionSpeak,
				Text: appComponent.appState.Language.Get("command_ready"),
			}
			appComponent.actionExecutor.Execute(wakeAction)
		}, wakeWord, "torre", "comand", "comanda")
	}
	stopWord := listener.AppConfig().StopWord
	listener.RegisterCommand(func(text string) {
		confirmStop := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appComponent.appState.Language.Get("command_stop"),
		}
		appComponent.actionExecutor.Execute(confirmStop)
		audio.StopCurrentSound()
		stopAction := models.ActionConfig{
			Type: models.ActionStop,
		}
		appComponent.actionExecutor.Execute(stopAction)
	}, stopWord, "parar", "cancel", "stop", "para", "para!")

	listener.RegisterCommand(func(text string) {
		log.Println("MAYDAY DETECTED - Triggering Emergency Protocol")
		protocolMayday := models.ActionConfig{
			Type:   models.ActionSpeakIA,
			Prompt: appComponent.appState.Language.Get("command_mayday"),
		}
		appComponent.actionExecutor.Execute(protocolMayday)
	}, "mayday", "emergencia")

	listener.RegisterCommand(func(text string) {
		log.Println("Time request command detected.")
		timeAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appComponent.appState.Language.Get("command_time"),
		}
		appComponent.actionExecutor.Execute(timeAction)
	}, "tempo", "time", "que horas são")

	listener.RegisterCommand(func(text string) {
		log.Println("Focus check command detected.")
		focusAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appComponent.appState.Language.Get("command_focus"),
		}
		appComponent.actionExecutor.Execute(focusAction)
	}, "check", "checagem", "checar")
}
