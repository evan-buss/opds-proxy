package convert

import (
	"fmt"
	"os/exec"
	"path/filepath"
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
		fmt.Println("TEST")
		path, err := exec.LookPath("kepubify")
		kc.available = err == nil && path != ""
	})

	return kc.available
}

func (kc *KepubConverter) Convert(input string) (string, error) {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	dir := filepath.Dir(input)
	kepubFile := filepath.Join(dir, strings.Replace(filepath.Base(input), ".epub", ".kepub.epub", 1))

	cmd := exec.Command("kepubify", "-v", "-u", "-o", kepubFile, input)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return kepubFile, nil
}
