package audio

func PlaySound(filePath string, volume float64) error {

	release := RequestAccess()
	defer release()

	if volume <= 0 {
		volume = 1.0
	}
	err := playSoundAmplified(filePath, volume)
	return err
}
