package convert

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func (mc *MobiConverter) Convert(input string) (string, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// KindleGen doesn't allow the input file to be in a different directory
	// So set the working directory to the input file.
	outDir, _ := filepath.Abs(filepath.Dir(input))
	mobiFile := filepath.Join(outDir, strings.Replace(filepath.Base(input), ".epub", ".mobi", 1))

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
		if exiterr, ok := err.(*exec.ExitError); ok {
			if exiterr.ExitCode() != 1 {
				fmt.Println(fmt.Sprint(err) + ": " + out.String() + ":" + stderr.String())
				return "", err
			}
		} else {
			fmt.Println(fmt.Sprint(err) + ": " + out.String() + ":" + stderr.String())
			return "", err
		}
	}

	return mobiFile, nil
}
