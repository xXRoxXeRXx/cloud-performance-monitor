package nextcloud

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnsureDirectory(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "MKCOL" {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "testuser", "testpass")

	err := client.EnsureDirectory("testdir")
	if err != nil {
		t.Errorf("EnsureDirectory failed: %v", err)
	}
}

func TestDownloadFile(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test content"))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "testuser", "testpass")

	body, err := client.DownloadFile("testfile.txt")
	if err != nil {
		t.Errorf("DownloadFile failed: %v", err)
	}
	defer body.Close()

	// Read and check content
	buf := make([]byte, 1024)
	n, _ := body.Read(buf)
	if string(buf[:n]) != "test content" {
		t.Errorf("Unexpected content: %s", string(buf[:n]))
	}
}

func TestDeleteFile(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "testuser", "testpass")

	err := client.DeleteFile("testfile.txt")
	if err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}
}
