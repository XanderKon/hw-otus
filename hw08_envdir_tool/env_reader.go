package main

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var (
	ErrReadingEnvironmentDir  = errors.New("cannot read target dir")
	ErrUnknowFileInfo         = errors.New("cannot get file info")
	ErrReadingEnvironmentFile = errors.New("cannot read environment file")
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	envFiles, err := os.ReadDir(dir)

	env := make(Environment)

	if err != nil {
		return nil, ErrReadingEnvironmentDir
	}

	for _, envFile := range envFiles {
		fileInfo, err := envFile.Info()
		if err != nil {
			return nil, ErrUnknowFileInfo
		}

		envVarName := fileInfo.Name()

		// skip reading env if file is empty.
		if fileInfo.Size() == 0 {
			env[envVarName] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}
			continue
		}

		// reading env.
		val, err := readEnvFromFile(fileInfo, dir)
		if err != nil {
			return nil, err
		}

		env[envVarName] = EnvValue{
			Value:      val,
			NeedRemove: false,
		}
	}

	return env, nil
}

func readEnvFromFile(envFile os.FileInfo, dir string) (string, error) {
	f, err := os.Open(dir + "/" + envFile.Name())
	if err != nil {
		f.Close()
		return "", ErrReadingEnvironmentFile
	}
	r := bufio.NewReader(f)
	val, _, err := r.ReadLine()
	if err != nil {
		f.Close()
		return "", ErrReadingEnvironmentFile
	}

	f.Close()
	return parseEnvValue(val), nil
}

// cut spaces and tabs and replace terminal null.
func parseEnvValue(value []byte) string {
	parsed := string(value)
	parsed = strings.ReplaceAll(parsed, "\x00", "\n")
	return strings.TrimRight(parsed, ` 	`)
}
