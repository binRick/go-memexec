package memexec

import (
	"github.com/k0kubun/pp"
	"io/ioutil"
	"os/exec"
	"testing"
)

var (
	TEST_CMD = `echo`
	//TEST_CMD = `echo.static`
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

	c := exe.Command("-n", "test")
	o, err := c.Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(o) != "test" {
		t.Errorf("command output = %q, want %q", string(o), "test")
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
