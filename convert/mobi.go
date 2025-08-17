package convert

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/evan-buss/opds-proxy/internal/formats"
)

type MobiConverter struct {
	mutex         sync.Mutex
	available     bool
	availableOnce sync.Once
}

func (mc *MobiConverter) Available() bool {
	mc.availableOnce.Do(func() {
		path, err := exec.LookPath("kindlegen")
		mc.available = err == nil && path != ""
	})
	return mc.available
}

func (mc *MobiConverter) HandlesInputFormat(format formats.Format) bool {
	return format == formats.EPUB
}

func (mc *MobiConverter) Convert(log *slog.Logger, input string) (string, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// KindleGen doesn't allow the input file to be in a different directory
	// So set the working directory to the input file.
	outDir, _ := filepath.Abs(filepath.Dir(input))
	mobiFile := filepath.Join(outDir, strings.Replace(filepath.Base(input), formats.EPUB.Extension, formats.MOBI.Extension, 1))

	// And remove the directory from file paths
	cmd := exec.Command("kindlegen",
		filepath.Base(input),
		"-dont_append_source", "-c1", "-o",
		filepath.Base(mobiFile),
	)
	cmd.Dir = outDir

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		isError := true
		if exiterr, ok := err.(*exec.ExitError); ok {
			// Sometimes warnings cause a 1 exit-code, but the file is still created
			log.Info("Exit code", slog.Any("code", exiterr.ExitCode()))
			if exiterr.ExitCode() == 1 {
				isError = false
			}
		}

		if isError {
			log.Error("Error converting file",
				slog.Any("error", err),
				slog.String("stdout", out.String()),
				slog.String("stderr", stderr.String()),
			)
			return "", fmt.Errorf("kindlegen conversion failed for %q: %w", input, err)
		}
	}

	return mobiFile, nil
}
