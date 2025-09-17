package hidrive_legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/MarcelWMeyer/cloud-performance-monitor/internal/utils"
)

// Client for interacting with the HiDrive HTTP REST API (Legacy)
type Client struct {
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	logger       utils.ClientLogger
}

const (
	DefaultTimeout    = 300 * time.Second
	HiDriveAPIBaseURL = "https://api.hidrive.strato.com/2.1"
	HiDriveOAuthURL   = "https://my.hidrive.com/oauth2/token"
)

// OAuth2TokenResponse represents the OAuth2 token response
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// FileInfo represents file information from HiDrive API
type FileInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	Path     string `json:"path"`
}

// UserInfo represents user information from HiDrive API
type UserInfo struct {
	Account string `json:"account"`
	Alias   string `json:"alias"`
	Home    string `json:"home"`
	HomeID  string `json:"home_id"`
}

// DirectoryInfo represents directory information
type DirectoryInfo struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Type  string     `json:"type"`
	Path  string     `json:"path"`
	Files []FileInfo `json:"members"`
}

// UploadResponse represents the response from file upload
type UploadResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// ErrorResponse represents an error response from HiDrive API
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// NewClient creates a new HiDrive Legacy API client
func NewClient(accessToken string) *Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	return &Client{
		AccessToken: accessToken,
		HTTPClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: t,
		},
		logger: &utils.DefaultClientLogger{},
	}
}

// NewClientWithOAuth2 creates a new HiDrive Legacy API client with OAuth2 refresh capability
func NewClientWithOAuth2(refreshToken, clientID, clientSecret string) (*Client, error) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	client := &Client{
		RefreshToken: refreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: t,
		},
		logger: &utils.DefaultClientLogger{},
	}

	// Generate initial access token
	if err := client.RefreshAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to generate initial access token: %v", err)
	}

	return client, nil
}

// GetAccessTokenFromCredentials exchanges client credentials for access token using OAuth2 with retry logic
func GetAccessTokenFromCredentials(clientID, clientSecret, authCode string) (*OAuth2TokenResponse, error) {
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxRetries = 2
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors, 
		"connection refused", "timeout", "temporary failure", "502", "503", "504")

	var tokenResp OAuth2TokenResponse
	err := retryConfig.WithRetry(context.Background(), "hidrive_legacy_oauth_initial", func(ctx context.Context) error {
		data := url.Values{}
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("grant_type", "authorization_code")
		data.Set("code", authCode)

		req, err := http.NewRequestWithContext(ctx, "POST", HiDriveOAuthURL, bytes.NewBufferString(data.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create OAuth2 request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{Timeout: DefaultTimeout}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("OAuth2 request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("OAuth2 failed with status %d: %s", resp.StatusCode, string(body))
		}

		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return fmt.Errorf("failed to decode OAuth2 response: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// RefreshAccessToken refreshes the access token using the refresh token with retry logic
func (c *Client) RefreshAccessToken() error {
	if c.RefreshToken == "" || c.ClientID == "" || c.ClientSecret == "" {
		return fmt.Errorf("refresh token, client ID, and client secret are required for token refresh")
	}

	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxRetries = 2 // Fewer retries for OAuth2 operations
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors, 
		"connection refused", "timeout", "temporary failure", "502", "503", "504")

	return retryConfig.WithRetry(context.Background(), "hidrive_legacy_oauth_refresh", func(ctx context.Context) error {
		data := url.Values{}
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", c.RefreshToken)
		data.Set("client_id", c.ClientID)
		data.Set("client_secret", c.ClientSecret)

		req, err := http.NewRequestWithContext(ctx, "POST", HiDriveOAuthURL, bytes.NewBufferString(data.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("refresh token request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("refresh token failed with status %d: %s", resp.StatusCode, string(body))
		}

		var tokenResp OAuth2TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return fmt.Errorf("failed to decode refresh token response: %v", err)
		}

		// Update access token
		c.AccessToken = tokenResp.AccessToken
		
		// Update refresh token if a new one was provided
		if tokenResp.RefreshToken != "" {
			c.RefreshToken = tokenResp.RefreshToken
		}

		c.logger.LogOperation(utils.INFO, "hidrive_legacy", "auth", "token_refresh", "success", 
			fmt.Sprintf("Access token refreshed successfully (expires in %d seconds)", tokenResp.ExpiresIn), 
			map[string]interface{}{"expires_in": tokenResp.ExpiresIn})
		return nil
	})
}

// newAPIRequest creates a new authenticated API request
func (c *Client) newAPIRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	fullURL := HiDriveAPIBaseURL + endpoint
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	return req, nil
}

// doRequestWithRetry performs an HTTP request with automatic token refresh on 401 errors
func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	// Clone the request to retry if needed
	originalBody := []byte{}
	if req.Body != nil {
		var err error
		originalBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(originalBody))
	}

	// First attempt
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If successful or refresh token not available, return response
	if resp.StatusCode != http.StatusUnauthorized || c.RefreshToken == "" {
		return resp, nil
	}

	// Close the response body
	resp.Body.Close()

	// Try to refresh the token
	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "auth", "token_refresh", "start", 
		"Received 401, attempting token refresh", nil)
	if err := c.RefreshAccessToken(); err != nil {
		return nil, fmt.Errorf("token refresh failed: %v", err)
	}

	// Recreate the request with new token
	newReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader(originalBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create retry request: %v", err)
	}

	// Copy headers from original request
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// Update authorization header with new token
	newReq.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// Retry the request
	return c.HTTPClient.Do(newReq)
}

// GetUserHome retrieves the user's home directory path
func (c *Client) GetUserHome() (string, error) {
	req, err := c.newAPIRequest("GET", "/user/me?fields=home", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create user info request: %v", err)
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return "", fmt.Errorf("user info request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info response: %v", err)
	}

	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "user_info", "success", 
		fmt.Sprintf("User home directory: %s", userInfo.Home), 
		map[string]interface{}{"home_directory": userInfo.Home})
	return userInfo.Home, nil
}

// EnsureDirectory ensures the test directory exists
func (c *Client) EnsureDirectory(dirPath string) error {
	// Get the user's home directory from API
	homePath, err := c.GetUserHome()
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "hidrive_legacy", "api", "directory", "home_error", 
			fmt.Sprintf("Failed to get user home directory: %v", err), 
			map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	// Remove root prefix from home path - API returns "root/users/myserver" but we need "/users/myserver"
	cleanHomePath := strings.TrimPrefix(homePath, "root")
	if !strings.HasPrefix(cleanHomePath, "/") {
		cleanHomePath = "/" + cleanHomePath
	}
	fullPath := cleanHomePath + "/" + dirPath
	
	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "directory", "create", 
		fmt.Sprintf("Creating directory %s (home: %s, cleanHome: %s, dirPath: %s)", fullPath, homePath, cleanHomePath, dirPath), 
		map[string]interface{}{"full_path": fullPath, "home_path": homePath, "clean_home": cleanHomePath, "dir_path": dirPath})
	
	// Check if directory exists
	req, err := c.newAPIRequest("GET", "/dir?path="+url.QueryEscape(fullPath), nil)
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "hidrive_legacy", "api", "directory", "check_error", 
			fmt.Sprintf("Failed to create directory check request: %v", err), 
			map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to create directory check request: %v", err)
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "hidrive_legacy", "api", "directory", "check_failed", 
			fmt.Sprintf("Directory check request failed: %v", err), 
			map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("directory check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "directory", "exists", 
			fmt.Sprintf("Directory %s already exists", fullPath), 
			map[string]interface{}{"full_path": fullPath})
		return nil
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "directory", "creating", 
		fmt.Sprintf("Directory does not exist, creating: %s", fullPath), 
		map[string]interface{}{"full_path": fullPath})

	// Directory doesn't exist, create it
	data := url.Values{}
	data.Set("path", fullPath)

	req, err = c.newAPIRequest("POST", "/dir", strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "hidrive_legacy", "api", "directory", "request_error", 
			fmt.Sprintf("Failed to create directory request: %v", err), 
			map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to create directory request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.doRequestWithRetry(req)
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "hidrive_legacy", "api", "directory", "creation_failed", 
			fmt.Sprintf("Directory creation request failed: %v", err), 
			map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("directory creation request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.LogOperation(utils.ERROR, "hidrive_legacy", "api", "directory", "creation_status_error", 
			fmt.Sprintf("Directory creation failed with status %d: %s", resp.StatusCode, string(body)), 
			map[string]interface{}{"status_code": resp.StatusCode, "response_body": string(body)})
		return fmt.Errorf("directory creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "directory", "created", 
		fmt.Sprintf("Directory %s created successfully", dirPath), 
		map[string]interface{}{"dir_path": dirPath})
	return nil
}

// UploadFile uploads a file using multipart upload or chunked upload for large files
func (c *Client) UploadFile(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	// Get user home directory and build full path
	homePath, err := c.GetUserHome()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	// Remove root prefix from home path and build full path
	cleanHomePath := strings.TrimPrefix(homePath, "root")
	if !strings.HasPrefix(cleanHomePath, "/") {
		cleanHomePath = "/" + cleanHomePath
	}
	fullPath := cleanHomePath + "/" + filePath
	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "upload", "start", 
		fmt.Sprintf("Uploading file to %s (home: %s, cleanHome: %s)", fullPath, homePath, cleanHomePath), 
		map[string]interface{}{"full_path": fullPath, "home_path": homePath, "clean_home": cleanHomePath})
	
	if size <= chunkSize {
		return c.uploadSimple(fullPath, reader, size)
	}
	return c.uploadChunked(fullPath, reader, size, chunkSize)
}

// uploadSimple uploads a file in a single request using multipart/form-data
func (c *Client) uploadSimple(filePath string, reader io.Reader, size int64) error {
	// Extract directory and filename from the full path
	dirPath := path.Dir(filePath)
	fileName := path.Base(filePath)
	
	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "upload", "simple", 
		fmt.Sprintf("Simple upload - dir: %s, filename: %s", dirPath, fileName), 
		map[string]interface{}{"dir_path": dirPath, "file_name": fileName})
	
	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file field with the correct filename (needed for HiDrive API)
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy file data
	_, err = io.Copy(fileWriter, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %v", err)
	}

	// Close the writer (no additional form fields needed for multipart upload)
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// Create request with only dir query parameter (filename is in the multipart form)
	endpoint := fmt.Sprintf("/file?dir=%s", url.QueryEscape(dirPath))
	req, err := c.newAPIRequest("POST", endpoint, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "upload", "completed", 
		fmt.Sprintf("Simple upload completed for %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}

// uploadChunked uploads a file using HiDrive's chunked upload (POST + PATCH)
func (c *Client) uploadChunked(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "upload", "chunked_start", 
		fmt.Sprintf("Starting chunked upload for %s (size: %d bytes, chunk size: %d)", filePath, size, chunkSize), 
		map[string]interface{}{"file_path": filePath, "size": size, "chunk_size": chunkSize})

	// Step 1: Create empty file with POST /file
	err := c.createEmptyFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to create empty file: %v", err)
	}

	// Step 2: Upload chunks sequentially using PATCH /file?offset=X
	var offset int64 = 0
	chunkNum := 1
	totalChunks := (size + chunkSize - 1) / chunkSize

	for offset < size {
		// Calculate chunk size for this iteration
		currentChunkSize := chunkSize
		if offset + chunkSize > size {
			currentChunkSize = size - offset
		}

		// Read chunk data from reader
		chunkData := make([]byte, currentChunkSize)
		bytesRead, err := io.ReadFull(reader, chunkData)
		if err != nil && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read chunk %d: %v", chunkNum, err)
		}
		
		// Adjust slice if we read less than expected (end of file)
		if int64(bytesRead) < currentChunkSize {
			chunkData = chunkData[:bytesRead]
			currentChunkSize = int64(bytesRead)
		}

		// Upload this chunk
		err = c.uploadChunkPatch(filePath, chunkData, offset)
		if err != nil {
			return fmt.Errorf("failed to upload chunk %d at offset %d: %v", chunkNum, offset, err)
		}

		c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "upload", "chunk_progress", 
			fmt.Sprintf("Uploaded chunk %d/%d (offset: %d, size: %d bytes)", chunkNum, totalChunks, offset, len(chunkData)), 
			map[string]interface{}{"chunk_num": chunkNum, "total_chunks": totalChunks, "offset": offset, "chunk_size": len(chunkData)})

		offset += currentChunkSize
		chunkNum++
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "upload", "chunked_completed", 
		fmt.Sprintf("Chunked upload completed for %s (%d chunks)", filePath, chunkNum-1), 
		map[string]interface{}{"file_path": filePath, "total_chunks": chunkNum - 1})
	return nil
}

// createEmptyFile creates an empty file on HiDrive using POST /file
func (c *Client) createEmptyFile(filePath string) error {
	// Extract directory and filename from the full path
	dirPath := path.Dir(filePath)
	fileName := path.Base(filePath)
	
	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "file", "create_empty", 
		fmt.Sprintf("Creating empty file - dir: %s, filename: %s", dirPath, fileName), 
		map[string]interface{}{"dir_path": dirPath, "filename": fileName})
	
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add empty file field with correct filename
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}
	
	// Write empty content
	_, err = fileWriter.Write([]byte{})
	if err != nil {
		return fmt.Errorf("failed to write empty content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// Create POST request with only dir query parameter (filename is in multipart form)
	endpoint := fmt.Sprintf("/file?dir=%s", url.QueryEscape(dirPath))
	req, err := c.newAPIRequest("POST", endpoint, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create empty file request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("empty file creation request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("empty file creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "file", "create_empty_success", 
		fmt.Sprintf("Empty file created: %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}

// uploadChunkPatch uploads a chunk using PATCH /file?offset=X
func (c *Client) uploadChunkPatch(filePath string, chunkData []byte, offset int64) error {
	// Create PATCH request with binary data
	requestBody := bytes.NewReader(chunkData)
	
	// Build URL with offset parameter
	endpoint := fmt.Sprintf("/file?path=%s&offset=%d", url.QueryEscape(filePath), offset)
	
	req, err := c.newAPIRequest("PATCH", endpoint, requestBody)
	if err != nil {
		return fmt.Errorf("failed to create PATCH request: %v", err)
	}
	
	// Set content type to application/octet-stream (required for PATCH)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("PATCH request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PATCH upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadFile downloads a file from HiDrive
func (c *Client) DownloadFile(filePath string) (io.ReadCloser, error) {
	// Get user home directory and build full path
	homePath, err := c.GetUserHome()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	// Remove root prefix from home path and build full path
	cleanHomePath := strings.TrimPrefix(homePath, "root")
	if !strings.HasPrefix(cleanHomePath, "/") {
		cleanHomePath = "/" + cleanHomePath
	}
	fullPath := cleanHomePath + "/" + filePath
	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "download", "path_constructed", 
		fmt.Sprintf("Downloading file from %s (home: %s, cleanHome: %s)", fullPath, homePath, cleanHomePath), 
		map[string]interface{}{"full_path": fullPath, "home_path": homePath, "clean_home": cleanHomePath, "file_path": filePath})
	
	req, err := c.newAPIRequest("GET", "/file?path="+url.QueryEscape(fullPath), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %v", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "download", "started", 
		fmt.Sprintf("Download started for %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return resp.Body, nil
}

// DeleteFile deletes a file from HiDrive
func (c *Client) DeleteFile(filePath string) error {
	// Get user home directory and build full path
	homePath, err := c.GetUserHome()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	// Remove root prefix from home path and build full path
	cleanHomePath := strings.TrimPrefix(homePath, "root")
	if !strings.HasPrefix(cleanHomePath, "/") {
		cleanHomePath = "/" + cleanHomePath
	}
	fullPath := cleanHomePath + "/" + filePath
	c.logger.LogOperation(utils.DEBUG, "hidrive_legacy", "api", "delete", "path_constructed", 
		fmt.Sprintf("Deleting file %s (home: %s, cleanHome: %s)", fullPath, homePath, cleanHomePath), 
		map[string]interface{}{"full_path": fullPath, "home_path": homePath, "clean_home": cleanHomePath, "file_path": filePath})
	
	data := url.Values{}
	data.Set("path", fullPath)

	req, err := c.newAPIRequest("DELETE", "/file", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "delete", "success", 
		fmt.Sprintf("File deleted: %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}

// GetFileInfo gets metadata for a file
func (c *Client) GetFileInfo(filePath string) (*FileInfo, error) {
	cleanPath := strings.TrimPrefix(filePath, "/")
	
	req, err := c.newAPIRequest("GET", "/file?path="+url.QueryEscape(cleanPath)+"&fields=id,name,type,size,modified,path", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file info request: %v", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("file info request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file info failed with status %d: %s", resp.StatusCode, string(body))
	}

	var fileInfo FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, fmt.Errorf("failed to decode file info response: %v", err)
	}

	return &fileInfo, nil
}

// TestConnection tests the connection to HiDrive API
func (c *Client) TestConnection() error {
	req, err := c.newAPIRequest("GET", "/app/me?fields=id,name", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %v", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("connection test failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "hidrive_legacy", "api", "connection", "test_success", 
		"Connection test successful", 
		map[string]interface{}{})
	return nil
}
