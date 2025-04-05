package cqrs_test

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/dmitrymomot/gokit/cqrs"
)

func TestNewErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		reason   string
		kv       []string
		expected *cqrs.PublicError[url.Values]
	}{
		{
			name:   "Basic error message",
			err:    errors.New("basic error"),
			reason: "BASIC_ERROR",
			kv:     []string{},
			expected: &cqrs.PublicError[url.Values]{
				Code:    1000,
				Error:   "basic error",
				Reason:  "BASIC_ERROR",
				Details: url.Values{},
			},
		},
		{
			name:   "Error message with details",
			err:    errors.New("detailed error"),
			reason: "DETAILED_ERROR",
			kv:     []string{"field1", "value1", "field2", "value2"},
			expected: &cqrs.PublicError[url.Values]{
				Code:    1000,
				Error:   "detailed error",
				Reason:  "DETAILED_ERROR",
				Details: url.Values{"field1": []string{"value1"}, "field2": []string{"value2"}},
			},
		},
		{
			name:   "Error message with odd number of kv pairs",
			err:    errors.New("odd kv error"),
			reason: "ODD_KV_ERROR",
			kv:     []string{"field1", "value1", "field2"},
			expected: &cqrs.PublicError[url.Values]{
				Code:    1000,
				Error:   "odd kv error",
				Reason:  "ODD_KV_ERROR",
				Details: url.Values{"field1": []string{"value1"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cqrs.NewErrorMessage(tt.err, tt.reason, tt.kv...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NewErrorMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}
