package binder_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"mime/multipart"

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

// Custom type for testing TextUnmarshaler
type CustomID string

func (c *CustomID) UnmarshalText(text []byte) error {
	*c = CustomID("ID-" + string(text))
	return nil
}

// Nested struct for testing nested binding
type Address struct {
	Street  string `json:"street" form:"street"`
	City    string `json:"city" form:"city"`
	ZipCode string `json:"zip_code" form:"zip_code"`
}

type UserWithAddress struct {
	Name    string  `json:"name" form:"name"`
	Email   string  `json:"email" form:"email"`
	Address Address `json:"address" form:"address"`
}

// Struct with time.Time fields for testing time binding
type Event struct {
	Title     string    `json:"title" form:"title"`
	StartDate time.Time `json:"start_date" form:"start_date"`
	EndDate   time.Time `json:"end_date" form:"end_date"`
}

// Struct with map field for testing map binding
type Settings struct {
	Name       string            `json:"name" form:"name"`
	Properties map[string]string `json:"properties" form:"properties"`
}

// Struct with custom type for testing TextUnmarshaler
type Resource struct {
	Name string   `json:"name" form:"name"`
	ID   CustomID `json:"id" form:"id"`
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

func TestBind_NestedStruct_JSON(t *testing.T) {
	data := UserWithAddress{
		Name:  "John Doe",
		Email: "john@example.com",
		Address: Address{
			Street:  "123 Main St",
			City:    "San Francisco",
			ZipCode: "94105",
		},
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result UserWithAddress
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Email, result.Email)
	assert.Equal(t, data.Address.Street, result.Address.Street)
	assert.Equal(t, data.Address.City, result.Address.City)
	assert.Equal(t, data.Address.ZipCode, result.Address.ZipCode)
}

func TestBind_NestedStruct_Form(t *testing.T) {
	formData := url.Values{}
	formData.Add("name", "Jane Doe")
	formData.Add("email", "jane@example.com")
	formData.Add("address.street", "456 Market St")
	formData.Add("address.city", "New York")
	formData.Add("address.zip_code", "10001")

	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com",
		strings.NewReader(formData.Encode()),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result UserWithAddress
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, "Jane Doe", result.Name)
	assert.Equal(t, "jane@example.com", result.Email)
	assert.Equal(t, "456 Market St", result.Address.Street)
	assert.Equal(t, "New York", result.Address.City)
	assert.Equal(t, "10001", result.Address.ZipCode)
}

func TestBind_NestedStruct_Query(t *testing.T) {
	u, err := url.Parse("http://example.com?name=Alice&email=alice@example.com&address.street=789+Broadway&address.city=Chicago&address.zip_code=60601")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)

	var result UserWithAddress
	err = binder.BindQuery(req, &result)

	require.NoError(t, err)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, "alice@example.com", result.Email)
	assert.Equal(t, "789 Broadway", result.Address.Street)
	assert.Equal(t, "Chicago", result.Address.City)
	assert.Equal(t, "60601", result.Address.ZipCode)
}

func TestBind_TimeField_JSON(t *testing.T) {
	startDate, err := time.Parse(time.RFC3339, "2025-05-15T09:00:00Z")
	require.NoError(t, err)

	endDate, err := time.Parse(time.RFC3339, "2025-05-17T18:00:00Z")
	require.NoError(t, err)

	data := Event{
		Title:     "Conference",
		StartDate: startDate,
		EndDate:   endDate,
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result Event
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, data.Title, result.Title)
	assert.Equal(t, data.StartDate.Format(time.RFC3339), result.StartDate.Format(time.RFC3339))
	assert.Equal(t, data.EndDate.Format(time.RFC3339), result.EndDate.Format(time.RFC3339))
}

func TestBind_TimeField_Form(t *testing.T) {
	formData := url.Values{}
	formData.Add("title", "Workshop")
	formData.Add("start_date", "2025-06-10T14:00:00Z")
	formData.Add("end_date", "2025-06-10T17:00:00Z")

	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com",
		strings.NewReader(formData.Encode()),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result Event
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, "Workshop", result.Title)

	expectedStart, err := time.Parse(time.RFC3339, "2025-06-10T14:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, expectedStart.Format(time.RFC3339), result.StartDate.Format(time.RFC3339))

	expectedEnd, err := time.Parse(time.RFC3339, "2025-06-10T17:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, expectedEnd.Format(time.RFC3339), result.EndDate.Format(time.RFC3339))
}

func TestBind_TimeField_SimpleFormat(t *testing.T) {
	u, err := url.Parse("http://example.com?title=Seminar&start_date=2025-07-01&end_date=2025-07-03")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)

	var result Event
	err = binder.BindQuery(req, &result)

	require.NoError(t, err)
	assert.Equal(t, "Seminar", result.Title)

	// Check year, month, day components
	assert.Equal(t, 2025, result.StartDate.Year())
	assert.Equal(t, time.July, result.StartDate.Month())
	assert.Equal(t, 1, result.StartDate.Day())

	assert.Equal(t, 2025, result.EndDate.Year())
	assert.Equal(t, time.July, result.EndDate.Month())
	assert.Equal(t, 3, result.EndDate.Day())
}

func TestBind_MapField_JSON(t *testing.T) {
	data := Settings{
		Name: "App Config",
		Properties: map[string]string{
			"theme":      "dark",
			"font_size":  "16px",
			"debug_mode": "true",
		},
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result Settings
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Properties["theme"], result.Properties["theme"])
	assert.Equal(t, data.Properties["font_size"], result.Properties["font_size"])
	assert.Equal(t, data.Properties["debug_mode"], result.Properties["debug_mode"])
}

func TestBind_MapField_Form(t *testing.T) {
	formData := url.Values{}
	formData.Add("name", "User Preferences")
	formData.Add("properties", "theme=light,font_size=14px,language=en")

	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com",
		strings.NewReader(formData.Encode()),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result Settings
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, "User Preferences", result.Name)
	assert.Equal(t, "light", result.Properties["theme"])
	assert.Equal(t, "14px", result.Properties["font_size"])
	assert.Equal(t, "en", result.Properties["language"])
}

func TestBind_CustomType_JSON(t *testing.T) {
	data := Resource{
		Name: "Document",
		ID:   "12345",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var result Resource
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, data.Name, result.Name)
	// The UnmarshalText is used for JSON as well in our implementation
	assert.Equal(t, CustomID("ID-12345"), result.ID)
}

func TestBind_CustomType_Form(t *testing.T) {
	formData := url.Values{}
	formData.Add("name", "File")
	formData.Add("id", "67890")

	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com",
		strings.NewReader(formData.Encode()),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result Resource
	err = binder.Bind(req, &result)

	require.NoError(t, err)
	assert.Equal(t, "File", result.Name)
	// The current binder implementation doesn't call UnmarshalText for form values
	assert.Equal(t, CustomID("67890"), result.ID)
}
