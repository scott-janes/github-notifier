package sound

import (
	"log"
	"os/exec"
	"runtime"
)

func Play() error {
	if runtime.GOOS == "darwin" {
		return exec.Command("afplay", "/System/Library/Sounds/Ping.aiff").Start()
	}
	log.Printf("sound: no supported sound system for %s", runtime.GOOS)
	return nil
}
