package main

import (
	"fmt"
	"strings"
)

func windowsBatchCommandLine(scriptPath string, args []string) (string, error) {
	if strings.ContainsAny(scriptPath, "\r\n\"") {
		return "", fmt.Errorf("unsafe Claude CLI path")
	}
	var command strings.Builder
	command.WriteString(`/D /S /C ""`)
	command.WriteString(scriptPath)
	command.WriteString(`"`)
	for _, arg := range args {
		if strings.ContainsAny(arg, "\r\n\"") {
			return "", fmt.Errorf("unsafe Claude CLI argument")
		}
		command.WriteString(` "`)
		command.WriteString(arg)
		command.WriteString(`"`)
	}
	command.WriteString(`"`)
	return command.String(), nil
}
