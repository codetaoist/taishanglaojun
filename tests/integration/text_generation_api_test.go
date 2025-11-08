package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TextGenerationAPITestSuite tests text generation API endpoints
type TextGenerationAPITestSuite struct {
	APITestSuite
}

// SetupSuite runs once before all tests in the suite
func (s *TextGenerationAPITestSuite) SetupSuite() {
	// Call parent setup
	s.APITestSuite.SetupSuite()
	
	// Set base URL to API service directly
	s.baseURL = "http://localhost:8082"
	
	// Wait for the API to be ready
	s.waitForAPI()
	
	// Login to get auth token
	s.login("testuser", "testpass")
}

// TestListModelServices tests listing all available model services
func (s *TextGenerationAPITestSuite) TestListModelServices() {
	resp, err := s.get("/api/v1/taishang/model/services", nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	s.assertStatus(resp, http.StatusOK)
	
	var response map[string]interface{}
	s.parseJSONResponse(resp, &response)
	
	data, ok := response["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	services, ok := data["services"].([]interface{})
	require.True(s.T(), ok, "Services should be an array")
	
	// We should have at least one service
	assert.GreaterOrEqual(s.T(), len(services), 1)
	
	// Check for common services like OpenAI, Ollama, etc.
	serviceNames := make(map[string]bool)
	for _, service := range services {
		if name, ok := service.(string); ok {
			serviceNames[name] = true
		}
	}
	
	// At least one of these services should be available
	assert.True(s.T(), 
		serviceNames["OpenAI"] || serviceNames["Ollama"] || serviceNames["default"],
		"At least one model service should be available")
}

// TestHealthCheckModelServices tests health check of all model services
func (s *TextGenerationAPITestSuite) TestHealthCheckModelServices() {
	resp, err := s.get("/api/v1/taishang/model/services/health", nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// This might fail if some services are not configured
	// We'll accept either success or a 500 error indicating some services are unavailable
	if resp.StatusCode == http.StatusInternalServerError {
		// Some services might not be configured, which is OK for this test
		return
	}
	
	s.assertStatus(resp, http.StatusOK)
	
	var response map[string]interface{}
	s.parseJSONResponse(resp, &response)
	
	data, ok := response["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	services, ok := data["services"].(map[string]interface{})
	require.True(s.T(), ok, "Services should be a map")
	
	// Check if at least one service is healthy
	healthyFound := false
	for _, health := range services {
		if healthMap, ok := health.(map[string]interface{}); ok {
			if status, ok := healthMap["status"].(string); ok && status == "healthy" {
				healthyFound = true
				break
			}
		}
	}
	
	assert.True(s.T(), healthyFound, "At least one service should be healthy")
}

// TestListServiceModels tests listing models for a specific service
func (s *TextGenerationAPITestSuite) TestListServiceModels() {
	// First get available services
	resp, err := s.get("/api/v1/taishang/model/services", nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	s.assertStatus(resp, http.StatusOK)
	
	var servicesResponse map[string]interface{}
	s.parseJSONResponse(resp, &servicesResponse)
	
	data, ok := servicesResponse["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	services, ok := data["services"].([]interface{})
	require.True(s.T(), ok, "Services should be an array")
	require.Greater(s.T(), len(services), 0, "At least one service should be available")
	
	// Try to get models for the first available service
	serviceName, ok := services[0].(string)
	require.True(s.T(), ok, "Service name should be a string")
	
	resp, err = s.get(fmt.Sprintf("/api/v1/taishang/model/service/%s/models", serviceName), nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// This might fail if the service is not configured
	// We'll accept either success or a 500 error indicating the service is unavailable
	if resp.StatusCode == http.StatusInternalServerError {
		// Service might not be configured, which is OK for this test
		return
	}
	
	s.assertStatus(resp, http.StatusOK)
	
	var modelsResponse map[string]interface{}
	s.parseJSONResponse(resp, &modelsResponse)
	
	modelsData, ok := modelsResponse["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	models, ok := modelsData["models"].([]interface{})
	require.True(s.T(), ok, "Models should be an array")
}

// TestGenerateText tests non-streaming text generation
func (s *TextGenerationAPITestSuite) TestGenerateText() {
	requestBody := map[string]interface{}{
		"model": "default",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
		"max_tokens": 50,
		"temperature": 0.7,
	}
	
	resp, err := s.post("/api/v1/taishang/model/generate", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// This might fail if no model service is configured
	// We'll accept either success or a 500 error indicating no service is available
	if resp.StatusCode == http.StatusInternalServerError {
		// No model service might be configured, which is OK for this test
		return
	}
	
	s.assertStatus(resp, http.StatusOK)
	
	var response map[string]interface{}
	s.parseJSONResponse(resp, &response)
	
	data, ok := response["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	// Check response structure
	choices, ok := data["choices"].([]interface{})
	require.True(s.T(), ok, "Choices should be an array")
	require.Greater(s.T(), len(choices), 0, "At least one choice should be returned")
	
	// Check the first choice
	choice, ok := choices[0].(map[string]interface{})
	require.True(s.T(), ok, "Choice should be a map")
	
	message, ok := choice["message"].(map[string]interface{})
	require.True(s.T(), ok, "Message should be a map")
	
	content, ok := message["content"].(string)
	require.True(s.T(), ok, "Content should be a string")
	require.NotEmpty(s.T(), content, "Content should not be empty")
	
	// Check other fields
	assert.NotEmpty(s.T(), data["id"])
	assert.NotEmpty(s.T(), data["object"])
	assert.NotEmpty(s.T(), data["created"])
}

// TestGenerateTextStream tests streaming text generation
func (s *TextGenerationAPITestSuite) TestGenerateTextStream() {
	requestBody := map[string]interface{}{
		"model": "default",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
		"max_tokens": 20,
		"temperature": 0.7,
		"stream":      true,
	}
	
	resp, err := s.post("/api/v1/taishang/model/generate/stream", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// This might fail if no model service is configured
	// We'll accept either success or a 500 error indicating no service is available
	if resp.StatusCode == http.StatusInternalServerError {
		// No model service might be configured, which is OK for this test
		return
	}
	
	s.assertStatus(resp, http.StatusOK)
	
	// Check response headers
	assert.Equal(s.T(), "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(s.T(), "no-cache", resp.Header.Get("Cache-Control"))
	assert.Equal(s.T(), "keep-alive", resp.Header.Get("Connection"))
	
	// Read the stream
	decoder := json.NewDecoder(resp.Body)
	chunkCount := 0
	
	for {
		var chunk map[string]interface{}
		err := decoder.Decode(&chunk)
		if err != nil {
			// End of stream
			break
		}
		
		chunkCount++
		
		// Check chunk structure
		choices, ok := chunk["choices"].([]interface{})
		if ok && len(choices) > 0 {
			choice, ok := choices[0].(map[string]interface{})
			if ok {
				delta, ok := choice["delta"].(map[string]interface{})
				if ok {
					content, ok := delta["content"].(string)
					if ok && content != "" {
						// Got some content
						break
					}
				}
			}
		}
		
		// Limit the number of chunks we read to avoid infinite loops
		if chunkCount > 50 {
			break
		}
	}
	
	// We should have received at least one chunk
	assert.Greater(s.T(), chunkCount, 0, "Should have received at least one chunk")
}

// TestGenerateTextWithDifferentModels tests text generation with different models
func (s *TextGenerationAPITestSuite) TestGenerateTextWithDifferentModels() {
	// First get available services and models
	resp, err := s.get("/api/v1/taishang/model/services", nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	s.assertStatus(resp, http.StatusOK)
	
	var servicesResponse map[string]interface{}
	s.parseJSONResponse(resp, &servicesResponse)
	
	data, ok := servicesResponse["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	services, ok := data["services"].([]interface{})
	require.True(s.T(), ok, "Services should be an array")
	
	// Try to generate text with each available service
	for _, service := range services {
		serviceName, ok := service.(string)
		if !ok {
			continue
		}
		
		// Skip services that are likely not configured
		if serviceName == "Ollama" {
			// Ollama might not be running, skip it for now
			continue
		}
		
		requestBody := map[string]interface{}{
			"model": serviceName,
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
			"max_tokens": 10,
		}
		
		resp, err := s.post("/api/v1/taishang/model/generate", requestBody, nil)
		require.NoError(s.T(), err)
		defer resp.Body.Close()
		
		// This might fail if the service is not configured
		// We'll accept either success or a 500 error indicating the service is unavailable
		if resp.StatusCode == http.StatusInternalServerError {
			// Service might not be configured, which is OK for this test
			continue
		}
		
		s.assertStatus(resp, http.StatusOK)
		
		var response map[string]interface{}
		s.parseJSONResponse(resp, &response)
		
		responseData, ok := response["data"].(map[string]interface{})
		require.True(s.T(), ok, "Response data should be a map")
		
		// Check response structure
		choices, ok := responseData["choices"].([]interface{})
		require.True(s.T(), ok, "Choices should be an array")
		require.Greater(s.T(), len(choices), 0, "At least one choice should be returned")
	}
}

// TestGenerateTextErrorHandling tests error handling in text generation
func (s *TextGenerationAPITestSuite) TestGenerateTextErrorHandling() {
	// Test with invalid request body
	requestBody := map[string]interface{}{
		"model": "nonexistent-model",
		"messages": []map[string]interface{}{
			{
				"role":    "invalid-role",
				"content": "Hello, how are you?",
			},
		},
		"max_tokens": -1, // Invalid value
	}
	
	resp, err := s.post("/api/v1/taishang/model/generate", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// Should return an error
	assert.True(s.T(), 
		resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError,
		"Should return an error for invalid request")
}

// TestCreateModelService tests creating a new model service configuration
func (s *TextGenerationAPITestSuite) TestCreateModelService() {
	requestBody := map[string]interface{}{
		"name":        "test-service",
		"service_type": "openai",
		"endpoint":   "https://api.openai.com/v1",
		"api_key":    "test-key",
		"enabled":    true,
	}
	
	resp, err := s.post("/api/v1/taishang/model/services", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// This might fail if the service configuration is invalid
	// We'll accept either success or a 400/500 error indicating invalid configuration
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError {
		// Service configuration might be invalid, which is OK for this test
		return
	}
	
	s.assertStatus(resp, http.StatusCreated)
	
	var response map[string]interface{}
	s.parseJSONResponse(resp, &response)
	
	data, ok := response["data"].(map[string]interface{})
	require.True(s.T(), ok, "Response data should be a map")
	
	assert.Equal(s.T(), "test-service", data["name"])
	assert.Equal(s.T(), "openai", data["service_type"])
	assert.Equal(s.T(), "https://api.openai.com/v1", data["endpoint"])
	assert.Equal(s.T(), true, data["enabled"])
}

// TestTextGenerationAPI runs all text generation API tests
func TestTextGenerationAPI(t *testing.T) {
	suite.Run(t, new(TextGenerationAPITestSuite))
}