// Package sdk provides the official Go SDK for AgentCloudService.
//
// AgentCloudService is a cloud hosting platform for autonomous AI agents.
// No CAPTCHA, no phone verification. Just API keys and code.
//
// Quick Start:
//
//	import "acs.vibecaas.app/sdk"
//
//	// Register a new agent (no auth required)
//	result, err := sdk.Register(sdk.RegisterOptions{Name: "my-agent"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("API Key:", result.APIKey)
//
//	// Create client
//	client := sdk.NewClient(result.APIKey)
//
//	// Deploy an agent
//	agent, err := client.Deploy(sdk.DeployOptions{
//	    Name: "research-agent",
//	    Type: "openclaw",
//	    Config: map[string]interface{}{
//	        "model":    "claude-sonnet-4",
//	        "channels": []string{"telegram"},
//	    },
//	})
package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	// DefaultBaseURL is the default ACS API endpoint
	DefaultBaseURL = "https://acs-api-v2.vibecaas.app/api/v1"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
	// Version is the SDK version
	Version = "1.0.0"
)

// Client is the ACS API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets a custom timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new ACS client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("ACS_API_KEY")
	}

	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Agent represents a deployed agent
type Agent struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Plan        string  `json:"plan"`
	Region      string  `json:"region"`
	URL         string  `json:"url,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	Requests24h int     `json:"requests24h,omitempty"`
	ComputeUsed float64 `json:"computeUsed,omitempty"`
}

// RegisterOptions for registering a new agent
type RegisterOptions struct {
	Name   string `json:"name"`
	Email  string `json:"email,omitempty"`
	Wallet string `json:"wallet,omitempty"`
}

// RegisterResult from registration
type RegisterResult struct {
	AgentID string `json:"agentId"`
	APIKey  string `json:"apiKey"`
	Message string `json:"message"`
}

// DeployOptions for deploying an agent
type DeployOptions struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
	Region string                 `json:"region"`
}

// Usage metrics
type Usage struct {
	Requests24h int     `json:"requests24h"`
	ComputeUsed float64 `json:"computeUsed"`
	StorageUsed float64 `json:"storageUsed"`
	CurrentBill float64 `json:"currentBill"`
}

// CheckoutResult from checkout session
type CheckoutResult struct {
	CheckoutURL string `json:"checkoutUrl"`
	SessionID   string `json:"sessionId"`
}

// ServiceStatus represents service health
type ServiceStatus struct {
	Status   string                 `json:"status"`
	Services map[string]interface{} `json:"services"`
	Regions  map[string]interface{} `json:"regions"`
}

// Error types
type ACSError struct {
	Message    string
	StatusCode int
	Response   interface{}
}

func (e *ACSError) Error() string {
	return fmt.Sprintf("ACS Error: %s (status: %d)", e.Message, e.StatusCode)
}

// AuthenticationError indicates invalid credentials
type AuthenticationError struct {
	*ACSError
}

// ValidationError indicates invalid request data
type ValidationError struct {
	*ACSError
}

func (c *Client) request(method, endpoint string, body interface{}, requireAuth bool) ([]byte, error) {
	if requireAuth && c.apiKey == "" {
		return nil, &AuthenticationError{&ACSError{
			Message:    "API key required. Set apiKey or ACS_API_KEY env var.",
			StatusCode: 401,
		}}
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "acs-sdk-go/"+Version)

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, &AuthenticationError{&ACSError{
			Message:    "Invalid API key",
			StatusCode: 401,
		}}
	}

	if resp.StatusCode == 400 {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		msg := "Validation failed"
		if e, ok := errResp["error"].(string); ok {
			msg = e
		}
		return nil, &ValidationError{&ACSError{
			Message:    msg,
			StatusCode: 400,
			Response:   errResp,
		}}
	}

	if resp.StatusCode >= 400 {
		return nil, &ACSError{
			Message:    fmt.Sprintf("Request failed with status %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
		}
	}

	return respBody, nil
}

// === Static Functions (No Auth Required) ===

// Register creates a new agent and returns API key
func Register(opts RegisterOptions, baseURL ...string) (*RegisterResult, error) {
	apiURL := DefaultBaseURL
	if len(baseURL) > 0 {
		apiURL = baseURL[0]
	}

	jsonBody, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL+"/register", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.Unmarshal(body, &errResp)
		msg := "Registration failed"
		if e, ok := errResp["error"].(string); ok {
			msg = e
		}
		return nil, &ACSError{Message: msg, StatusCode: resp.StatusCode}
	}

	var result RegisterResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Status gets ACS service status
func Status(baseURL ...string) (*ServiceStatus, error) {
	apiURL := DefaultBaseURL
	if len(baseURL) > 0 {
		apiURL = baseURL[0]
	}

	resp, err := http.Get(apiURL + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status ServiceStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// === Client Methods ===

// Deploy creates a new agent instance
func (c *Client) Deploy(opts DeployOptions) (*Agent, error) {
	if opts.Type == "" {
		opts.Type = "openclaw"
	}
	if opts.Region == "" {
		opts.Region = "us-east-1"
	}
	if opts.Config == nil {
		opts.Config = make(map[string]interface{})
	}

	body, err := c.request("POST", "/deploy", opts, true)
	if err != nil {
		return nil, err
	}

	var result struct {
		InstanceID string `json:"instanceId"`
		URL        string `json:"url"`
		Status     string `json:"status"`
		Plan       string `json:"plan"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &Agent{
		ID:     result.InstanceID,
		Name:   opts.Name,
		Status: result.Status,
		Plan:   result.Plan,
		Region: opts.Region,
		URL:    result.URL,
	}, nil
}

// ListAgents returns all deployed agents
func (c *Client) ListAgents(status ...string) ([]Agent, error) {
	endpoint := "/agents"
	if len(status) > 0 && status[0] != "" {
		endpoint += "?status=" + url.QueryEscape(status[0])
	}

	body, err := c.request("GET", endpoint, nil, true)
	if err != nil {
		return nil, err
	}

	var agents []Agent
	if err := json.Unmarshal(body, &agents); err != nil {
		return nil, err
	}

	return agents, nil
}

// GetAgent returns agent details
func (c *Client) GetAgent(agentID string) (*Agent, error) {
	body, err := c.request("GET", "/agents/"+agentID, nil, true)
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := json.Unmarshal(body, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

// StartAgent starts a stopped agent
func (c *Client) StartAgent(agentID string) (*Agent, error) {
	body, err := c.request("POST", "/agents/"+agentID+"/start", nil, true)
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := json.Unmarshal(body, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

// StopAgent stops a running agent
func (c *Client) StopAgent(agentID string) (*Agent, error) {
	body, err := c.request("POST", "/agents/"+agentID+"/stop", nil, true)
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := json.Unmarshal(body, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

// DeleteAgent deletes an agent and its resources
func (c *Client) DeleteAgent(agentID string) error {
	_, err := c.request("DELETE", "/agents/"+agentID, nil, true)
	return err
}

// GetUsage returns current usage metrics
func (c *Client) GetUsage() (*Usage, error) {
	body, err := c.request("GET", "/usage", nil, true)
	if err != nil {
		return nil, err
	}

	var usage Usage
	if err := json.Unmarshal(body, &usage); err != nil {
		return nil, err
	}

	return &usage, nil
}

// GetBilling returns billing information
func (c *Client) GetBilling() (map[string]interface{}, error) {
	body, err := c.request("GET", "/billing", nil, true)
	if err != nil {
		return nil, err
	}

	var billing map[string]interface{}
	if err := json.Unmarshal(body, &billing); err != nil {
		return nil, err
	}

	return billing, nil
}

// Checkout creates a Stripe checkout session
func (c *Client) Checkout(plan string) (*CheckoutResult, error) {
	body, err := c.request("POST", "/checkout", map[string]string{"plan": plan}, true)
	if err != nil {
		return nil, err
	}

	var result CheckoutResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PayUSDC pays with USDC via x402 protocol
func (c *Client) PayUSDC(amount string) (map[string]interface{}, error) {
	body, err := c.request("POST", "/x402/pay", map[string]string{"amount": amount}, true)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
