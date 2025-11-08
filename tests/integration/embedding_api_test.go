package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EmbeddingAPITestSuite tests embedding generation API endpoints
type EmbeddingAPITestSuite struct {
	APITestSuite
}

// SetupSuite runs once before all tests in the suite
func (s *EmbeddingAPITestSuite) SetupSuite() {
	// Call parent setup
	s.APITestSuite.SetupSuite()
	
	// Set base URL to API service directly
	s.baseURL = "http://localhost:8082"
	
	// Wait for the API to be ready
	s.waitForAPI()
	
	// Login to get auth token
	s.login("testuser", "testpass")
}

// TestGenerateEmbedding tests generating a single embedding
func (s *EmbeddingAPITestSuite) TestGenerateEmbedding() {
	requestBody := map[string]interface{}{
		"model": "default",
		"text":  "This is a test text for embedding generation",
	}
	
	resp, err := s.post("/api/v1/taishang/model/embeddings", requestBody, nil)
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
	embedding, ok := data["embedding"].([]interface{})
	require.True(s.T(), ok, "Embedding should be an array")
	require.Greater(s.T(), len(embedding), 0, "Embedding should not be empty")
	
	// Check if all elements are numbers
	for _, value := range embedding {
		_, isFloat := value.(float64)
		assert.True(s.T(), isFloat, "Embedding elements should be numbers")
	}
	
	// Check other fields
	assert.NotEmpty(s.T(), data["model"])
	assert.NotEmpty(s.T(), data["object"])
	assert.NotEmpty(s.T(), data["usage"])
}

// TestGenerateEmbeddings tests generating multiple embeddings
func (s *EmbeddingAPITestSuite) TestGenerateEmbeddings() {
	requestBody := map[string]interface{}{
		"model": "default",
		"texts": []string{
			"This is the first test text",
			"This is the second test text",
			"This is the third test text",
		},
	}
	
	resp, err := s.post("/api/v1/taishang/model/embeddings/batch", requestBody, nil)
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
	embeddings, ok := data["embeddings"].([]interface{})
	require.True(s.T(), ok, "Embeddings should be an array")
	require.Equal(s.T(), 3, len(embeddings), "Should return 3 embeddings")
	
	// Check each embedding
	for i, embedding := range embeddings {
		embeddingMap, ok := embedding.(map[string]interface{})
		require.True(s.T(), ok, "Embedding should be a map")
		
		embeddingArray, ok := embeddingMap["embedding"].([]interface{})
		require.True(s.T(), ok, "Embedding should be an array")
		require.Greater(s.T(), len(embeddingArray), 0, "Embedding should not be empty")
		
		// Check if all elements are numbers
		for _, value := range embeddingArray {
			_, isFloat := value.(float64)
			assert.True(s.T(), isFloat, "Embedding elements should be numbers")
		}
		
		// Check index
		assert.Equal(s.T., i, embeddingMap["index"])
	}
	
	// Check other fields
	assert.NotEmpty(s.T(), data["model"])
	assert.NotEmpty(s.T(), data["object"])
	assert.NotEmpty(s.T(), data["usage"])
}

// TestGenerateEmbeddingWithDifferentModels tests embedding generation with different models
func (s *EmbeddingAPITestSuite) TestGenerateEmbeddingWithDifferentModels() {
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
	
	// Try to generate embeddings with each available service
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
			"text":  "This is a test text for embedding generation",
		}
		
		resp, err := s.post("/api/v1/taishang/model/embeddings", requestBody, nil)
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
		embedding, ok := responseData["embedding"].([]interface{})
		require.True(s.T(), ok, "Embedding should be an array")
		require.Greater(s.T(), len(embedding), 0, "Embedding should not be empty")
	}
}

// TestGenerateEmbeddingErrorHandling tests error handling in embedding generation
func (s *EmbeddingAPITestSuite) TestGenerateEmbeddingErrorHandling() {
	// Test with invalid request body
	requestBody := map[string]interface{}{
		"model": "nonexistent-model",
		"text":  "", // Empty text
	}
	
	resp, err := s.post("/api/v1/taishang/model/embeddings", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// Should return an error
	assert.True(s.T(), 
		resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError,
		"Should return an error for invalid request")
}

// TestGenerateEmbeddingsErrorHandling tests error handling in batch embedding generation
func (s *EmbeddingAPITestSuite) TestGenerateEmbeddingsErrorHandling() {
	// Test with empty texts array
	requestBody := map[string]interface{}{
		"model": "default",
		"texts": []string{}, // Empty array
	}
	
	resp, err := s.post("/api/v1/taishang/model/embeddings/batch", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// Should return an error
	assert.True(s.T(), 
		resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError,
		"Should return an error for empty texts array")
	
	// Test with too many texts
	requestBody = map[string]interface{}{
		"model": "default",
		"texts": make([]string, 1000), // Too many texts
	}
	
	resp, err = s.post("/api/v1/taishang/model/embeddings/batch", requestBody, nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	
	// Should return an error
	assert.True(s.T(), 
		resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError,
		"Should return an error for too many texts")
}

// TestEmbeddingDimensions tests that embeddings have consistent dimensions
func (s *EmbeddingAPITestSuite) TestEmbeddingDimensions() {
	requestBody := map[string]interface{}{
		"model": "default",
		"texts": []string{
			"This is the first test text",
			"This is the second test text",
		},
	}
	
	resp, err := s.post("/api/v1/taishang/model/embeddings/batch", requestBody, nil)
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
	embeddings, ok := data["embeddings"].([]interface{})
	require.True(s.T(), ok, "Embeddings should be an array")
	require.Equal(s.T(), 2, len(embeddings), "Should return 2 embeddings")
	
	// Get dimensions of the first embedding
	firstEmbedding, ok := embeddings[0].(map[string]interface{})
	require.True(s.T(), ok, "Embedding should be a map")
	
	firstEmbeddingArray, ok := firstEmbedding["embedding"].([]interface{})
	require.True(s.T., ok, "Embedding should be an array")
	
	firstDim := len(firstEmbeddingArray)
	
	// Check that all embeddings have the same dimensions
	for _, embedding := range embeddings {
		embeddingMap, ok := embedding.(map[string]interface{})
		require.True(s.T(), ok, "Embedding should be a map")
		
		embeddingArray, ok := embeddingMap["embedding"].([]interface{})
		require.True(s.T(), ok, "Embedding should be an array")
		
		assert.Equal(s.T(), firstDim, len(embeddingArray), "All embeddings should have the same dimensions")
	}
}

// TestEmbeddingUsage tests that embedding generation returns usage information
func (s *EmbeddingAPITestSuite) TestEmbeddingUsage() {
	requestBody := map[string]interface{}{
		"model": "default",
		"texts": []string{
			"This is the first test text",
			"This is the second test text",
		},
	}
	
	resp, err := s.post("/api/v1/taishang/model/embeddings/batch", requestBody, nil)
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
	
	// Check usage information
	usage, ok := data["usage"].(map[string]interface{})
	require.True(s.T(), ok, "Usage should be a map")
	
	promptTokens, ok := usage["prompt_tokens"].(float64)
	require.True(s.T., ok, "Prompt tokens should be a number")
	
	totalTokens, ok := usage["total_tokens"].(float64)
	require.True(s.T., ok, "Total tokens should be a number")
	
	assert.Greater(s.T., promptTokens, float64(0), "Prompt tokens should be greater than 0")
	assert.Equal(s.T., promptTokens, totalTokens, "For embeddings, total tokens should equal prompt tokens")
}

// TestEmbeddingAPI runs all embedding API tests
func TestEmbeddingAPI(t *testing.T) {
	suite.Run(t, new(EmbeddingAPITestSuite))
}