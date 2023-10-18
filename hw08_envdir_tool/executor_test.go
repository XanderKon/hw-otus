package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	env := make(Environment)

	env["FOO"] = EnvValue{
		Value:      "TEST",
		NeedRemove: false,
	}
	env["BAR"] = EnvValue{
		Value:      "HELLO",
		NeedRemove: false,
	}

	cmd := make([]string, 3)
	cmd[0] = "/bin/bash"
	cmd[1] = "./testdata/echo.sh"
	cmd[2] = "arg1=test1 arg2=test2"

	t.Run("case without error", func(t *testing.T) {
		exitCode := RunCmd(cmd, env)

		val, exist := os.LookupEnv("FOO")
		require.Equal(t, val, "TEST")
		require.Equal(t, exist, true)

		val, exist = os.LookupEnv("BAR")
		require.Equal(t, val, "HELLO")
		require.Equal(t, exist, true)

		require.Equal(t, exitCode, 0)
	})

	t.Run("case with error", func(t *testing.T) {
		exitCode := RunCmd([]string{}, env)
		require.Equal(t, exitCode, -1)

		exitCode = RunCmd([]string{"/bin/bash", "test"}, env)
		require.Equal(t, exitCode, 126)

		exitCode = RunCmd([]string{"/bin/bash", "<1"}, env)
		require.Equal(t, exitCode, 127)
	})
}
