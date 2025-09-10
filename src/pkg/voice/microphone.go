package voice

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gordonklaus/portaudio"
)


type MicrophoneManager struct {
	configPath string
}


func NewMicrophoneManager() *MicrophoneManager {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting user config directory: %v", err)
		configDir = "."
	}
	return &MicrophoneManager{
		configPath: filepath.Join(configDir, "focushelper", "microphone.txt"),
	}
}


func (mm *MicrophoneManager) GetMicrophoneDevice(askForSelection bool, preferredIndex int, dockerMode bool) (*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}


	var inputDevices []*portaudio.DeviceInfo
	for _, dev := range devices {
		if dev.MaxInputChannels > 0 {
			inputDevices = append(inputDevices, dev)
		}
	}

	if len(inputDevices) == 0 {
		return nil, fmt.Errorf("no input devices found")
	}


	if dockerMode {
		device := inputDevices[0]
		log.Printf("Docker mode: using first available microphone: %s", device.Name)
		return device, nil
	}


	if preferredIndex >= 0 && preferredIndex < len(inputDevices) {
		device := inputDevices[preferredIndex]
		log.Printf("Using configured microphone: %s", device.Name)
		return device, nil
	}


	if !askForSelection {
		device := inputDevices[0]
		log.Printf("Using default microphone: %s", device.Name)
		mm.saveMicrophoneIndex(0)
		return device, nil
	}


	return mm.selectMicrophone(inputDevices)
}


func (mm *MicrophoneManager) selectMicrophone(devices []*portaudio.DeviceInfo) (*portaudio.DeviceInfo, error) {
	fmt.Println("\n🎤 Microphone Selection")
	fmt.Println("Available input devices:")
	for i, dev := range devices {
		fmt.Printf("[%d] %s (Input Channels: %d, DefaultSampleRate: %.0f)\n",
			i, dev.Name, dev.MaxInputChannels, dev.DefaultSampleRate)
	}

	var choice int
	for {
		fmt.Print("\nSelect input device by number (or press Enter for default): ")
		var input string
		fmt.Scanln(&input)


		if input == "" {
			choice = 0
			break
		}

		_, err := fmt.Sscanf(input, "%d", &choice)
		if err != nil || choice < 0 || choice >= len(devices) {
			fmt.Println("Invalid choice, try again.")
			continue
		}
		break
	}

	selectedDevice := devices[choice]
	log.Printf("Selected device: %s", selectedDevice.Name)


	mm.saveMicrophoneIndex(choice)

	return selectedDevice, nil
}


func (mm *MicrophoneManager) saveMicrophoneIndex(index int) {

	dir := filepath.Dir(mm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create config directory: %v", err)
		return
	}


	if err := os.WriteFile(mm.configPath, []byte(fmt.Sprintf("%d", index)), 0644); err != nil {
		log.Printf("Failed to save microphone selection: %v", err)
	}
}


func (mm *MicrophoneManager) loadMicrophoneIndex() int {
	data, err := os.ReadFile(mm.configPath)
	if err != nil {
		return -1 // Not set
	}

	var index int
	if _, err := fmt.Sscanf(string(data), "%d", &index); err != nil {
		return -1 // Invalid data
	}

	return index
}


func (mm *MicrophoneManager) isSampleRateSupported(device *portaudio.DeviceInfo, sampleRate float64) bool {

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: 1,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: 1024,
	}
	params.Output.Channels = 0


	err := portaudio.IsFormatSupported(params)
	return err == nil
}
