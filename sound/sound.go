package sound

import (
	"os/exec"
	"runtime"
)

func Play() error {
	if runtime.GOOS == "darwin" {
		return exec.Command("afplay", "/System/Library/Sounds/Ping.aiff").Start()
	}
	return nil
}
