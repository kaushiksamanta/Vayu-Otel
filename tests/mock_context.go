package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
)

// MockContext is a simple mock for vayu.Context
type MockContext struct {
	HTTPRequest *http.Request
	Values      map[string]interface{}
	Params      map[string]string
	StatusCode  int
	Writer      *MockResponseWriter
	stringMap   map[string]string
	intMap      map[string]int
	floatMap    map[string]float64
	boolMap     map[string]bool
	sliceMap    map[string][]string
	jsonMap     map[string]map[string]interface{}
}

// MockResponseWriter is a simple mock for http.ResponseWriter
type MockResponseWriter struct {
	Headers    http.Header
	StatusCode int
	Body       bytes.Buffer
}

// Header returns the header map to set HTTP response headers
func (w *MockResponseWriter) Header() http.Header {
	if w.Headers == nil {
		w.Headers = make(http.Header)
	}
	return w.Headers
}

// Write writes the data to the response body buffer
func (w *MockResponseWriter) Write(data []byte) (int, error) {
	return w.Body.Write(data)
}

// WriteHeader sets the response status code
func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

// Status returns the response status code
func (w *MockResponseWriter) Status() int {
	return w.StatusCode
}

// NewMockContext creates a new MockContext for testing
func NewMockContext(method, path string) *MockContext {
	req := httptest.NewRequest(method, path, nil)
	mockWriter := &MockResponseWriter{
		Headers:    make(http.Header),
		StatusCode: 200,
	}
	return &MockContext{
		HTTPRequest: req,
		Writer:      mockWriter,
		StatusCode:  200,
		Params:      make(map[string]string),
		Values:      make(map[string]interface{}),
		stringMap:   make(map[string]string),
		intMap:      make(map[string]int),
		floatMap:    make(map[string]float64),
		boolMap:     make(map[string]bool),
		sliceMap:    make(map[string][]string),
		jsonMap:     make(map[string]map[string]interface{}),
	}
}

// SetValue sets a value in the context store
func (c *MockContext) SetValue(key string, value interface{}) {
	c.Values[key] = value
}

// GetValue retrieves a value from the context store
func (c *MockContext) GetValue(key string) (interface{}, bool) {
	value, exists := c.Values[key]
	return value, exists
}

// Get retrieves a value from the context store (Vayu API compatibility)
func (c *MockContext) Get(key string) (interface{}, bool) {
	return c.GetValue(key)
}

// Set sets a value in the context store (Vayu API compatibility)
func (c *MockContext) Set(key string, value interface{}) {
	c.SetValue(key, value)
}

// SetString sets a string value in the context
func (c *MockContext) SetString(key, value string) {
	c.stringMap[key] = value
}

// GetString gets a string value from the context
func (c *MockContext) GetString(key string) string {
	return c.stringMap[key]
}

// SetInt sets an int value in the context
func (c *MockContext) SetInt(key string, value int) {
	c.intMap[key] = value
}

// GetInt gets an int value from the context
func (c *MockContext) GetInt(key string) int {
	return c.intMap[key]
}

// SetFloat sets a float64 value in the context
func (c *MockContext) SetFloat(key string, value float64) {
	c.floatMap[key] = value
}

// GetFloat gets a float64 value from the context
func (c *MockContext) GetFloat(key string) float64 {
	return c.floatMap[key]
}

// SetBool sets a bool value in the context
func (c *MockContext) SetBool(key string, value bool) {
	c.boolMap[key] = value
}

// GetBool gets a bool value from the context
func (c *MockContext) GetBool(key string) bool {
	return c.boolMap[key]
}

// SetStringSlice sets a string slice in the context
func (c *MockContext) SetStringSlice(key string, value []string) {
	c.sliceMap[key] = value
}

// GetStringSlice gets a string slice from the context
func (c *MockContext) GetStringSlice(key string) []string {
	return c.sliceMap[key]
}

// JSONMap renders JSON data with the given status code
func (c *MockContext) JSONMap(statusCode int, data map[string]interface{}) {
	c.StatusCode = statusCode
	// In a real implementation, this would render JSON
}

// Param gets a path parameter (mock implementation always returns empty string)
func (c *MockContext) Param(name string) string {
	return ""
}

// Query gets a query parameter (mock implementation always returns empty string)
func (c *MockContext) Query(name string) string {
	return ""
}

// Request returns the HTTP request for this context
func (c *MockContext) Request() *http.Request {
	return c.HTTPRequest
}
