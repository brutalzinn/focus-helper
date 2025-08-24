// cmd/focus-helper/main.go
package main

import (
	"context"
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
	"sync"
	"syscall"
	"time"

	"github.com/gordonklaus/portaudio"
	_ "github.com/mattn/go-sqlite3"
)

type appComponents struct {
	activityMonitor *activity.Activity
	appState        *state.AppState
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer utils.ClearTempAudioOnExit()
	defer portaudio.Terminate()
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

	components, err := initComponents(ctx, &wg, appConfig)
	if err != nil {
		log.Fatalf("Error initializing components: %v", err)
	}
	defer components.appState.DB.Close()
	actions.Init(components.appState)
	setupCustomVariables(components.appState)
	startServices(ctx, &wg, components)
	<-sigChan
	log.Println("Interrupt signal received, initiating shutdown...")
	cancel()
	wg.Wait()
	log.Println("All services stopped gracefully. Exiting.")
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

func initComponents(ctx context.Context, wg *sync.WaitGroup, appConfig *models.Config) (*appComponents, error) {
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
		Persona:      currentPersona,
		Language:     lm,
		LLMAdapter:   llmAdapter,
		VarProcessor: variablesProcessor,
		DB:           db,
		Notifier:     notifier,
		AppConfig:    appConfig,
	}
	appState := state.NewAppState(appStateDependencies)
	wg.Add(1)
	go appState.EventLoop(ctx, wg)
	fmt.Println("Event loop started in the background.")
	activityMonitor := activity.NewActivity(appState)
	return &appComponents{
		activityMonitor: activityMonitor,
		appState:        appState,
	}, nil
}

func startServices(ctx context.Context, wg *sync.WaitGroup, c *appComponents) {
	wg.Add(1)
	go server.StartServer(ctx, wg)
	wg.Add(1)
	go c.activityMonitor.ActivityLoop(ctx, wg)
	if c.appState.AppConfig.WellbeingQuestionsEnabled {
		wg.Add(1)
		go sheduler.SchedulerLoop(ctx, wg, c.appState)
	} else {
		log.Println("Wellbeing questions disabled.")
	}

	if c.appState.AppConfig.ListenerEnabled {
		err := portaudio.Initialize()
		if err != nil {
			log.Printf("Cant initliaze portaudio")
		}
		listener, err := voice.NewListener(c.appState)
		if err != nil {
			log.Fatalf("Failed to initialize voice listener: %v", err)
		}
		registerVoiceCommands(listener, c.appState)
		wg.Add(1)
		go listener.ListenContinuously(ctx, wg)
	} else {
		log.Println("Voice command listener is disabled in the config.")
	}

	startActions := []models.ActionConfig{
		{
			Type:      models.ActionSound,
			SoundFile: "airplane_communication_start.mp3",
		},
		{
			Type: models.ActionSpeak,
			Text: c.appState.Language.Get("hello_prompt"),
		},
	}
	go actions.ExecuteSequence(startActions)
}

func setupCustomVariables(appState *state.AppState) {
	appState.VarProcessor.RegisterHandler("level", func(context ...string) string {
		if appState.Hyperfocus != nil {
			return appState.Language.Get(appState.Hyperfocus.Level)
		}
		return appState.Language.Get("no_hyperfocus")
	})
	appState.VarProcessor.RegisterHandler("activity_duration", func(context ...string) string {
		usageDuration := time.Since(appState.ContinuousUsageStartTime)
		hoursUnit := appState.Language.Get("hour")
		minutesUnit := appState.Language.Get("minute")
		secondsUnit := appState.Language.Get("second")
		formatDuration := utils.FormatDuration(usageDuration, hoursUnit, minutesUnit, secondsUnit)
		return formatDuration

	})
	appState.VarProcessor.RegisterHandler("mode", func(context ...string) string {
		if appState.AppConfig.DEBUG {
			return appState.Language.Get("debug_on")
		}
		return appState.Language.Get("debug_off")
	})
	appState.VarProcessor.RegisterHandler("username", func(context ...string) string {
		return appState.AppConfig.Username
	})
	appState.VarProcessor.RegisterHandler("person", func(context ...string) string {
		return appState.Persona.GetName()
	})
	appState.VarProcessor.RegisterHandler("date", func(context ...string) string {
		now := time.Now()
		monthName := appState.Language.Get(fmt.Sprintf("months.%d", now.Month()))
		dateFormat := appState.Language.Get("date_format")
		result := strings.ReplaceAll(dateFormat, "{day}", fmt.Sprintf("%d", now.Day()))
		result = strings.ReplaceAll(result, "{month}", monthName)
		result = strings.ReplaceAll(result, "{year}", fmt.Sprintf("%d", now.Year()))
		return result
	})
	appState.VarProcessor.RegisterHandler("time", func(context ...string) string {
		loc, _ := time.LoadLocation(appState.AppConfig.TimeLocation)
		now := time.Now().In(loc)
		return now.Format(appState.Language.Get("time_format"))
	})
}
func registerVoiceCommands(listener *voice.Listener, appState *state.AppState) {

	wakeWord := listener.AppConfig().ActivationWord
	if wakeWord != "" {
		listener.RegisterWakeUpWord(func(text string) {
			wakeAction := models.ActionConfig{
				Type: models.ActionSpeak,
				Text: appState.Language.Get("command_ready"),
			}
			actions.Execute(wakeAction)
		}, strings.Split(appState.Language.Get("command_wakeup_words"), ","))
	}

	listener.RegisterCommand(func(text string) {
		startActions := []models.ActionConfig{
			{
				Type: models.ActionStop,
			},
			{
				Type:      models.ActionSound,
				SoundFile: "airplane_communication_start.mp3",
			},
		}
		go actions.ExecuteSequence(startActions)
	}, strings.Split(listener.AppConfig().StopWord, ","))

	listener.RegisterWakeUpWord(func(text string) {
		log.Println("MAYDAY DETECTED - Triggering Emergency Protocol")
		protocolMayday := models.ActionConfig{
			Type:   models.ActionSpeakIA,
			Prompt: appState.Language.Get("command_mayday"),
		}
		database.LogMaydayEvent(appState.DB)
		actions.Execute(protocolMayday)
	}, strings.Split(appState.Language.Get("command_mayday_words"), ","))

	listener.RegisterCommand(func(text string) {
		log.Println("Time request command detected.")
		timeAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appState.Language.Get("command_time"),
		}
		actions.Execute(timeAction)
	}, strings.Split(appState.Language.Get("command_time_words"), ","))

	listener.RegisterCommand(func(text string) {
		log.Println("Focus check command detected.")
		focusAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appState.Language.Get("command_focus"),
		}
		actions.Execute(focusAction)
	}, strings.Split(appState.Language.Get("command_focus_words"), ","))
}
