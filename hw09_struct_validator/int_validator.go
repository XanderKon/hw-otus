package hw09structvalidator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	errValidationIncorrectMin       = errors.New("field has value less than min")
	errValidationIncorrectMax       = errors.New("field has value more than max")
	errValidationNotInPlentyNumbers = errors.New("field not in plenty of numbers")
)

type IntValidator struct{}

func (v IntValidator) Validate(rule string, ruleValue interface{}, checkValue interface{}) error {
	switch t := checkValue.(type) {
	case []int:
		for _, vl := range t {
			err := IntValidator.Validate(v, rule, ruleValue, vl)
			if err != nil {
				return err
			}
		}
		return nil
	default:
	}

	switch rule {
	case "min":
		num, err := strconv.Atoi(ruleValue.(string))
		if err != nil {
			return fmt.Errorf("incorrect tag value: %w", errSystemError)
		}

		if checkValue.(int) < num {
			return errValidationIncorrectMin
		}

	case "max":
		num, err := strconv.Atoi(ruleValue.(string))
		if err != nil {
			return fmt.Errorf("incorrect tag value: %w", errSystemError)
		}

		if checkValue.(int) > num {
			return errValidationIncorrectMax
		}

	case "in":
		targetSlice := strings.Split(ruleValue.(string), ",")
		for _, sliceVal := range targetSlice {
			num, err := strconv.Atoi(sliceVal)
			if err != nil {
				return fmt.Errorf("incorrect tag value: %w", errSystemError)
			}

			if num != checkValue.(int) {
				return errValidationNotInPlentyNumbers
			}

			return nil
		}
	}

	return nil
}
