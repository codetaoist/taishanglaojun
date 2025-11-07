package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// APITestSuite provides a base suite for API integration tests
type APITestSuite struct {
	suite.Suite
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// SetupSuite runs once before all tests in the suite
func (s *APITestSuite) SetupSuite() {
	// Initialize HTTP client with timeout
	s.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Set base URL from environment or use default
	s.baseURL = "http://localhost:8080" // API Gateway

	// Wait for the API to be ready
	s.waitForAPI()
}

// TearDownSuite runs once after all tests in the suite
func (s *APITestSuite) TearDownSuite() {
	// Cleanup resources if needed
}

// waitForAPI waits for the API to be ready
func (s *APITestSuite) waitForAPI() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := s.httpClient.Get(s.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	s.T().Fatal("API is not ready after maximum attempts")
}

// makeRequest makes an HTTP request with the given method, path, and body
func (s *APITestSuite) makeRequest(method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, s.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if s.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.authToken)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return s.httpClient.Do(req)
}

// get makes a GET request
func (s *APITestSuite) get(path string, headers map[string]string) (*http.Response, error) {
	return s.makeRequest(http.MethodGet, path, nil, headers)
}

// post makes a POST request
func (s *APITestSuite) post(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return s.makeRequest(http.MethodPost, path, body, headers)
}

// put makes a PUT request
func (s *APITestSuite) put(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return s.makeRequest(http.MethodPut, path, body, headers)
}

// delete makes a DELETE request
func (s *APITestSuite) delete(path string, headers map[string]string) (*http.Response, error) {
	return s.makeRequest(http.MethodDelete, path, nil, headers)
}

// assertStatus asserts that the response has the expected status code
func (s *APITestSuite) assertStatus(resp *http.Response, expectedStatus int) {
	assert.Equal(s.T(), expectedStatus, resp.StatusCode, 
		fmt.Sprintf("Expected status %d, got %d. Response body: %s", 
			expectedStatus, resp.StatusCode, s.getResponseBody(resp)))
}

// assertSuccess asserts that the response has a success status code (2xx)
func (s *APITestSuite) assertSuccess(resp *http.Response) {
	assert.True(s.T(), resp.StatusCode >= 200 && resp.StatusCode < 300, 
		fmt.Sprintf("Expected success status, got %d. Response body: %s", 
			resp.StatusCode, s.getResponseBody(resp)))
}

// getResponseBody returns the response body as a string
func (s *APITestSuite) getResponseBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Failed to read response body: %v", err)
	}
	return string(body)
}

// parseJSONResponse parses the response body as JSON into the provided interface
func (s *APITestSuite) parseJSONResponse(resp *http.Response, v interface{}) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), json.Unmarshal(body, v))
}

// login performs a login request and stores the auth token
func (s *APITestSuite) login(username, password string) {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := s.post("/auth/login", loginData, nil)
	require.NoError(s.T(), err)
	s.assertStatus(resp, http.StatusOK)

	var loginResp struct {
		Token string `json:"token"`
	}
	s.parseJSONResponse(resp, &loginResp)
	require.NotEmpty(s.T(), loginResp.Token)

	s.authToken = loginResp.Token
}

// logout clears the stored auth token
func (s *APITestSuite) logout() {
	s.authToken = ""
}

// TestHealthCheck tests the health check endpoint
func (s *APITestSuite) TestHealthCheck() {
	resp, err := s.get("/health", nil)
	require.NoError(s.T(), err)
	s.assertStatus(resp, http.StatusOK)

	var healthResp map[string]interface{}
	s.parseJSONResponse(resp, &healthResp)
	assert.Equal(s.T(), "ok", healthResp["status"])
}

// TestMetrics tests the metrics endpoint
func (s *APITestSuite) TestMetrics() {
	resp, err := s.get("/metrics", nil)
	require.NoError(s.T(), err)
	s.assertStatus(resp, http.StatusOK)

	// Metrics endpoint should return text content
	assert.Equal(s.T(), "text/plain", resp.Header.Get("Content-Type"))
}

// TestAPIGatewayRoutes tests that the API Gateway routes are working
func (s *APITestSuite) TestAPIGatewayRoutes() {
	// Test API route
	resp, err := s.get("/api/taishang/domains", nil)
	if err == nil {
		// If the request succeeds, it should return either 200 (success) or 401 (unauthorized)
		assert.True(s.T(), resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized)
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	// Test Auth route
	resp, err = s.get("/auth/health", nil)
	if err == nil {
		s.assertStatus(resp, http.StatusOK)
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	// Test Notification route
	resp, err = s.get("/notification/health", nil)
	if err == nil {
		// If the request succeeds, it should return either 200 (success) or 401 (unauthorized)
		assert.True(s.T(), resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized)
		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

// RunIntegrationTests runs all integration tests
func RunIntegrationTests(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}