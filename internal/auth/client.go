// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Authentication client implementation.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Credentials represents the authentication credentials
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	CreatedAt    time.Time `json:"created_at"`
}

// Client handles the authentication process
type Client struct {
	server *http.Server
	creds  chan *Credentials
	err    chan error
}

// NewClient creates a new authentication client
func NewClient() *Client {
	return &Client{
		creds: make(chan *Credentials),
		err:   make(chan error),
	}
}

// WaitForCallback starts a local server and waits for the authentication callback
func (c *Client) WaitForCallback() (*Credentials, error) {
	// Start local server to handle callback
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", c.handleCallback)

	c.server = &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := c.server.ListenAndServe(); err != http.ErrServerClosed {
			c.err <- fmt.Errorf("server error: %v", err)
		}
	}()

	// Wait for callback or error
	select {
	case creds := <-c.creds:
		return creds, nil
	case err := <-c.err:
		return nil, err
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timed out")
	}
}

func (c *Client) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	creds.CreatedAt = time.Now()
	c.creds <- &creds

	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authentication successful! You can close this window."))

	// Shutdown server
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.server.Shutdown(ctx)
	}()
}

// SaveToFile saves the credentials to a file
func (c *Credentials) SaveToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadFromFile loads credentials from a file
func LoadFromFile(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// IsExpired checks if the credentials are expired
func (c *Credentials) IsExpired() bool {
	expiresAt := c.CreatedAt.Add(time.Duration(c.ExpiresIn) * time.Second)
	return time.Now().After(expiresAt)
}
