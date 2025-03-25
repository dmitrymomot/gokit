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
	"mime/multipart"
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

func TestBindMultipartForm(t *testing.T) {
	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add form fields
	require.NoError(t, writer.WriteField("name", "David"))
	require.NoError(t, writer.WriteField("age", "50"))
	require.NoError(t, writer.WriteField("email", "david@example.com"))
	require.NoError(t, writer.WriteField("tags", "tag9"))
	require.NoError(t, writer.WriteField("tags", "tag10"))
	
	// Close the writer to set the terminating boundary
	require.NoError(t, writer.Close())
	
	// Create the request with the multipart form
	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com",
		body,
	)
	require.NoError(t, err)
	
	// Set the content type with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Bind the form data
	var result testStruct
	err = binder.BindForm(req, &result)
	
	// Verify the results
	require.NoError(t, err)
	assert.Equal(t, "David", result.Name)
	assert.Equal(t, 50, result.Age)
	assert.Equal(t, "david@example.com", result.Email)
	assert.Equal(t, []string{"tag9", "tag10"}, result.Tags)
}

func TestBind_MultipartForm(t *testing.T) {
	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add form fields
	require.NoError(t, writer.WriteField("name", "Emma"))
	require.NoError(t, writer.WriteField("age", "55"))
	require.NoError(t, writer.WriteField("email", "emma@example.com"))
	require.NoError(t, writer.WriteField("tags", "tag11"))
	require.NoError(t, writer.WriteField("tags", "tag12"))
	
	// Close the writer to set the terminating boundary
	require.NoError(t, writer.Close())
	
	// Create the request with the multipart form
	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com",
		body,
	)
	require.NoError(t, err)
	
	// Set the content type with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Bind the form data using the general Bind function
	var result testStruct
	err = binder.Bind(req, &result)
	
	// Verify the results
	require.NoError(t, err)
	assert.Equal(t, "Emma", result.Name)
	assert.Equal(t, 55, result.Age)
	assert.Equal(t, "emma@example.com", result.Email)
	assert.Equal(t, []string{"tag11", "tag12"}, result.Tags)
}
