package convert

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
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

	// KindleGen doesn't allow the input file to be in a different directory
	// So set the working directory to the input file.
	outDir, _ := filepath.Abs(filepath.Dir(input))

	// And remove the directory from file paths
	cmd := exec.Command("kindlegen",
		filepath.Base(input),
		"-dont_append_source", "-c1", "-o",
		filepath.Base(output),
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
				return err
			}
		} else {
			fmt.Println(fmt.Sprint(err) + ": " + out.String() + ":" + stderr.String())
			return err
		}
	}

	return nil
}
