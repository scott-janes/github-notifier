package sound

import (
	"log"
	"os/exec"
	"runtime"
)

const defaultSoundPath = "/System/Library/Sounds/Ping.aiff"

func Play(soundPath string) error {
	if runtime.GOOS == "darwin" {
		if soundPath == "" {
			soundPath = defaultSoundPath
		}
		return exec.Command("afplay", soundPath).Start()
	}
	log.Printf("sound: no supported sound system for %s", runtime.GOOS)
	return nil
}
