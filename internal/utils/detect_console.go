package utils

import "runtime"

func DetectDefaultConsole() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "sh"
}
