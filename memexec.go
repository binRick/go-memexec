package memexec

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"

	"os/exec"
)

// Exec is an in-memory executable code unit.
type Exec struct {
	executor
	Path    string
	TmpPath string
	Hash    string
}

const (
	TEMP_FILE_PREFIX = `wgcs-`
)

var (
	DEBUG_MODE = false
)

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
	return &exe, nil
}

// Command is an equivalent of `exec.Command`,
// except that the path to the executable is be omitted.
func (m *Exec) Command(arg ...string) *exec.Cmd {
	if false {
		if DEBUG_MODE {
			fmt.Printf("Path=%v\n", m.Path)
			fmt.Printf("ProcExecution=%v\n", m.ProcExecution)
		}
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
