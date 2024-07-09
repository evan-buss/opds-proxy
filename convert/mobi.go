package convert

import (
	"os/exec"
	"sync"
)

type MobiConverter struct {
	mutex sync.Mutex
}

func (mc *MobiConverter) Available() bool {
	path, err := exec.LookPath("kindlegen")
	if err != nil {
		return false
	}
	return path != ""
}

func (mc *MobiConverter) Convert(input string, output string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	return nil
}
