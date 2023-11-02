package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	BrokenStructValidationTag struct {
		Code int    `validate:"in:200,404,500|max:S"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			User{
				ID:     "1",
				Name:   "Doesn't matter",
				Age:    55,
				Email:  "",
				Role:   "admin",
				Phones: []string{"12345678910"},
				meta:   json.RawMessage{},
			},
			ValidationErrors{
				ValidationError{"ID", errValidationIncorrectLen},
				ValidationError{"Age", errValidationIncorrectMax},
				ValidationError{"Email", errValidationNotMatchToRegexp},
			},
		},
		{
			User{
				ID:     "1dsadsadsadasdas23123123asdsadkbituj",
				Name:   "Doesn't matter",
				Age:    35,
				Email:  "asdas@sdsd.qq",
				Role:   "admin",
				Phones: []string{"12345678910"},
				meta:   json.RawMessage{},
			},
			ValidationErrors{},
		},
		{
			App{
				Version: "v1234",
			},
			ValidationErrors{},
		},
		{
			Token{
				Header:    []byte{11, 22},
				Payload:   []byte{22, 11},
				Signature: []byte{33, 44},
			},
			ValidationErrors{},
		},
		{
			Response{
				Code: 200,
				Body: "Test Body",
			},
			ValidationErrors{},
		},
		{
			BrokenStructValidationTag{
				Code: 200,
				Body: "Test Body",
			},
			errSystemError,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			var validationErr ValidationErrors
			ve := Validate(tt.in)

			if errors.As(ve, &validationErr) {
				require.EqualError(t, ve, tt.expectedErr.Error())
			} else if !errors.Is(ve, tt.expectedErr) {
				t.Errorf("Error: Expected: %v, but received: %v", tt.expectedErr, ve)
			}
		})
	}
}
