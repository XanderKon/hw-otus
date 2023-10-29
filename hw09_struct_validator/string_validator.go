package hw09structvalidator

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	errValidationIncorrectLen       = errors.New("field has incorrect length")
	errValidationNotMatchToRegexp   = errors.New("field not match to Regexp")
	errValidationNotInPlentyStrings = errors.New("field not in plenty string")
)

type StringValidator struct{}

func (v StringValidator) Validate(rule string, ruleValue interface{}, checkValue interface{}) error {
	switch t := checkValue.(type) {
	case []string:
		for _, vl := range t {
			err := StringValidator.Validate(v, rule, ruleValue, vl)
			if err != nil {
				return err
			}
		}
		return nil
	default:
	}

	switch rule {
	case "len":
		num, err := strconv.Atoi(ruleValue.(string))
		if err != nil {
			return fmt.Errorf("incorrect tag value: %w", errSystemError)
		}

		if len(checkValue.(string)) != num {
			return errValidationIncorrectLen
		}

	case "regexp":
		match, err := regexp.MatchString(ruleValue.(string), checkValue.(string))
		if err != nil {
			return fmt.Errorf("incorrect Regexp value: %w", errSystemError)
		}

		if !match {
			return errValidationNotMatchToRegexp
		}

	case "in":
		targetSlice := strings.Split(ruleValue.(string), ",")
		for _, sliceVal := range targetSlice {
			if sliceVal == "" {
				return fmt.Errorf("incorrect tag value: %w", errSystemError)
			}
			if sliceVal != checkValue.(string) {
				return errValidationNotInPlentyStrings
			}
			return nil
		}
	}

	return nil
}
