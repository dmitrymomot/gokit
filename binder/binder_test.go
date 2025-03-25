package binder_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dmitrymomot/gokit/binder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name  string   `json:"name" form:"name"`
	Age   int      `json:"age" form:"age"`
	Email string   `json:"email" form:"email"`
	Tags  []string `json:"tags" form:"tags"`
}

func TestBind_JSON(t *testing.T) {
	data := testStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
		Tags:  []string{"tag1", "tag2"},
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result testStruct
	err = binder.Bind(req, &result)
	
	require.NoError(t, err)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Age, result.Age)
	assert.Equal(t, data.Email, result.Email)
	assert.Equal(t, data.Tags, result.Tags)
}

func TestBindJSON(t *testing.T) {
	data := testStruct{
		Name:  "Jane Doe",
		Age:   25,
		Email: "jane@example.com",
		Tags:  []string{"tag3", "tag4"},
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(jsonData))
	require.NoError(t, err)

	var result testStruct
	err = binder.BindJSON(req, &result)
	
	require.NoError(t, err)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Age, result.Age)
	assert.Equal(t, data.Email, result.Email)
	assert.Equal(t, data.Tags, result.Tags)
}

func TestBindQuery(t *testing.T) {
	u, err := url.Parse("http://example.com?name=Alice&age=35&email=alice@example.com&tags=tag5&tags=tag6")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)

	var result testStruct
	err = binder.BindQuery(req, &result)
	
	require.NoError(t, err)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, 35, result.Age)
	assert.Equal(t, "alice@example.com", result.Email)
	assert.Equal(t, []string{"tag5", "tag6"}, result.Tags)
}

func TestBindForm(t *testing.T) {
	formData := url.Values{}
	formData.Add("name", "Bob")
	formData.Add("age", "40")
	formData.Add("email", "bob@example.com")
	formData.Add("tags", "tag7")
	formData.Add("tags", "tag8")

	req, err := http.NewRequest(
		http.MethodPost, 
		"http://example.com", 
		strings.NewReader(formData.Encode()),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result testStruct
	err = binder.BindForm(req, &result)
	
	require.NoError(t, err)
	assert.Equal(t, "Bob", result.Name)
	assert.Equal(t, 40, result.Age)
	assert.Equal(t, "bob@example.com", result.Email)
	assert.Equal(t, []string{"tag7", "tag8"}, result.Tags)
}

func TestBind_QueryForGet(t *testing.T) {
	u, err := url.Parse("http://example.com?name=Charlie&age=45&email=charlie@example.com")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)
	// Don't set content type for GET request

	var result testStruct
	err = binder.Bind(req, &result)
	
	require.NoError(t, err)
	assert.Equal(t, "Charlie", result.Name)
	assert.Equal(t, 45, result.Age)
	assert.Equal(t, "charlie@example.com", result.Email)
}

func TestBind_ErrorEmptyBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result testStruct
	err = binder.Bind(req, &result)
	
	require.Error(t, err)
	assert.ErrorIs(t, err, binder.ErrEmptyBody)
}

func TestBind_ErrorInvalidJSON(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodPost, 
		"http://example.com", 
		bytes.NewBuffer([]byte(`{"name": "John", "age": "invalid"}`)),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result testStruct
	err = binder.Bind(req, &result)
	
	require.Error(t, err)
	assert.ErrorIs(t, err, binder.ErrInvalidJSON)
}

func TestBind_ErrorInvalidContentType(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	var result testStruct
	err = binder.Bind(req, &result)
	
	require.Error(t, err)
	assert.ErrorIs(t, err, binder.ErrInvalidContentType)
}

func TestBind_UnsupportedType(t *testing.T) {
	formData := url.Values{}
	formData.Add("name", "Test")

	req, err := http.NewRequest(
		http.MethodPost, 
		"http://example.com", 
		strings.NewReader(formData.Encode()),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result string // Not a struct pointer
	err = binder.BindForm(req, result)
	
	require.Error(t, err)
	assert.ErrorIs(t, err, binder.ErrUnsupportedType)
}
