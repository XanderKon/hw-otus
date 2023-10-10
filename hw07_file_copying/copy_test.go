package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	// Создаем временный файл.
	tempFile, _ := os.CreateTemp(".", "unittest-")
	testTargetPath := tempFile.Name()

	testSourcePath := "testdata/input.txt"
	testSourcePathWithOffsetAndLimit := "testdata/out_offset100_limit1000.txt"

	t.Run("case with original file", func(t *testing.T) {
		err := Copy(testSourcePath, testTargetPath, 0, 0)
		require.Nil(t, err)

		origBytes, _ := os.ReadFile(testSourcePath)
		targetBytes, err := os.ReadFile(testTargetPath)

		require.Nil(t, err)
		require.Equal(t, origBytes, targetBytes)
	})

	t.Run("case with offset and limit", func(t *testing.T) {
		err := Copy(testSourcePath, testTargetPath, 100, 1000)
		require.Nil(t, err)

		origBytes, _ := os.ReadFile(testSourcePathWithOffsetAndLimit)
		targetBytes, err := os.ReadFile(testTargetPath)

		require.Nil(t, err)
		require.Equal(t, origBytes, targetBytes)
	})

	t.Run("case with offset error", func(t *testing.T) {
		err := Copy(testSourcePath, testTargetPath, 1000000, 0)
		require.NotNil(t, err)
		require.EqualError(t, err, ErrOffsetExceedsFileSize.Error())
	})

	// Удаляем временный файл.
	defer os.Remove(testTargetPath)
}
