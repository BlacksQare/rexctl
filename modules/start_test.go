package modules

import (
	"testing"
)

func TestCmdUp_MissingArgsDies(t *testing.T) {
	origDie := Die
	died := false
	Die = func(format string, a ...any) {
		died = true
		panic("died")
	}
	defer func() {
		Die = origDie
		if r := recover(); r == nil {
			t.Fatal("expected panic from Die when no args passed to CmdUp")
		}
		if !died {
			t.Error("expected Die to be called")
		}
	}()

	CmdUp([]string{})
}

func TestCmdStart_MissingArgsDies(t *testing.T) {
	origDie := Die
	died := false
	Die = func(format string, a ...any) {
		died = true
		panic("died")
	}
	defer func() {
		Die = origDie
		if r := recover(); r == nil {
			t.Fatal("expected panic from Die when no args passed to CmdStart")
		}
		if !died {
			t.Error("expected Die to be called")
		}
	}()

	CmdStart([]string{})
}
