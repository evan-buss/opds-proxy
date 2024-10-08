package convert

import (
	"log/slog"
	"os/exec"
	"strings"
	"sync"
)

type KepubConverter struct {
	mutex         sync.Mutex
	available     bool
	availableOnce sync.Once
}

func (kc *KepubConverter) Available() bool {
	kc.availableOnce.Do(func() {
		path, err := exec.LookPath("kepubify")
		kc.available = err == nil && path != ""
	})

	return kc.available
}

func (kc *KepubConverter) Convert(_ *slog.Logger, input string) (string, error) {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	kepubFile := strings.Replace(input, ".epub", ".kepub.epub", 1)

	cmd := exec.Command("kepubify", "-v", "-u", "-o", kepubFile, input)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return kepubFile, nil
}
