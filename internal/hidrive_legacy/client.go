package hidrive_legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// Client for interacting with the HiDrive HTTP REST API (Legacy)
type Client struct {
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
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
	}

	// Generate initial access token
	if err := client.RefreshAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to generate initial access token: %v", err)
	}

	return client, nil
}

// GetAccessTokenFromCredentials exchanges client credentials for access token using OAuth2
func GetAccessTokenFromCredentials(clientID, clientSecret, authCode string) (*OAuth2TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)

	resp, err := http.PostForm(HiDriveOAuthURL, data)
	if err != nil {
		return nil, fmt.Errorf("OAuth2 request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OAuth2 failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuth2TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode OAuth2 response: %v", err)
	}

	return &tokenResp, nil
}

// RefreshAccessToken refreshes the access token using the refresh token
func (c *Client) RefreshAccessToken() error {
	if c.RefreshToken == "" || c.ClientID == "" || c.ClientSecret == "" {
		return fmt.Errorf("refresh token, client ID, and client secret are required for token refresh")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.RefreshToken)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)

	resp, err := http.PostForm(HiDriveOAuthURL, data)
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

	log.Printf("HiDrive Legacy: Access token refreshed successfully (expires in %d seconds)", tokenResp.ExpiresIn)
	return nil
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
	log.Printf("HiDrive Legacy: Received 401, attempting token refresh...")
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

	log.Printf("HiDrive Legacy: User home directory: %s", userInfo.Home)
	return userInfo.Home, nil
}

// EnsureDirectory ensures the test directory exists
func (c *Client) EnsureDirectory(dirPath string) error {
	// Get the user's home directory from API
	homePath, err := c.GetUserHome()
	if err != nil {
		log.Printf("HiDrive Legacy: Failed to get user home directory: %v", err)
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	// Remove root prefix from home path - API returns "root/users/myserver" but we need "/users/myserver"
	cleanHomePath := strings.TrimPrefix(homePath, "root")
	if !strings.HasPrefix(cleanHomePath, "/") {
		cleanHomePath = "/" + cleanHomePath
	}
	fullPath := cleanHomePath + "/" + dirPath
	
	log.Printf("HiDrive Legacy: Creating directory %s (home: %s, cleanHome: %s, dirPath: %s)", fullPath, homePath, cleanHomePath, dirPath)
	
	// Check if directory exists
	req, err := c.newAPIRequest("GET", "/dir?path="+url.QueryEscape(fullPath), nil)
	if err != nil {
		log.Printf("HiDrive Legacy: Failed to create directory check request: %v", err)
		return fmt.Errorf("failed to create directory check request: %v", err)
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		log.Printf("HiDrive Legacy: Directory check request failed: %v", err)
		return fmt.Errorf("directory check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("HiDrive Legacy: Directory %s already exists", fullPath)
		return nil
	}

	log.Printf("HiDrive Legacy: Directory does not exist, creating: %s", fullPath)

	// Directory doesn't exist, create it
	data := url.Values{}
	data.Set("path", fullPath)

	req, err = c.newAPIRequest("POST", "/dir", strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("HiDrive Legacy: Failed to create directory request: %v", err)
		return fmt.Errorf("failed to create directory request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.doRequestWithRetry(req)
	if err != nil {
		log.Printf("HiDrive Legacy: Directory creation request failed: %v", err)
		return fmt.Errorf("directory creation request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("HiDrive Legacy: Directory creation failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("directory creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("HiDrive Legacy: Directory %s created successfully", dirPath)
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
	log.Printf("HiDrive Legacy: Uploading file to %s (home: %s, cleanHome: %s)", fullPath, homePath, cleanHomePath)
	
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
	
	log.Printf("HiDrive Legacy: Simple upload - dir: %s, filename: %s", dirPath, fileName)
	
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

	log.Printf("HiDrive Legacy: Simple upload completed for %s", filePath)
	return nil
}

// uploadChunked uploads a file using HiDrive's chunked upload (POST + PATCH)
func (c *Client) uploadChunked(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	log.Printf("HiDrive Legacy: Starting chunked upload for %s (size: %d bytes, chunk size: %d)", filePath, size, chunkSize)

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

		log.Printf("HiDrive Legacy: Uploaded chunk %d/%d (offset: %d, size: %d bytes)", 
			chunkNum, totalChunks, offset, len(chunkData))

		offset += currentChunkSize
		chunkNum++
	}

	log.Printf("HiDrive Legacy: Chunked upload completed for %s (%d chunks)", filePath, chunkNum-1)
	return nil
}

// createEmptyFile creates an empty file on HiDrive using POST /file
func (c *Client) createEmptyFile(filePath string) error {
	// Extract directory and filename from the full path
	dirPath := path.Dir(filePath)
	fileName := path.Base(filePath)
	
	log.Printf("HiDrive Legacy: Creating empty file - dir: %s, filename: %s", dirPath, fileName)
	
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

	log.Printf("HiDrive Legacy: Empty file created: %s", filePath)
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
	log.Printf("HiDrive Legacy: Downloading file from %s (home: %s, cleanHome: %s)", fullPath, homePath, cleanHomePath)
	
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

	log.Printf("HiDrive Legacy: Download started for %s", filePath)
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
	log.Printf("HiDrive Legacy: Deleting file %s (home: %s, cleanHome: %s)", fullPath, homePath, cleanHomePath)
	
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

	log.Printf("HiDrive Legacy: File deleted: %s", filePath)
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

	log.Printf("HiDrive Legacy: Connection test successful")
	return nil
}
