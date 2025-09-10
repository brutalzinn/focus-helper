package main

import (
	"context"
	"errors"
	"math/rand"

	"flag"
	"fmt"
	"focus-helper/src/pkg/actions"
	"focus-helper/src/pkg/activity"
	"focus-helper/src/pkg/config"
	"focus-helper/src/pkg/database"
	"focus-helper/src/pkg/language"
	"focus-helper/src/pkg/llm"
	logging "focus-helper/src/pkg/loggin"
	"focus-helper/src/pkg/mcp"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/notifications"
	"focus-helper/src/pkg/persona"
	"focus-helper/src/pkg/server"
	"focus-helper/src/pkg/sheduler"
	"focus-helper/src/pkg/state"
	"focus-helper/src/pkg/utils"
	"focus-helper/src/pkg/variables"
	"focus-helper/src/pkg/voice"
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
	mcpServer       *mcp.MCPServer
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

	sessionTimeout := components.appState.AppConfig.SessionTimeout.Duration
	if sessionTimeout == 0 {
		sessionTimeout = 1 * time.Hour
		log.Println("Session timeout not configured, using default: 1 hour")
	} else {
		log.Printf("Session timeout configured: %v", sessionTimeout)
	}

	if err := database.CleanupOldSessions(components.appState.DB, sessionTimeout); err != nil {
		log.Printf("Warning: Failed to cleanup old sessions: %v", err)
	}

	currentSession, err := database.GetCurrentSessionWithTimeout(components.appState.DB, sessionTimeout)
	if err != nil {
		log.Fatalf("Error checking for active session: %v", err)
	}

	if currentSession != nil {
		components.appState.CurrentSessionID = currentSession.ID
		components.appState.ContinuousUsageStartTime = currentSession.StartTime
		components.appState.LastActivityTime = time.Now()
		log.Printf("Resumed existing session (ID: %s) started at: %s",
			currentSession.ID, currentSession.StartTime.Format("2006-01-02 15:04:05"))
	} else {
		components.appState.ContinuousUsageStartTime = time.Now()
		components.appState.LastActivityTime = time.Now()
		log.Println("Starting fresh - no recent session to resume")
	}

	actions.Init(components.appState)
	setupCustomVariables(components.appState)
	startServices(ctx, &wg, components)

	<-sigChan
	log.Println("Interrupt signal received, initiating shutdown...")

	if components.appState.CurrentSessionID != "" {
		log.Println("Closing current session...")
		if err := database.EndSession(components.appState.DB, components.appState.CurrentSessionID); err != nil {
			log.Printf("Error closing session: %v", err)
		}
	}

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All services stopped gracefully. Exiting.")
	case <-time.After(10 * time.Second):
		log.Println("Shutdown timeout reached, forcing exit.")
	}
}

func loadConfiguration() (*models.Config, error) {
	debugFlag := flag.Bool("debug", false, "Enable debug mode for faster testing.")
	profileFlag := flag.String("profile", "default", "Profile name to load from profiles.json.")
	selectMicrophoneFlag := flag.Bool("select-microphone", false, "Force microphone selection dialog.")
	dockerModeFlag := flag.Bool("docker", false, "Enable Docker-compatible mode")
	mcpServerFlag := flag.Bool("mcp", false, "Enable MCP server")
	mcpPortFlag := flag.Int("mcp-port", 8089, "MCP server port")
	flag.Parse()

	if profileFlag == nil || *profileFlag == "" {
		return nil, errors.New("profile flag cannot be empty")
	}

	cfg, err := config.LoadConfig(*profileFlag, *debugFlag)
	if err != nil {
		return nil, err
	}

	if *selectMicrophoneFlag {
		cfg.AskForMicrophoneSelection = true
		cfg.MicrophoneDeviceIndex = -1
	}

	if *dockerModeFlag {
		cfg.DockerMode = true
		cfg.AskForMicrophoneSelection = false
	}

	if *mcpServerFlag {
		cfg.MCPServerEnabled = true
		cfg.MCPServerPort = *mcpPortFlag
	}

	return cfg, nil
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

	var mcpServer *mcp.MCPServer
	if appConfig.MCPServerEnabled {
		mcpServer = mcp.NewMCPServer(appState, appState.DB)
		log.Printf("MCP server initialized on port %d", appConfig.MCPServerPort)
	}

	return &appComponents{
		activityMonitor: activityMonitor,
		appState:        appState,
		mcpServer:       mcpServer,
	}, nil
}

func startServices(ctx context.Context, wg *sync.WaitGroup, c *appComponents) {
	defer wg.Done()
	wg.Add(1)
	go server.StartServer(ctx, wg)
	wg.Add(1)
	go c.activityMonitor.ActivityLoop(ctx, wg)

	if !c.appState.AppConfig.WellbeingQuestionsEnabled {
		log.Println("Wellbeing questions disabled.")
	} else {
		wg.Add(1)
		go sheduler.SchedulerLoop(ctx, wg, c.appState)
	}

	if !c.appState.AppConfig.ListenerEnabled {
		log.Println("Voice command listener is disabled in the config.")
	} else {
		err := portaudio.Initialize()
		if err != nil {
			log.Printf("Cant initliaze portaudio")
		}
		defer portaudio.Terminate()
		listener, err := voice.NewListener(c.appState)
		if err != nil {
			log.Fatalf("Failed to initialize voice listener: %v", err)
		}
		registerVoiceCommands(listener, c.appState)
		wg.Add(1)
		go listener.ListenContinuously(ctx, wg)
		log.Println("Voice listener is ready.")
	}

	if c.mcpServer == nil {
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.mcpServer.Start(c.appState.AppConfig.MCPServerPort); err != nil {
			log.Printf("MCP server error: %v", err)
		}
	}()
	log.Printf("MCP server started on port %d", c.appState.AppConfig.MCPServerPort)

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

	wg.Add(1)
	go ExecuteAfterRandomOddTimeAsync(ctx, wg, "minute", 1, 59, []models.ActionConfig{
		models.ActionConfig{
			Type: models.ActionYoutubeAudio,
			URL:  "https://www.youtube.com/watch?v=q47QIEua2eA",
		},
	})

	wg.Add(1)
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

// /We will move that for  a more simple way to register commands
// // but for now works and we should concern first into implement unit tests
// / and integration tests using mcp. Because this program needs be a tool for others llms agents.
func registerVoiceCommands(listener *voice.Listener, appState *state.AppState) {
	activationWords := strings.Split(appState.AppConfig.ActivationWord, ",")
	for i, word := range activationWords {
		activationWords[i] = strings.TrimSpace(word)
	}

	listener.RegisterWakeUpWord(func(ctx *voice.CommandContext) {
		wakeAction := models.ActionConfig{
			Type: models.ActionYoutubeAudio,
			///hardcoded urls for testing only porpuses
			URL: "https://www.youtube.com/watch?v=ECiqZE3JATI",
		}
		go actions.Execute(wakeAction)
		log.Println("App is now awake and ready to receive commands.")
	}, activationWords)

	listener.RegisterCommand(func(ctx *voice.CommandContext) {
		log.Println("MAYDAY DETECTED - Triggering Emergency Protocol")
		newAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: "Entendi sua solicitação. Vamos vetorizar você até o aeroporto mais próximo. Você deve confirmar com sim ou não.",
		}
		actions.Execute(newAction)
		go func() {
			select {
			case response := <-ctx.Response:
				if strings.Contains(response, "sim") {
					log.Println("User confirmed Mayday alert.")
					actions.Execute(models.ActionConfig{
						Type: models.ActionSpeak,
						Text: "Protocolo de emergência ativo.",
					})
					database.LogMaydayEvent(appState.DB)
				} else {
					log.Println("User canceled Mayday alert.")
					actions.Execute(models.ActionConfig{
						Type: models.ActionSpeak,
						Text: "Que bom que está tudo bem.",
					})
				}
			case <-time.After(15 * time.Second):
				log.Println("Timeout excedido.")
				actions.Execute(models.ActionConfig{
					Type: models.ActionSpeak,
					Text: "Alerta de mayday cancelado",
				})
			}
		}()

	}, strings.Split(appState.Language.Get("command_mayday_words"), ","))

	stopWords := strings.Split(appState.AppConfig.StopWord, ",")
	for i, word := range stopWords {
		stopWords[i] = strings.TrimSpace(word)
	}

	listener.RegisterCommand(func(ctx *voice.CommandContext) {
		startActions := []models.ActionConfig{
			{Type: models.ActionStop},
			{
				Type:      models.ActionSound,
				SoundFile: "airplane_communication_start.mp3",
			},
		}
		actions.ExecuteSequence(startActions)
	}, stopWords)

	listener.RegisterCommand(func(ctx *voice.CommandContext) {
		log.Println("Time request command detected.")
		timeAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appState.Language.Get("command_time"),
		}
		actions.Execute(timeAction)
	}, strings.Split(appState.Language.Get("command_time_words"), ","))

	listener.RegisterCommand(func(ctx *voice.CommandContext) {
		log.Println("Focus check command detected.")
		focusAction := models.ActionConfig{
			Type: models.ActionSpeak,
			Text: appState.Language.Get("command_focus"),
		}
		actions.Execute(focusAction)
	}, strings.Split(appState.Language.Get("command_focus_words"), ","))
}

func ExecuteAfterRandomOddTimeAsync(ctx context.Context, wg *sync.WaitGroup, timeUnit string, min, max int, actionPools []models.ActionConfig) {
	defer wg.Done()
	fmt.Printf("Ambient simulator started\n")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Ambient simulator stopped")
			return
		default:
			randomOdd := utils.GenerateRandomOdd(min, max)
			var waitDuration time.Duration
			switch timeUnit {
			case "minute":
				waitDuration = time.Duration(randomOdd) * time.Minute
			case "hour":
				waitDuration = time.Duration(randomOdd) * time.Hour
			default:
				waitDuration = time.Duration(randomOdd) * time.Minute
			}
			fmt.Printf("Scheduled to execute in %d %s\n", randomOdd, timeUnit)
			select {
			case <-ctx.Done():
				fmt.Println("Ambient simulator stopped during wait")
				return
			case <-time.After(waitDuration):
			}
			idx := rand.Intn(len(actionPools))
			nextAction := actionPools[idx]
			if nextAction.Type == models.ActionYoutubeAudio {
				startSecond, endSecond := utils.GenerateRandomSquareIntervalVaried(0, 59)
				nextAction.StartAt = utils.FormatTime(startSecond)
				nextAction.EndAt = utils.FormatTime(endSecond)
			}
			actions.Execute(nextAction)
			fmt.Println("Action completed, scheduling next execution...")
		}
	}
}
