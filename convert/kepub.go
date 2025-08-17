package convert

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

	"github.com/evan-buss/opds-proxy/internal/formats"
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

func (kc *KepubConverter) HandlesInputFormat(format formats.Format) bool {
	return format == formats.EPUB
}

func (kc *KepubConverter) Convert(_ *slog.Logger, input string) (string, error) {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	kepubFile := strings.Replace(input, formats.EPUB.Extension, formats.KEPUB.Extension, 1)

	cmd := exec.Command("kepubify", "-v", "-u", "-o", kepubFile, input)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("kepubify conversion failed for %q: %w", input, err)
	}

	return kepubFile, nil
}
