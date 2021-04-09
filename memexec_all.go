//+build !linux

package memexec

import "os"

type executor struct {
	f             *os.File
	ProcExecution bool
	TmpPath       string
}

func (e *executor) prepare(t *os.File, PROC_EXECUTION bool) error {
	e.f = t
	e.ProcExecution = PROC_EXECUTION
	//e.ProcExecution = false
	return nil
}

func (e *executor) path() string {
	return e.f.Name()
}

func (e *executor) close() error {
	return os.Remove(e.f.Name())
}
