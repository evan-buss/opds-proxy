package convert

import (
	"os/exec"
	"sync"
)

type KepubConverter struct {
	mutex sync.Mutex
}

func (kc *KepubConverter) Available() bool {
	path, err := exec.LookPath("kepubify")
	if err != nil {
		return false
	}
	return path != ""
}

func (kc *KepubConverter) Convert(input string, output string) error {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()
	cmd := exec.Command("kepubify", "-v", "-u", "-o", output, input)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
