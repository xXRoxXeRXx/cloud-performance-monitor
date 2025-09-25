package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// Client for interacting with the Dropbox API v2
type Client struct {
	AccessToken  string
	RefreshToken string
	AppKey       string
	AppSecret    string
	HTTPClient   *http.Client
	tokenMutex   sync.RWMutex
	logger       utils.ClientLogger
}

const (
	DefaultTimeout      = 300 * time.Second
	DropboxAPIURL       = "https://api.dropboxapi.com/2"
	DropboxContentURL   = "https://content.dropboxapi.com/2"
	DropboxOAuthURL     = "https://api.dropboxapi.com/oauth2/token"
	DropboxMaxChunkSize = 8 * 1024 * 1024 // 8MB chunks for chunked uploads
)

// OAuth2TokenResponse represents the OAuth2 token response
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// Session represents an upload session for chunked uploads
type UploadSession struct {
	SessionID string `json:"session_id"`
}

// UploadSessionStartResult contains the response from upload session start
type UploadSessionStartResult struct {
	SessionID string `json:"session_id"`
}

// UploadSessionAppendV2Args contains arguments for appending to upload session
type UploadSessionAppendV2Args struct {
	Cursor struct {
		SessionID string `json:"session_id"`
		Offset    uint64 `json:"offset"`
	} `json:"cursor"`
	Close bool `json:"close"`
}

// UploadSessionFinishArgs contains arguments for finishing upload session
type UploadSessionFinishArgs struct {
	Cursor struct {
		SessionID string `json:"session_id"`
		Offset    uint64 `json:"offset"`
	} `json:"cursor"`
	Commit struct {
		Path       string `json:"path"`
		Mode       string `json:"mode"`
		Autorename bool   `json:"autorename"`
	} `json:"commit"`
}

// FileMetadata represents file metadata from Dropbox API
type FileMetadata struct {
	Name           string    `json:"name"`
	ID             string    `json:"id"`
	Size           int64     `json:"size"`
	ServerModified time.Time `json:"server_modified"`
	PathLower      string    `json:"path_lower"`
	PathDisplay    string    `json:"path_display"`
}

// ErrorResponse represents an error response from Dropbox API
type ErrorResponse struct {
	ErrorSummary string `json:"error_summary"`
	Error        struct {
		Tag string `json:".tag"`
	} `json:"error"`
}

// NewClient creates a new Dropbox API client with OAuth2 refresh capability
func NewClient(accessToken, refreshToken, appKey, appSecret string, logger utils.ClientLogger) *Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	return &Client{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AppKey:       appKey,
		AppSecret:    appSecret,
		HTTPClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: t,
		},
		logger: logger,
	}
}

// NewClientWithOAuth2 creates a new Dropbox API client with OAuth2 refresh capability (alias for NewClient)
func NewClientWithOAuth2(accessToken, refreshToken, appKey, appSecret string, logger utils.ClientLogger) *Client {
	return NewClient(accessToken, refreshToken, appKey, appSecret, logger)
}

// RefreshAccessToken refreshes the access token using the refresh token with retry logic
func (c *Client) RefreshAccessToken() error {
	if c.RefreshToken == "" || c.AppKey == "" || c.AppSecret == "" {
		return fmt.Errorf("refresh token, app key, or app secret not available")
	}

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxRetries = 2 // Fewer retries for OAuth2 operations
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors, 
		"connection refused", "timeout", "temporary failure", "502", "503", "504")

	return retryConfig.WithRetry(context.Background(), "dropbox_oauth_refresh", func(ctx context.Context) error {
		data := url.Values{}
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", c.RefreshToken)

		req, err := http.NewRequestWithContext(ctx, "POST", DropboxOAuthURL, bytes.NewBufferString(data.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(c.AppKey, c.AppSecret)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.logger.LogOperation(utils.ERROR, "dropbox", "oauth", "token", "refresh_request_error", 
				fmt.Sprintf("Refresh request failed: %v", err), 
				map[string]interface{}{"error": err.Error()})
			return fmt.Errorf("refresh request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			c.logger.LogOperation(utils.ERROR, "dropbox", "oauth", "token", "refresh_status_error", 
				fmt.Sprintf("Refresh failed with status %d: %s", resp.StatusCode, string(body)), 
				map[string]interface{}{"status_code": resp.StatusCode, "response_body": string(body)})
			return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
		}

		var tokenResp OAuth2TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return fmt.Errorf("failed to decode refresh response: %v", err)
		}

		// Update access token
		c.AccessToken = tokenResp.AccessToken
		
		// Update refresh token if provided (some providers rotate refresh tokens)
		if tokenResp.RefreshToken != "" {
			c.RefreshToken = tokenResp.RefreshToken
		}

		c.logger.LogOperation(utils.INFO, "dropbox", "oauth", "token", "refresh_success", 
			fmt.Sprintf("Access token refreshed successfully (expires in %d seconds)", tokenResp.ExpiresIn), 
			map[string]interface{}{"expires_in": tokenResp.ExpiresIn})
		return nil
	})
}

// newAPIRequest creates a new authenticated API request
func (c *Client) newAPIRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	c.tokenMutex.RLock()
	token := c.AccessToken
	c.tokenMutex.RUnlock()

	fullURL := DropboxAPIURL + endpoint
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// doRequestWithRetry performs an HTTP request with automatic token refresh on 401 errors
func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	// Make the initial request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If request succeeded or refresh not available, return response
	if resp.StatusCode != http.StatusUnauthorized || c.RefreshToken == "" {
		return resp, nil
	}

	// Close the failed response
	resp.Body.Close()

	// Attempt to refresh token
	c.logger.LogOperation(utils.INFO, "dropbox", "oauth", "token", "refresh_attempt", 
		"Access token expired, attempting refresh...", 
		map[string]interface{}{})
	if err := c.RefreshAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %v", err)
	}

	// Update authorization header with new token
	c.tokenMutex.RLock()
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	c.tokenMutex.RUnlock()

	// Retry the request with new token
	c.logger.LogOperation(utils.DEBUG, "dropbox", "oauth", "token", "retry_request", 
		"Retrying request with refreshed token...", 
		map[string]interface{}{})
	return c.HTTPClient.Do(req)
}

// newContentRequest creates a new authenticated content request
func (c *Client) newContentRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	c.tokenMutex.RLock()
	token := c.AccessToken
	c.tokenMutex.RUnlock()

	fullURL := DropboxContentURL + endpoint
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}

// EnsureDirectory ensures the test directory exists (Dropbox creates folders automatically)
func (c *Client) EnsureDirectory(dirPath string) error {
	// Dropbox creates directories automatically when uploading files
	// So we just need to validate the path format
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}
	if dirPath[0] != '/' {
		return fmt.Errorf("directory path must start with /")
	}
	c.logger.LogOperation(utils.DEBUG, "dropbox", "api", "directory", "auto_create", 
		fmt.Sprintf("Directory %s will be created automatically on upload", dirPath), 
		map[string]interface{}{"dir_path": dirPath})
	return nil
}

// UploadFile uploads a file using chunked upload for large files or simple upload for small files
func (c *Client) UploadFile(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	c.logger.LogOperation(utils.INFO, "dropbox", "api", "upload", "start", 
		fmt.Sprintf("Starting upload for %s (%d bytes)", filePath, size), 
		map[string]interface{}{"file_path": filePath, "file_size": size, "chunk_size": chunkSize})
		
	if size <= DropboxMaxChunkSize {
		return c.uploadSimple(filePath, reader)
	}
	return c.uploadChunked(filePath, reader, size, chunkSize)
}

// uploadSimple uploads a file in a single request
func (c *Client) uploadSimple(filePath string, reader io.Reader) error {
	// Create the API args
	args := map[string]interface{}{
		"path":       filePath,
		"mode":       "overwrite",
		"autorename": false,
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal upload args: %v", err)
	}

	req, err := c.newContentRequest("POST", "/files/upload", reader)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	req.Header.Set("Dropbox-API-Arg", string(argsJSON))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "dropbox", "api", "upload", "request_error", 
			fmt.Sprintf("Upload request failed: %v", err), 
			map[string]interface{}{"file_path": filePath, "error": err.Error()})
		return fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.LogOperation(utils.ERROR, "dropbox", "api", "upload", "status_error", 
			fmt.Sprintf("Upload failed with status %d: %s", resp.StatusCode, string(body)), 
			map[string]interface{}{"file_path": filePath, "status_code": resp.StatusCode, "response_body": string(body)})
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "dropbox", "api", "upload", "simple_completed", 
		fmt.Sprintf("Simple upload completed for %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}

// uploadChunked uploads a file using chunked upload
func (c *Client) uploadChunked(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	// Start upload session
	sessionID, err := c.startUploadSession()
	if err != nil {
		return fmt.Errorf("failed to start upload session: %v", err)
	}

	c.logger.LogOperation(utils.INFO, "dropbox", "api", "upload", "session_started", 
		fmt.Sprintf("Started chunked upload session %s for %s (size: %d bytes)", sessionID, filePath, size), 
		map[string]interface{}{"session_id": sessionID, "file_path": filePath, "file_size": size})

	// Upload chunks
	var offset uint64 = 0
	chunkNum := 1
	buffer := make([]byte, chunkSize)

	for {
		n, err := io.ReadFull(reader, buffer)
		if err == io.EOF {
			break
		}
		if err != nil && err != io.ErrUnexpectedEOF {
			c.logger.LogOperation(utils.ERROR, "dropbox", "api", "upload", "read_error", 
				fmt.Sprintf("Failed to read chunk %d: %v", chunkNum, err), 
				map[string]interface{}{"chunk_num": chunkNum, "error": err.Error()})
			return fmt.Errorf("failed to read chunk %d: %v", chunkNum, err)
		}

		chunk := buffer[:n]
		isLast := offset+uint64(n) >= uint64(size)

		if isLast {
			// Finish the upload session
			err = c.finishUploadSession(sessionID, offset, bytes.NewReader(chunk), filePath)
		} else {
			// Append chunk to session
			err = c.appendUploadSession(sessionID, offset, bytes.NewReader(chunk))
		}

		if err != nil {
			c.logger.LogOperation(utils.ERROR, "dropbox", "api", "upload", "chunk_error", 
				fmt.Sprintf("Failed to upload chunk %d: %v", chunkNum, err), 
				map[string]interface{}{"chunk_num": chunkNum, "offset": offset, "error": err.Error()})
			return fmt.Errorf("failed to upload chunk %d: %v", chunkNum, err)
		}

		c.logger.LogOperation(utils.DEBUG, "dropbox", "api", "upload", "chunk_progress", 
			fmt.Sprintf("Uploaded chunk %d/%d (offset: %d, size: %d)", chunkNum, (size+chunkSize-1)/chunkSize, offset, n), 
			map[string]interface{}{"chunk_num": chunkNum, "total_chunks": (size + chunkSize - 1) / chunkSize, "offset": offset, "chunk_size": n})

		offset += uint64(n)
		chunkNum++

		if isLast {
			break
		}
	}

	c.logger.LogOperation(utils.INFO, "dropbox", "api", "upload", "chunked_completed", 
		fmt.Sprintf("Chunked upload completed for %s (%d chunks)", filePath, chunkNum-1), 
		map[string]interface{}{"file_path": filePath, "total_chunks": chunkNum - 1})
	return nil
}

// startUploadSession starts a new upload session
func (c *Client) startUploadSession() (string, error) {
	req, err := c.newContentRequest("POST", "/files/upload_session/start", bytes.NewReader([]byte{}))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("start session failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result UploadSessionStartResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode session start response: %v", err)
	}

	return result.SessionID, nil
}

// appendUploadSession appends data to an existing upload session
func (c *Client) appendUploadSession(sessionID string, offset uint64, data io.Reader) error {
	args := UploadSessionAppendV2Args{
		Close: false,
	}
	args.Cursor.SessionID = sessionID
	args.Cursor.Offset = offset

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return err
	}

	req, err := c.newContentRequest("POST", "/files/upload_session/append_v2", data)
	if err != nil {
		return err
	}
	req.Header.Set("Dropbox-API-Arg", string(argsJSON))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("append session failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// finishUploadSession finishes an upload session
func (c *Client) finishUploadSession(sessionID string, offset uint64, data io.Reader, filePath string) error {
	args := UploadSessionFinishArgs{}
	args.Cursor.SessionID = sessionID
	args.Cursor.Offset = offset
	args.Commit.Path = filePath
	args.Commit.Mode = "overwrite"
	args.Commit.Autorename = false

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return err
	}

	req, err := c.newContentRequest("POST", "/files/upload_session/finish", data)
	if err != nil {
		return err
	}
	req.Header.Set("Dropbox-API-Arg", string(argsJSON))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("finish session failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadFile downloads a file from Dropbox
func (c *Client) DownloadFile(filePath string) (io.ReadCloser, error) {
	args := map[string]string{
		"path": filePath,
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal download args: %v", err)
	}

	req, err := c.newContentRequest("POST", "/files/download", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %v", err)
	}
	req.Header.Set("Dropbox-API-Arg", string(argsJSON))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "dropbox", "api", "download", "started", 
		fmt.Sprintf("Download started for %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return resp.Body, nil
}

// DeleteFile deletes a file from Dropbox
func (c *Client) DeleteFile(filePath string) error {
	args := map[string]string{
		"path": filePath,
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal delete args: %v", err)
	}

	req, err := c.newAPIRequest("POST", "/files/delete_v2", bytes.NewReader(argsJSON))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %v", err)
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.LogOperation(utils.INFO, "dropbox", "api", "delete", "success", 
		fmt.Sprintf("File deleted: %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}

// GetFileInfo gets metadata for a file
func (c *Client) GetFileInfo(filePath string) (*FileMetadata, error) {
	args := map[string]interface{}{
		"path":                        filePath,
		"include_media_info":          false,
		"include_deleted":             false,
		"include_has_explicit_shared_members": false,
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get_metadata args: %v", err)
	}

	req, err := c.newAPIRequest("POST", "/files/get_metadata", bytes.NewReader(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create get_metadata request: %v", err)
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("get_metadata request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get_metadata failed with status %d: %s", resp.StatusCode, string(body))
	}

	var metadata FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata response: %v", err)
	}

	return &metadata, nil
}
