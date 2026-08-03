package logging

import (
	"fmt"
	"os"
)

func LogInfo(format string, a ...any) { fmt.Printf("[\033[32mINFO\033[0m] "+format+"\n", a...) }
func LogWarn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[\033[33mWARN\033[0m] "+format+"\n", a...)
}
func LogErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[\033[31mERROR\033[0m] "+format+"\n", a...)
}
