// cmd/focus-helper/main.go
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"focus-helper/pkg/actions"
	"focus-helper/pkg/activity"
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
		return nil, err
	}

	appState := state.NewAppState()
	appState.Persona = currentPersona
	appState.Language = lm
	appState.LLMAdapter = &llmAdapter

	executorDeps := actions.ExecutorDependencies{
		AppConfig:    appConfig,
		AppState:     appState,
		VarProcessor: variablesProcessor,
		Notifier:     notifier,
		LLMAdapter:   llmAdapter,
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
		listener, err := voice.NewListener(appConfig, c.appState)
		if err != nil {
			log.Fatalf("Failed to initialize voice listener: %v", err)
		}
		defer listener.Close()

		registerVoiceCommands(listener, c.actionExecutor)

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
		return appState.Language.Get(appState.Hyperfocus.Level)
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
func registerVoiceCommands(listener *voice.Listener, executor *actions.Executor) {
	maydayWord := listener.AppConfig().ActivationWord
	if maydayWord != "" {
		listener.RegisterCommand(maydayWord, func(text string) {
			log.Println("MAYDAY DETECTED - Triggering Emergency Protocol")
			protocolMayday := models.ActionConfig{
				Type:   models.ActionSpeakIA,
				Prompt: "You must guide the pilot to a safe landing immediately. He is in an emergency and may lose all engines.",
			}
			go executor.Execute(protocolMayday)
		})
	}

	listener.RegisterCommand("what time is it", func(text string) {
		log.Println("Time request command detected.")
		timeAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: "The current time is {time}.",
		}
		go executor.Execute(timeAction)
	})

	listener.RegisterCommand("check my focus", func(text string) {
		log.Println("Focus check command detected.")
		focusAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: "Your current focus level is {level}. You have been active for {activity_duration}.",
		}
		go executor.Execute(focusAction)
	})
}
