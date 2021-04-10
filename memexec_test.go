package memexec

import (
	//	"bytes"
	//	"fmt"
	"github.com/k0kubun/pp"
	"io/ioutil"
	//	"os"
	"os/exec"
	//"syscall"
	"testing"
)

var (
	//TEST_CMD = `echo.static`
	TEST_CMD      = `echo`
	TEST_CMD_ARGS = []string{`-n`, `test`}
	//TEST_CMD      = `ls`
	//TEST_CMD_ARGS = []string{`/`}
	//TEST_CMD_ARGS = []string{`/`, `/2`}

)

func TestCommand(t *testing.T) {
	exe := newEchoExec(t)
	if false {
		pp.Println(exe)
	}
	defer func() {
		if err := exe.Close(); err != nil {
			t.Fatalf("close error: %s", err)
		}
	}()
	run_result := exe.Run(TEST_CMD_ARGS)
	pp.Println(run_result)
	if string(run_result.Stdout) != "test" {
		t.Errorf("command output = %q, want %q", string(run_result.Stdout), "test")
	}
}

func BenchmarkCommand(b *testing.B) {
	exe := newEchoExec(b)
	defer exe.Close()
	for i := 0; i < b.N; i++ {
		cmd := exe.Command("-n", "test")
		if _, err := cmd.Output(); err != nil {
			b.Fatal(err)
		}
	}
}

func newEchoExec(t testing.TB) *Exec {
	// lookup echo binary that is provided on all unix systems
	// and it's not a built-in opposed to `ls` and `type`
	path, err := exec.LookPath(TEST_CMD)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	exe, err := New(b)
	if err != nil {
		t.Fatal(err)
	}
	return exe
}
