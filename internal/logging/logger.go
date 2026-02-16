package logging

import "fmt"

var Verbose bool

func Info(msg string) {
	fmt.Println(msg)
}

func Warn(msg string) {
	fmt.Println("[warning]", msg)
}

func Debug(msg string) {
	if Verbose {
		fmt.Println("[debug]", msg)
	}
}

func DebugErr(context string, err error) {
	if Verbose && err != nil {
		fmt.Printf("[debug] %s: %v\n", context, err)
	}
}