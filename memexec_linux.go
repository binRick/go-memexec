//+build linux

package memexec

import (
	"fmt"
	"github.com/k0kubun/pp"
	"github.com/u-root/u-root/pkg/ldd"
	"os"
)

type executor struct {
	f             *os.File
	ProcExecution bool
	TmpPath       string
	Hash          string
}

//type CachedExecs map[string]string

var (
	cached_execs = make(map[string]int)

//&CachedExecs{}
)

func init() {
}

// on linux we can keep a read only fd of the temp file and remove it,
// kernel buffers its content in memory until all fds are closed.
func (e *executor) prepare(t *os.File, hash_str string) error {
	f, err := os.OpenFile(t.Name(), os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
		}
	}()

	// check if /proc is mounted
	path := fmt.Sprintf("/proc/self/fd/%d", int(f.Fd()))
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s dosn't exist, probably /proc is not mounted", path)
		}
		return err
	}

	fmt.Printf("hash=%s\n", hash_str)
	//	pp.Println(e)
	//	pp.Println(t)

	ldd_qty := 0
	if cached_execs[hash_str] > 0 {
		fmt.Printf("cache hit: %s => %d\n", hash_str, cached_execs[hash_str])
		ldd_qty = cached_execs[hash_str]
	} else {
		fmt.Println("cache mmiss")
		files_list := []string{
			f.Name(),
		}
		ldd_info, err := ldd.Ldd(files_list)
		if err == nil {
			if DEBUG_MODE {
				fmt.Printf("ldd (%d)=%v\n", len(ldd_info), ldd_info)
			}
			if false {
				pp.Println(ldd_info)
			}
			ldd_qty = len(ldd_info)
			cached_execs[hash_str] = ldd_qty
		} else {
			if DEBUG_MODE {
				fmt.Printf("ldd failed\n")
			}
		}
	}

	pp.Println(cached_execs)

	if err != nil && ldd_qty > 1 {
		e.ProcExecution = true
	} else {
		e.ProcExecution = false
	}
	if DEBUG_MODE {
		fmt.Printf("ProcExecution? %v\n", e.ProcExecution)
	}
	if e.ProcExecution {
		if err = os.Remove(t.Name()); err != nil {
			return err
		}
	}

	e.f = f
	return nil
}

func (e *executor) path() string {
	return fmt.Sprintf("/proc/self/fd/%d", int(e.f.Fd()))
}

func (e *executor) close() error {
	if !e.ProcExecution {
		if err := os.Remove(e.f.Name()); err != nil {
			fmt.Printf("Failed to remove %s\n", e.f.Name())
		} else {
			if DEBUG_MODE {
				fmt.Printf("Removed %s\n", e.f.Name())
			}
		}
	}
	return e.f.Close()
}
