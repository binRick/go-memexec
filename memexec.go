package memexec

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/k0kubun/pp"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Exec is an in-memory executable code unit.
type Exec struct {
	executor
	Path          string
	TmpPath       string
	Hash          string
	ProcExecution bool
}

const (
	TEMP_FILE_PREFIX = `wgcs-`
)

var (
	DEBUG_MODE = false
)

type RunResult struct {
	Command       string
	Path          string
	TmpPath       string
	Hash          string
	Arguments     []string
	Duration      time.Duration
	Stdout        string
	Stderr        string
	ExitCode      int64
	ProcExecution bool
}

func init() {
	debug_mode := os.Getenv("DEBUG_MODE")
	if debug_mode == `1` {
		DEBUG_MODE = true
	}

}

// New creates new memory execution object that can be
// used for executing commands on a memory based binary.
func New(b []byte) (*Exec, error) {
	f, err := ioutil.TempFile("", TEMP_FILE_PREFIX)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()

	// we need only read and execution privileges
	// ioutil.TempFile creates files with 0600 perms
	if err = os.Chmod(f.Name(), 0500); err != nil {
		return nil, err
	}
	if _, err := f.Write(b); err != nil {
		return nil, err
	}

	hash := md5.New()
	hash.Write(b)
	hash_str := fmt.Sprintf("%x", md5.Sum(b))
	if DEBUG_MODE {
		fmt.Printf("hash=%s\n", hash_str)
	}

	exe := Exec{
		Path:    f.Name(),
		TmpPath: f.Name(),
		Hash:    hash_str,
	}

	if err = exe.prepare(f, exe.Hash); err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	if DEBUG_MODE {
		pp.Println(exe)
	}
	return &exe, nil
}

func (m *Exec) Run(args []string) *RunResult {
	started := time.Now()
	var waitStatus syscall.WaitStatus
	var stdout, stderr bytes.Buffer
	rr := &RunResult{
		Path:          m.Path,
		TmpPath:       m.TmpPath,
		Hash:          m.Hash,
		Arguments:     args,
		Command:       m.path(),
		ProcExecution: m.ProcExecution,
	}

	c := m.Command(args...)

	c.Stdout = &stdout
	c.Stderr = &stderr
	if run_err := c.Run(); run_err != nil {
		if run_err != nil {
			if DEBUG_MODE {
				os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", run_err.Error()))
			}
		}
		if exitError, ok := run_err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			rr.ExitCode, _ = strconv.ParseInt(fmt.Sprintf(`%d`, waitStatus.ExitStatus()), 10, 32)
		}
	} else {
		// Success
		waitStatus = c.ProcessState.Sys().(syscall.WaitStatus)
		rr.ExitCode, _ = strconv.ParseInt(fmt.Sprintf(`%d`, waitStatus.ExitStatus()), 10, 32)
	}

	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if DEBUG_MODE {
		fmt.Printf("Exit Code: %d\n", rr.ExitCode)
		fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	}

	rr.Stdout = outStr
	rr.Stderr = errStr
	rr.Duration = time.Since(started)
	ms := fmt.Sprintf(`%s: `, rr.Command)
	if strings.HasPrefix(rr.Stderr, ms) {
		r1 := regexp.MustCompile(fmt.Sprintf(`^%s`, ms))
		rr.Stderr = r1.ReplaceAllString(rr.Stderr, ``)
	}
	return rr
}

// Command is an equivalent of `exec.Command`,
// except that the path to the executable is be omitted.
func (m *Exec) Command(arg ...string) *exec.Cmd {
	if DEBUG_MODE {
		fmt.Printf("Path=%v\n", m.Path)
		fmt.Printf("ProcExecution=%v\n", m.ProcExecution)
	}
	//exec_path := m.Path
	//	if m.ProcExecution {
	exec_path := m.path()
	//	}

	return exec.Command(exec_path, arg...)
}

// Close closes Exec object.
//
// Any further command will fail, it's client's responsibility
// to control the flow by using synchronization algorithms.
func (m *Exec) Close() error {
	return m.close()
}
