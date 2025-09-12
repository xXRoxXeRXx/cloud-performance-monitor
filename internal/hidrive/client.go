package hidrive

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"
)

// Client for interacting with the HiDrive Next WebDAV API
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

const (
	DefaultTimeout = 300 * time.Second
)

// NewClient creates a new HiDrive WebDAV client
func NewClient(baseURL, username, password string) *Client {
       t := http.DefaultTransport.(*http.Transport).Clone()
       t.MaxIdleConns = 100
       t.MaxConnsPerHost = 100
       t.MaxIdleConnsPerHost = 100
       return &Client{
	       BaseURL:    baseURL,
	       Username:   username,
	       Password:   password,
	       HTTPClient: &http.Client{
		       Timeout:   DefaultTimeout,
		       Transport: t,
	       },
       }
}

// newRequest is a helper to create authenticated WebDAV requests
func (c *Client) newRequest(method, urlPath string, body io.Reader) (*http.Request, error) {
	fullURL := c.BaseURL + urlPath
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	return req, nil
}

// EnsureDirectory ensures the test directory exists
func (c *Client) EnsureDirectory(dirPath string) error {
	fullPath := path.Join("/remote.php/dav/files/", c.Username, dirPath)
	req, err := c.newRequest("MKCOL", fullPath, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 405 is returned if the directory already exists, which is not an error for us.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusMethodNotAllowed {
		return fmt.Errorf("failed to create directory %s, status: %s", dirPath, resp.Status)
	}
	return nil
}

// UploadFile uploads a file using the chunking API
func (c *Client) UploadFile(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	transferID := uuid.New().String()
	chunkDir := path.Join("/remote.php/dav/uploads/", c.Username, transferID)

	// 1. Create temporary directory for chunks on the server
	req, err := http.NewRequest("MKCOL", c.BaseURL+chunkDir, nil)
	if err != nil {
		return fmt.Errorf("could not create MKCOL request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("MKCOL request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("MKCOL for chunks failed with status %s", resp.Status)
	}

	// 2. Upload file in chunks
	if err := c.uploadChunks(chunkDir, reader, chunkSize); err != nil {
		return err
	}

	// 3. Assemble chunks by moving the directory
	destinationURL := c.BaseURL + path.Join("/remote.php/dav/files/", c.Username, filePath)
	req, err = http.NewRequest("MOVE", c.BaseURL+chunkDir+"/.file", nil)
	if err != nil {
		return fmt.Errorf("could not create MOVE request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Destination", destinationURL)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("MOVE request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("final MOVE to assemble chunks failed with status %s", resp.Status)
	}

	log.Printf("Chunked upload successful for %s", filePath)
	return nil
}

// uploadChunks uploads the file in chunks to the server
func (c *Client) uploadChunks(chunkDir string, reader io.Reader, chunkSize int64) error {
	chunk := make([]byte, chunkSize)
	for i := 0; ; i++ {
		bytesRead, readErr := reader.Read(chunk)
		if bytesRead > 0 {
			chunkPath := fmt.Sprintf("%s/%016d", chunkDir, i)
			chunkReader := bytes.NewReader(chunk[:bytesRead])

			req, err := http.NewRequest("PUT", c.BaseURL+chunkPath, chunkReader)
			if err != nil {
				return fmt.Errorf("could not create PUT request for chunk %d: %w", i, err)
			}
			req.SetBasicAuth(c.Username, c.Password)

			resp, err := c.HTTPClient.Do(req)
			if err != nil {
				return fmt.Errorf("PUT request for chunk %d failed: %w", i, err)
			}
			
			// Immediately close response body to avoid resource leaks
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
				resp.Body.Close()
				return fmt.Errorf("upload of chunk %d failed with status %s", i, resp.Status)
			}
			resp.Body.Close()
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("failed to read chunk %d: %w", i, readErr)
		}
	}
	return nil
}

// DownloadFile downloads a file
func (c *Client) DownloadFile(filePath string) (io.ReadCloser, error) {
	fullPath := path.Join("/remote.php/dav/files/", c.Username, filePath)
	req, err := c.newRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed, status: %s", resp.Status)
	}
	return resp.Body, nil
}

// DeleteFile deletes a file or directory
func (c *Client) DeleteFile(filePath string) error {
	fullPath := path.Join("/remote.php/dav/files/", c.Username, filePath)
	req, err := c.newRequest("DELETE", fullPath, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed, status: %s", resp.Status)
	}
	return nil
}
