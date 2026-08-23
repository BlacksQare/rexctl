package modules

import (
	"testing"
)

func TestCmdBuild_MissingArgsDies(t *testing.T) {
	origDie := Die
	died := false
	Die = func(format string, a ...any) {
		died = true
		panic("died")
	}
	defer func() {
		Die = origDie
		if r := recover(); r == nil {
			t.Fatal("expected panic from Die when no args and no rex.yaml in cwd")
		}
		if !died {
			t.Error("expected Die to be called")
		}
	}()

	CmdBuild([]string{})
}
