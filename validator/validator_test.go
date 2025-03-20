package validator_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/dmitrymomot/gokit/validator"
)

func TestValidateStruct(t *testing.T) {
	// Custom error translator for testing
	translator := func(key string, label string, params ...string) string {
		return key + ":" + label
	}

	v := validator.NewValidator(translator)

	tests := []struct {
		name          string
		input         interface{}
		expectedError validator.ValidationErrors
	}{
		{
			name: "Valid basic struct",
			input: struct {
				Name  string `validate:"required" label:"Name"`
				Email string `validate:"email" label:"Email"`
			}{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectedError: validator.ValidationErrors(url.Values{}),
		},
		{
			name: "Invalid basic struct",
			input: struct {
				Name  string `validate:"required" label:"Name"`
				Email string `validate:"email" label:"Email"`
			}{
				Name:  "",
				Email: "invalid-email",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Name":  []string{"validation.required:Name"},
				"Email": []string{"validation.email:Email"},
			}),
		},
		{
			name: "Numeric validations",
			input: struct {
				Age    int `validate:"range:18,100" label:"Age"`
				Score  int `validate:"min:0|max:100" label:"Score"`
				Count  int `validate:"required|numeric" label:"Count"`
				Amount int `validate:"between:1,1000" label:"Amount"`
			}{
				Age:    15,
				Score:  150,
				Count:  0,
				Amount: 1500,
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Age":    []string{"validation.range:Age"},
				"Score":  []string{"validation.max:Score"},
				"Count":  []string{"validation.required:Count"},
				"Amount": []string{"validation.between:Amount"},
			}),
		},
		{
			name: "String validations",
			input: struct {
				Username string `validate:"required|username" label:"Username"`
				Password string `validate:"required|password" label:"Password"`
				Slug     string `validate:"required|slug" label:"Slug"`
				Hexcolor string `validate:"required|hexcolor" label:"Hexcolor"`
				Fullname string `validate:"required|fullname" label:"Fullname"`
				Phone    string `validate:"required|phone" label:"Phone"`
			}{
				Username: "user@name",
				Password: "weak",
				Slug:     "Invalid Slug",
				Hexcolor: "invalid",
				Fullname: "John123",
				Phone:    "123",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Username": []string{"validation.username:Username"},
				"Password": []string{"validation.password:Password"},
				"Slug":     []string{"validation.slug:Slug"},
				"Hexcolor": []string{"validation.hexcolor:Hexcolor"},
				"Fullname": []string{"validation.fullname:Fullname"},
				"Phone":    []string{"validation.phone:Phone"},
			}),
		},
		{
			name: "Custom validation",
			input: struct {
				CustomField string `validate:"custom" label:"Custom"`
			}{
				CustomField: "test",
			},
			expectedError: validator.ValidationErrors(url.Values{}),
		},
		{
			name: "Nested struct",
			input: struct {
				User struct {
					Name  string `validate:"required" label:"Name"`
					Email string `validate:"email" label:"Email"`
				} `validate:"required"`
			}{
				User: struct {
					Name  string `validate:"required" label:"Name"`
					Email string `validate:"email" label:"Email"`
				}{
					Name:  "",
					Email: "invalid-email",
				},
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Name":  []string{"validation.required:Name"},
				"Email": []string{"validation.email:Email"},
			}),
		},
		{
			name:          "Non-struct input",
			input:         "not a struct",
			expectedError: validator.ValidationErrors(url.Values{}),
		},
		{
			name: "Multiple validations per field",
			input: struct {
				Password string `validate:"required|min:8|password" label:"Password"`
			}{
				Password: "weak",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Password": []string{"validation.min:Password", "validation.password:Password"},
			}),
		},
		{
			name: "URL validation",
			input: struct {
				Website string `validate:"url" label:"Website"`
			}{
				Website: "not-a-url",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Website": []string{"validation.url:Website"},
			}),
		},
		{
			name: "IP address validation",
			input: struct {
				IP string `validate:"ip" label:"IP"`
			}{
				IP: "not-an-ip",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"IP": []string{"validation.ip:IP"},
			}),
		},
		{
			name: "Date validation",
			input: struct {
				Date string `validate:"date:2006-01-02" label:"Date"`
			}{
				Date: "invalid-date",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Date": []string{"validation.date:Date"},
			}),
		},
		{
			name: "Enum validation",
			input: struct {
				Status string `validate:"in:active,inactive" label:"Status"`
			}{
				Status: "invalid",
			},
			expectedError: validator.ValidationErrors(url.Values{
				"Status": []string{"validation.in:Status"},
			}),
		},
	}

	// Register custom validation for testing
	validator.RegisterValidation("custom", func(fieldValue interface{}, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		return nil
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.input)
			if err == nil && len(tt.expectedError) > 0 {
				t.Errorf("ValidateStruct() expected error %v, got nil", tt.expectedError)
				return
			}
			if err != nil {
				validationErr, ok := err.(validator.ValidationErrors)
				if !ok {
					t.Errorf("ValidateStruct() returned wrong error type")
					return
				}
				if !reflect.DeepEqual(validationErr, tt.expectedError) {
					t.Errorf("ValidateStruct() = %v, want %v", validationErr, tt.expectedError)
				}
			}
		})
	}
}
