package zfs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"github.com/docker/docker/pkg/log"
)

/*
 * TODO: Add infrastructire to enable/disable debugging; inspired by
 * http://play.golang.org/p/mOSbdHwSYR
 */

/*
 * Check if the slice contains a string
 */
func sliceContainsString(list []string, a string) (bool) {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

/*
 * Check if ZFS is supported
 */
func supportsZFS() error {
	f, err := os.Open("/proc/filesystems")
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		words := strings.Fields(s.Text())
		if sliceContainsString(words, "zfs") {
			return nil
		}
	}
	return fmt.Errorf("ZFS was not found in /proc/filesystems")
}

func dbg(format string, a ... interface{}) {
	log.Debugf("[zfs] " + format, a...)
}

/*
 * TODO: Make calls of TrimSPace() optional, dependent on a parameter which is on
 * by default.
 *
 * TODO: Take an optional param with a default of true, to let the caller choose if
 * they want to trim the stdout/stderr strings before being returned.
 *
 * TODO: Calculate and log how long each command execution takes.
 */
func execCmd(name string, args ... string) (string, string, error) {
	cmd := exec.Command(name, args...)
	dbg("Command: %v", cmd)
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()

	/*
	 * Trim the output/error strings to remove any spurious space and trailing
	 * new-lines.
	 */
	outString := strings.TrimSpace(outBuf.String())
	errString := strings.TrimSpace(errBuf.String())

	if outString != "" {
		dbg("outStream: %s", outString)
	}

	if errString != "" {
		dbg("errStream: %s", errString)
	}

	if err != nil {
		dbg("error: %v", err)
	}

	return outString, errString, err
}

func funcEnter() string {

	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	dbg("Entering: %s", funcName)
	return funcName
}

func funcLeave(funcName string) {
	dbg("Leaving: %s", funcName)
}

func (d *Driver) getDataset(id string) string {
	return path.Join(d.root_dataset, id)
}

func (d *Driver) getPath(id string) string {
	return path.Join(d.root_mountpoint, id)
}
