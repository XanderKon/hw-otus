package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var errSystemError = errors.New("system error")

// Интерфейс для валидатора.
type Validator interface {
	Validate(string, interface{}, interface{}) error
}

// Дефолтный валидатор. Не проводит никакой валидации.
type DefaultValidator struct{}

func (v DefaultValidator) Validate(string, interface{}, interface{}) error {
	return nil
}

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for _, err := range v {
		sb.WriteString(fmt.Sprintf("%s: %s\n", err.Field, err.Err))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	fields := reflect.ValueOf(v)
	types := reflect.TypeOf(v)
	var valErrors ValidationErrors

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		_type := types.Field(i)

		// Получаем валидатор для поля.
		validator := getValidator(field.Type().String())

		// Основной тег для валидации
		tag := _type.Tag.Get("validate")

		if tag == "" {
			continue
		}

		// Список тегов в виде слайса.
		tags := strings.Split(tag, "|")

		// Валидируем.
		for _, tag := range tags {
			tagValue := strings.Split(tag, ":")

			if len(tagValue) < 2 {
				return fmt.Errorf("incorrect tag value: %w", errSystemError)
			}

			err := validator.Validate(tagValue[0], tagValue[1], fields.Field(i).Interface())
			if err != nil {
				if errors.Is(err, errSystemError) {
					return err
				}
				valErrors = append(valErrors, ValidationError{
					Field: _type.Name,
					Err:   err,
				})
			}
		}
	}

	return valErrors
}

func getValidator(field string) Validator {
	switch field {
	case "string", "[]string":
		validator := StringValidator{}
		return validator
	case "int", "[]int":
		validator := IntValidator{}
		return validator
	}
	return DefaultValidator{}
}
