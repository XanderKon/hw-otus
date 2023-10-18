package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	// создаем временную директорию.
	tempDir, _ := os.MkdirTemp(".", "unit-env")

	// создаем временные файлы
	tempFile, _ := os.CreateTemp(tempDir, "unit-env-test")
	tempFile.Write([]byte("BAR"))
	tempFile.Close()
	tempFileBaseName := path.Base(tempFile.Name())

	tempFile2, _ := os.CreateTemp(tempDir, "unit-env-test")
	tempFile2.Write([]byte("FOO"))
	tempFile2.Close()
	tempFile2BaseName := path.Base(tempFile2.Name())

	// создаем пустой файл
	tempFile3, _ := os.CreateTemp(tempDir, "unit-env-test")
	tempFile3.Close()
	tempFile3BaseName := path.Base(tempFile3.Name())

	tempFile4, _ := os.CreateTemp(tempDir, "unit-env-test")
	tempFile4.Write([]byte("FOO REPLACE	"))
	tempFile4.Close()
	tempFile4BaseName := path.Base(tempFile4.Name())

	t.Run("case without error", func(t *testing.T) {
		env, err := ReadDir(tempDir)
		require.Nil(t, err)

		_, ok := env[tempFileBaseName]
		_, ok2 := env[tempFile2BaseName]
		_, ok3 := env[tempFile3BaseName]

		require.Equal(t, ok, true)
		require.Equal(t, ok2, true)
		require.Equal(t, ok3, true)

		require.Equal(t, env[tempFileBaseName].Value, "BAR")
		require.Equal(t, env[tempFileBaseName].NeedRemove, false)
		require.Equal(t, env[tempFile2BaseName].Value, "FOO")
		require.Equal(t, env[tempFile2BaseName].NeedRemove, false)
		require.Equal(t, env[tempFile3BaseName].Value, "")
		require.Equal(t, env[tempFile3BaseName].NeedRemove, true)
	})

	t.Run("case without error with replace", func(t *testing.T) {
		env, err := ReadDir(tempDir)
		require.Nil(t, err)

		_, ok := env[tempFile4BaseName]
		require.Equal(t, ok, true)

		require.Equal(t, env[tempFile4BaseName].Value, "FOO REPLACE")
		require.Equal(t, env[tempFile4BaseName].NeedRemove, false)
	})

	t.Run("case with error", func(t *testing.T) {
		_, err := ReadDir("1")
		require.NotNil(t, err)
		require.EqualError(t, err, ErrReadingEnvironmentDir.Error())
	})

	// удаляем временную директорию.
	defer os.RemoveAll(tempDir)
}
