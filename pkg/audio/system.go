package audio

import (
	"log"
	"sync"
)

var SystemLock sync.Mutex

// RequestAccess locks the audio system for exclusive use.
// It returns a function that should be called to release the lock.
func RequestAccess() func() {
	log.Println("AUDIO_LOCK: Requesting access to audio system...")
	SystemLock.Lock()
	log.Println("AUDIO_LOCK: Access granted.")
	return func() {
		log.Println("AUDIO_LOCK: Releasing audio system.")
		SystemLock.Unlock()
	}
}
