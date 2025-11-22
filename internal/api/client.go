// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/auth"
)

// APIClient defines the interface for interacting with the Challenge Service API
type APIClient interface {
	// M1 endpoints
	ListChallenges(ctx context.Context) ([]Challenge, error)
	ListChallengesWithFilter(ctx context.Context, activeOnly bool) ([]Challenge, error)
	GetChallenge(ctx context.Context, challengeID string) (*Challenge, error)
	ClaimReward(ctx context.Context, challengeID, goalID string) (*ClaimResult, error)

	// M3 endpoints
	InitializePlayer(ctx context.Context) (*InitializeResponse, error)
	SetGoalActive(ctx context.Context, challengeID, goalID string, isActive bool) (*SetGoalActiveResponse, error)

	// M4 endpoints
	BatchSelectGoals(ctx context.Context, challengeID string, req *BatchSelectRequest) (*BatchSelectResponse, error)
	RandomSelectGoals(ctx context.Context, challengeID string, req *RandomSelectRequest) (*RandomSelectResponse, error)

	// Debug
	GetLastRequest() *RequestDebugInfo
	GetLastResponse() *ResponseDebugInfo
}

// HTTPAPIClient implements APIClient using net/http
type HTTPAPIClient struct {
	baseURL      string
	httpClient   *http.Client
	authProvider auth.AuthProvider
	userID       string // User ID for mock authentication header

	// Debug instrumentation
	lastRequest  *RequestDebugInfo
	lastResponse *ResponseDebugInfo
}

// NewHTTPAPIClient creates a new HTTP API client
func NewHTTPAPIClient(baseURL string, authProvider auth.AuthProvider) *HTTPAPIClient {
	return &HTTPAPIClient{
		baseURL:      baseURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		authProvider: authProvider,
		userID:       "", // Will be set via SetUserID for mock auth
	}
}

// SetUserID sets the user ID for mock authentication (used when backend auth is disabled)
func (c *HTTPAPIClient) SetUserID(userID string) {
	c.userID = userID
}

// GetLastRequest returns the last recorded request for debugging
func (c *HTTPAPIClient) GetLastRequest() *RequestDebugInfo {
	return c.lastRequest
}

// GetLastResponse returns the last recorded response for debugging
func (c *HTTPAPIClient) GetLastResponse() *ResponseDebugInfo {
	return c.lastResponse
}

// ListChallenges retrieves all challenges with user progress
func (c *HTTPAPIClient) ListChallenges(ctx context.Context) ([]Challenge, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/challenges", nil)
	if err != nil {
		return nil, fmt.Errorf("list challenges: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var response GetChallengesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return response.Challenges, nil
}

// GetChallenge retrieves a specific challenge by ID
func (c *HTTPAPIClient) GetChallenge(ctx context.Context, challengeID string) (*Challenge, error) {
	path := fmt.Sprintf("/v1/challenges/%s", challengeID)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("get challenge: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var challenge Challenge
	if err := json.NewDecoder(resp.Body).Decode(&challenge); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &challenge, nil
}

// ClaimReward claims the reward for a completed goal
func (c *HTTPAPIClient) ClaimReward(ctx context.Context, challengeID, goalID string) (*ClaimResult, error) {
	path := fmt.Sprintf("/v1/challenges/%s/goals/%s/claim", challengeID, goalID)
	// Send empty JSON body ({}) as required by gRPC-Gateway for POST requests
	resp, err := c.doRequest(ctx, "POST", path, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("claim reward: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var result ClaimResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// M3: InitializePlayer initializes player goals with default assignments
func (c *HTTPAPIClient) InitializePlayer(ctx context.Context) (*InitializeResponse, error) {
	// Send empty JSON object as body (required by gRPC-Gateway)
	emptyBody := map[string]interface{}{}
	resp, err := c.doRequest(ctx, "POST", "/v1/challenges/initialize", emptyBody)
	if err != nil {
		return nil, fmt.Errorf("initialize player: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var result InitializeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// M3: SetGoalActive activates or deactivates a goal for the player
func (c *HTTPAPIClient) SetGoalActive(ctx context.Context, challengeID, goalID string, isActive bool) (*SetGoalActiveResponse, error) {
	path := fmt.Sprintf("/v1/challenges/%s/goals/%s/active", challengeID, goalID)
	// Use camelCase for JSON field name (matching gRPC-Gateway camelCase output)
	body := map[string]bool{"isActive": isActive}

	resp, err := c.doRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, fmt.Errorf("set goal active: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var result SetGoalActiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// M4: BatchSelectGoals activates multiple goals at once
func (c *HTTPAPIClient) BatchSelectGoals(ctx context.Context, challengeID string, req *BatchSelectRequest) (*BatchSelectResponse, error) {
	path := fmt.Sprintf("/v1/challenges/%s/goals/batch-select", challengeID)
	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, fmt.Errorf("batch select goals: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var result BatchSelectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// M4: RandomSelectGoals randomly activates N goals from a challenge
func (c *HTTPAPIClient) RandomSelectGoals(ctx context.Context, challengeID string, req *RandomSelectRequest) (*RandomSelectResponse, error) {
	path := fmt.Sprintf("/v1/challenges/%s/goals/random-select", challengeID)
	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, fmt.Errorf("random select goals: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var result RandomSelectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// M3: ListChallengesWithFilter retrieves all challenges with optional active_only filter
func (c *HTTPAPIClient) ListChallengesWithFilter(ctx context.Context, activeOnly bool) ([]Challenge, error) {
	path := "/v1/challenges"
	if activeOnly {
		path += "?active_only=true"
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("list challenges: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var response GetChallengesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return response.Challenges, nil
}

// doRequest performs an HTTP request with retry logic
func (c *HTTPAPIClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	// Serialize body if provided
	var reqBody io.Reader
	var bodyStr string
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
		bodyStr = string(jsonBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set mock user ID header if configured (for testing with auth disabled)
	if c.userID != "" {
		req.Header.Set("x-mock-user-id", c.userID)
	}

	// Get auth token
	token, err := c.authProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	// Record request for debug mode
	c.recordRequest(req, bodyStr)

	// Perform request with retry
	var resp *http.Response
	var lastErr error

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		startTime := time.Now()
		resp, lastErr = c.httpClient.Do(req)
		duration := time.Since(startTime)

		if lastErr != nil {
			continue
		}

		// Record response for debug mode
		c.recordResponse(resp, duration)

		// Check status code
		if resp.StatusCode >= 500 {
			// Server error, retry
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}

		// Success or client error (don't retry)
		return resp, nil
	}

	// All retries exhausted
	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// checkStatusCode checks if the response status code is OK
func (c *HTTPAPIClient) checkStatusCode(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read error response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
}

// recordRequest stores request details for debugging
func (c *HTTPAPIClient) recordRequest(req *http.Request, body string) {
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	c.lastRequest = &RequestDebugInfo{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: headers,
		Body:    body,
	}
}

// recordResponse stores response details for debugging
func (c *HTTPAPIClient) recordResponse(resp *http.Response, duration time.Duration) {
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Read body for debug (we'll need to restore it for caller)
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	c.lastResponse = &ResponseDebugInfo{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
		Duration:   duration,
	}
}
