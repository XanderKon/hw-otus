package main

import (
	"fmt"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	setEnvironment(env)

	var cmdName string
	var cmdParams []string
	if len(cmd) > 1 {
		cmdName = cmd[0]
		cmdParams = cmd[1:]
	}

	cmdExec := exec.Command(cmdName, cmdParams...)

	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr

	err := cmdExec.Run()
	if err != nil {
		fmt.Println("there is an error with running cmd")
	}

	return cmdExec.ProcessState.ExitCode()
}

func setEnvironment(env Environment) {
	for key, value := range env {
		if value.NeedRemove {
			os.Unsetenv(key)
			continue
		}
		if _, exist := os.LookupEnv(key); exist {
			os.Unsetenv(key)
		}
		os.Setenv(key, value.Value)
	}
}
