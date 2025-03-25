package webhook

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Request represents a webhook request
type Request struct {
	URL     string
	Method  string
	Headers map[string]string
	Params  any
	Timeout time.Duration
}

// Response represents a webhook response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
	Request    *Request // Reference to the original request
}

// IsSuccessful returns true if the response status code is in the 2xx range
func (r *Response) IsSuccessful() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// methodSupportsRequestBody returns true if the HTTP method typically supports a request body
func methodSupportsRequestBody(method string) bool {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}

// createHTTPRequest creates a new HTTP request with the given parameters
func createHTTPRequest(req *Request) (*http.Request, error) {
	if req.URL == "" {
		return nil, ErrInvalidURL
	}

	var httpReq *http.Request
	var err error

	// Create base URL
	baseURL := req.URL

	// Handle parameters based on HTTP method
	if methodSupportsRequestBody(req.Method) {
		// For methods that support a body (POST, PUT, PATCH, etc.)
		data, err := marshalParams(req.Params)
		if err != nil {
			return nil, err
		}
		httpReq, err = http.NewRequest(req.Method, baseURL, data)
	} else {
		// For methods that don't support a body (GET, HEAD, DELETE, etc.)
		// Convert params to query string and append to URL
		if req.Params != nil {
			parsedURL, err := url.Parse(baseURL)
			if err != nil {
				return nil, ErrInvalidURL
			}

			q := parsedURL.Query()
			if err := addParamsToQuery(q, req.Params); err != nil {
				return nil, err
			}

			parsedURL.RawQuery = q.Encode()
			baseURL = parsedURL.String()
		}
		
		httpReq, err = http.NewRequest(req.Method, baseURL, nil)
	}

	if err != nil {
		return nil, err
	}

	// Add headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

// addParamsToQuery adds parameters to the query string values
func addParamsToQuery(q url.Values, params any) error {
	if params == nil {
		return nil
	}

	v := reflect.ValueOf(params)
	
	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		// Handle map[string]any, map[string]string, etc.
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			val := iter.Value()
			
			// Convert value to string based on type
			strVal := valueToString(val)
			q.Add(key, strVal)
		}
		
	case reflect.Struct:
		// For structs, we would ideally use tags to determine field names
		// This is a simplified version that just uses field names
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if field.IsExported() {
				fieldName := field.Name
				// Check for json tag
				if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
					parts := strings.Split(tag, ",")
					if parts[0] != "" {
						fieldName = parts[0]
					}
				}
				fieldValue := v.Field(i)
				q.Add(fieldName, valueToString(fieldValue))
			}
		}
		
	default:
		// For primitive types or unsupported types, add as single value
		q.Add("value", valueToString(v))
	}

	return nil
}

// valueToString converts a reflect.Value to its string representation
func valueToString(v reflect.Value) string {
	// Handle nil values
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return ""
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return fmt.Sprintf("%v", v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", v.Float())
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			return string(v.Bytes())
		}
		// Fall through to default
	}

	// Default conversion
	return fmt.Sprintf("%v", v.Interface())
}
