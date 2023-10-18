package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")

	ErrWriteFileError             = errors.New("there is some error with writing file")
	ErrCannotCreateWriteFileError = errors.New("there is error with creating error file")

	ErrFromAndToAreEqual = errors.New("source and target cannot be the same")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	err := checkSafeTargetAndSource(fromPath, toPath)
	if err != nil {
		return err
	}

	fileRead, err := os.OpenFile(fromPath, os.O_RDONLY, 0o644)
	if err != nil {
		return ErrUnsupportedFile
	}

	defer fileRead.Close()

	// Получаем размер файла.
	fi, _ := fileRead.Stat()
	fileSize := fi.Size()

	// Проверяем на корректное указание offset.
	if offset != 0 && offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	// Если лимит не указан, пишем все.
	if limit == 0 {
		limit = fileSize
	}

	// Проверка на превышение лимита, если указан отступ.
	if fileSize-offset < limit {
		limit = fileSize - offset
	}

	return writer(fileRead, toPath, limit, offset)
}

func writer(src *os.File, toPath string, limit int64, offset int64) error {
	file, err := os.Create(toPath)
	if err != nil {
		return ErrCannotCreateWriteFileError
	}

	// Инициализируем прогресс-бар.
	bar := pb.Full.Start64(limit)

	// Создаем ридер из слайса байт с указанными limit и offset
	buf := make([]byte, limit)
	src.ReadAt(buf, offset)
	reader := bytes.NewReader(buf)

	// Создаем прокси ридер, чтобы выводить прогресс-бар.
	barReader := bar.NewProxyReader(reader)

	if _, err := io.Copy(file, barReader); err != nil {
		return ErrWriteFileError
	}

	// Завершаем прогресс-бар.
	defer bar.Finish()
	// Закрываем файл.
	defer file.Close()

	return nil
}

// Дополнительные проверки.
func checkSafeTargetAndSource(fromPath, toPath string) error {
	pathFrom, _ := filepath.Abs(fromPath)
	pathTo, _ := filepath.Abs(toPath)

	if pathFrom == pathTo {
		return ErrFromAndToAreEqual
	}

	if fromPath == "/dev/random" || fromPath == "/dev/urandom" {
		return ErrUnsupportedFile
	}

	return nil
}
