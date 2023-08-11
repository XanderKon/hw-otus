package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	var res strings.Builder

	runeString := []rune(s)

	for i, symbol := range runeString {
		isDigit := unicode.IsDigit(symbol)

		isEnd, nextVal, nextErr := getNextVal(runeString, i)

		// skip string starts with digit & numbers
		if isDigit && (i == 0 || (nextErr == nil && !isEnd)) {
			return "", ErrInvalidString
		}

		// if next is digit
		if nextErr == nil && !isEnd {
			res.WriteString(strings.Repeat(string(symbol), nextVal))
		} else if !isDigit {
			res.WriteRune(symbol)
		}
	}

	return res.String(), nil
}

func getNextVal(r []rune, i int) (bool, int, error) {
	if i == len(r)-1 {
		return true, 0, nil
	}

	nextVal, nextErr := strconv.Atoi(string(r[i+1]))

	return false, nextVal, nextErr
}
