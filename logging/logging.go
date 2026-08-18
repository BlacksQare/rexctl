package logging

import (
	"fmt"
	"io"
	"os"
)

var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

func LogInfo(format string, a ...any) {
	fmt.Fprintf(Stdout, "[\033[32mINFO\033[0m] "+format+"\n", a...)
}
func LogWarn(format string, a ...any) {
	fmt.Fprintf(Stderr, "[\033[33mWARN\033[0m] "+format+"\n", a...)
}
func LogErr(format string, a ...any) {
	fmt.Fprintf(Stderr, "[\033[31mERROR\033[0m] "+format+"\n", a...)
}

