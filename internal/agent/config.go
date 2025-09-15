package agent

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the configuration for a single storage instance (Nextcloud, HiDrive, HiDrive Legacy, Dropbox, or MagentaCLOUD)
type Config struct {
	InstanceName    string
	ServiceType     string // "nextcloud", "hidrive", "hidrive_legacy", "dropbox", or "magentacloud"
	URL             string
	Username        string
	Password        string
	ANID            string // For MagentaCLOUD Account Number ID
	AccessToken     string // For HiDrive Legacy API only
	RefreshToken    string // For Dropbox OAuth2 and HiDrive Legacy OAuth2
	AppKey          string // For Dropbox OAuth2 (App Key)
	AppSecret       string // For Dropbox OAuth2 (App Secret)  
	ClientID        string // For HiDrive Legacy OAuth2
	ClientSecret    string // For HiDrive Legacy OAuth2
	TestFileSizeMB  int
	TestIntervalSec int
	TestChunkSizeMB int
}

const (
	DefaultFileSizeMB  = 10
	DefaultIntervalSec = 300
	DefaultChunkSizeMB = 5 // Kleinere Chunks für HiDrive (5MB statt 10MB)
)

// LoadConfigs loads configurations for all specified Nextcloud, HiDrive, HiDrive Legacy, Dropbox, and MagentaCLOUD instances
func LoadConfigs() ([]*Config, error) {
	var configs []*Config

	// Nextcloud Instanzen
	i := 1
	for {
		urlKey := fmt.Sprintf("NC_INSTANCE_%d_URL", i)
		userKey := fmt.Sprintf("NC_INSTANCE_%d_USER", i)
		passKey := fmt.Sprintf("NC_INSTANCE_%d_PASS", i)
		url := os.Getenv(urlKey)
		if i == 1 && url == "" {
			// Noch nicht zwingend Fehler, andere Services könnten konfiguriert sein
			break
		}
		if url == "" {
			break
		}
		user := os.Getenv(userKey)
		pass := os.Getenv(passKey)
		if user == "" || pass == "" {
			return nil, fmt.Errorf("error: %s and %s must be set for instance %d", userKey, passKey, i)
		}
		fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
		if fileSize == 0 {
			fileSize = DefaultFileSizeMB
		}
		if fileSize <= 0 {
			return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
		}
		interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
		if interval == 0 {
			interval = DefaultIntervalSec
		}
		if interval <= 0 {
			return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
		}
		chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
		if chunkSize == 0 {
			chunkSize = DefaultChunkSizeMB
		}
		if chunkSize <= 0 {
			return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
		}
		config := &Config{
			InstanceName:    url,
			ServiceType:     "nextcloud",
			URL:             url,
			Username:        user,
			Password:        pass,
			TestFileSizeMB:  fileSize,
			TestIntervalSec: interval,
			TestChunkSizeMB: chunkSize,
		}
		configs = append(configs, config)
		i++
	}

	// HiDrive Instanzen (WebDAV)
	j := 1
	for {
		urlKey := fmt.Sprintf("HIDRIVE_INSTANCE_%d_URL", j)
		userKey := fmt.Sprintf("HIDRIVE_INSTANCE_%d_USER", j)
		passKey := fmt.Sprintf("HIDRIVE_INSTANCE_%d_PASS", j)
		url := os.Getenv(urlKey)
		if j == 1 && url == "" {
			// No HiDrive instances configured, continue to next service type
			break
		}
		if url == "" {
			break
		}
		user := os.Getenv(userKey)
		pass := os.Getenv(passKey)
		if user == "" || pass == "" {
			return nil, fmt.Errorf("error: %s and %s must be set for instance %d", userKey, passKey, j)
		}
		fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
		if fileSize == 0 {
			fileSize = DefaultFileSizeMB
		}
		if fileSize <= 0 {
			return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
		}
		interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
		if interval == 0 {
			interval = DefaultIntervalSec
		}
		if interval <= 0 {
			return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
		}
		chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
		if chunkSize == 0 {
			chunkSize = DefaultChunkSizeMB
		}
		if chunkSize <= 0 {
			return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
		}
		config := &Config{
			InstanceName:    url,
			ServiceType:     "hidrive",
			URL:             url,
			Username:        user,
			Password:        pass,
			TestFileSizeMB:  fileSize,
			TestIntervalSec: interval,
			TestChunkSizeMB: chunkSize,
		}
		configs = append(configs, config)
		j++
	}

	// HiDrive Legacy Instanzen (HTTP REST API)
	l := 1
	for {
		refreshTokenKey := fmt.Sprintf("HIDRIVE_LEGACY_INSTANCE_%d_REFRESH_TOKEN", l)
		clientIDKey := fmt.Sprintf("HIDRIVE_LEGACY_INSTANCE_%d_CLIENT_ID", l)
		clientSecretKey := fmt.Sprintf("HIDRIVE_LEGACY_INSTANCE_%d_CLIENT_SECRET", l)
		nameKey := fmt.Sprintf("HIDRIVE_LEGACY_INSTANCE_%d_NAME", l)

		refreshToken := os.Getenv(refreshTokenKey)
		clientID := os.Getenv(clientIDKey)
		clientSecret := os.Getenv(clientSecretKey)

		if l == 1 && (refreshToken == "" || clientID == "" || clientSecret == "") {
			// No HiDrive Legacy instances configured, continue to next service type
			break
		}
		if refreshToken == "" || clientID == "" || clientSecret == "" {
			break
		}

		instanceName := os.Getenv(nameKey)
		if instanceName == "" {
			instanceName = fmt.Sprintf("hidrive-legacy-instance-%d", l)
		}

		fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
		if fileSize == 0 {
			fileSize = DefaultFileSizeMB
		}
		if fileSize <= 0 {
			return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
		}
		interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
		if interval == 0 {
			interval = DefaultIntervalSec
		}
		if interval <= 0 {
			return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
		}
		chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
		if chunkSize == 0 {
			chunkSize = DefaultChunkSizeMB
		}
		if chunkSize <= 0 {
			return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
		}

		config := &Config{
			InstanceName:    instanceName,
			ServiceType:     "hidrive_legacy",
			URL:             "https://api.hidrive.strato.com",
			RefreshToken:    refreshToken,
			ClientID:        clientID,
			ClientSecret:    clientSecret,
			TestFileSizeMB:  fileSize,
			TestIntervalSec: interval,
			TestChunkSizeMB: chunkSize,
		}
		configs = append(configs, config)
		l++
	}

	// Dropbox Instanzen (OAuth2 only)
	k := 1
	for {
		refreshTokenKey := fmt.Sprintf("DROPBOX_INSTANCE_%d_REFRESH_TOKEN", k)
		appKeyKey := fmt.Sprintf("DROPBOX_INSTANCE_%d_APP_KEY", k)
		appSecretKey := fmt.Sprintf("DROPBOX_INSTANCE_%d_APP_SECRET", k)
		nameKey := fmt.Sprintf("DROPBOX_INSTANCE_%d_NAME", k)
		
		refreshToken := os.Getenv(refreshTokenKey)
		appKey := os.Getenv(appKeyKey)
		appSecret := os.Getenv(appSecretKey)
		
		if k == 1 && (refreshToken == "" || appKey == "" || appSecret == "") {
			// Check if we need to continue with other service types
			break
		}
		if refreshToken == "" || appKey == "" || appSecret == "" {
			break
		}
		instanceName := os.Getenv(nameKey)
		if instanceName == "" {
			instanceName = fmt.Sprintf("dropbox-instance-%d", k)
		}
		fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
		if fileSize == 0 {
			fileSize = DefaultFileSizeMB
		}
		if fileSize <= 0 {
			return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
		}
		interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
		if interval == 0 {
			interval = DefaultIntervalSec
		}
		if interval <= 0 {
			return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
		}
		chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
		if chunkSize == 0 {
			chunkSize = DefaultChunkSizeMB
		}
		if chunkSize <= 0 {
			return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
		}
		config := &Config{
			InstanceName:    instanceName,
			ServiceType:     "dropbox",
			URL:             "https://api.dropboxapi.com",
			RefreshToken:    refreshToken,
			AppKey:          appKey,
			AppSecret:       appSecret,
			TestFileSizeMB:  fileSize,
			TestIntervalSec: interval,
			TestChunkSizeMB: chunkSize,
		}
		configs = append(configs, config)
		k++
	}

	// MagentaCLOUD Instanzen (WebDAV mit ANID)
	m := 1
	for {
		urlKey := fmt.Sprintf("MAGENTACLOUD_INSTANCE_%d_URL", m)
		userKey := fmt.Sprintf("MAGENTACLOUD_INSTANCE_%d_USER", m)
		passKey := fmt.Sprintf("MAGENTACLOUD_INSTANCE_%d_PASS", m)
		anidKey := fmt.Sprintf("MAGENTACLOUD_INSTANCE_%d_ANID", m)
		
		url := os.Getenv(urlKey)
		user := os.Getenv(userKey)
		pass := os.Getenv(passKey)
		anid := os.Getenv(anidKey)
		
		if m == 1 && (url == "" || user == "" || pass == "" || anid == "") {
			// No MagentaCLOUD instances configured, continue
			break
		}
		if url == "" || user == "" || pass == "" || anid == "" {
			break
		}
		
		fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
		if fileSize == 0 {
			fileSize = DefaultFileSizeMB
		}
		if fileSize <= 0 {
			return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
		}
		interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
		if interval == 0 {
			interval = DefaultIntervalSec
		}
		if interval <= 0 {
			return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
		}
		chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
		if chunkSize == 0 {
			chunkSize = DefaultChunkSizeMB
		}
		if chunkSize <= 0 {
			return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
		}
		
		config := &Config{
			InstanceName:    url,
			ServiceType:     "magentacloud",
			URL:             url,
			Username:        user,
			Password:        pass,
			ANID:            anid,
			TestFileSizeMB:  fileSize,
			TestIntervalSec: interval,
			TestChunkSizeMB: chunkSize,
		}
		configs = append(configs, config)
		m++
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("error: no instances configured. Please set NC_INSTANCE_1_..., HIDRIVE_INSTANCE_1_..., HIDRIVE_LEGACY_INSTANCE_1_..., DROPBOX_INSTANCE_1_REFRESH_TOKEN (with OAuth2 credentials), or MAGENTACLOUD_INSTANCE_1_... (with ANID)")
	}

	// Validate all configurations
	for i, cfg := range configs {
		if err := validateConfig(cfg); err != nil {
			return nil, fmt.Errorf("configuration validation failed for instance %d: %w", i+1, err)
		}
		fmt.Printf("[Config] Validated instance: %s, ServiceType: %s, URL: %s\n", cfg.InstanceName, cfg.ServiceType, cfg.URL)
	}

	// Debug: Log all loaded configs
	for _, c := range configs {
		fmt.Printf("[Config] Loaded instance: %s, ServiceType: %s, URL: %s\n", c.InstanceName, c.ServiceType, c.URL)
	}
	return configs, nil
}

// validateConfig validates a single configuration instance
func validateConfig(cfg *Config) error {
	if cfg.InstanceName == "" {
		return fmt.Errorf("instance name cannot be empty")
	}

	switch cfg.ServiceType {
	case "nextcloud", "hidrive":
		if cfg.URL == "" {
			return fmt.Errorf("URL cannot be empty for %s", cfg.ServiceType)
		}
		if cfg.Username == "" {
			return fmt.Errorf("username cannot be empty for %s", cfg.ServiceType)
		}
		if cfg.Password == "" {
			return fmt.Errorf("password cannot be empty for %s", cfg.ServiceType)
		}
	case "magentacloud":
		if cfg.URL == "" {
			return fmt.Errorf("URL cannot be empty for MagentaCLOUD")
		}
		if cfg.Username == "" {
			return fmt.Errorf("username cannot be empty for MagentaCLOUD")
		}
		if cfg.Password == "" {
			return fmt.Errorf("password cannot be empty for MagentaCLOUD")
		}
		if cfg.ANID == "" {
			return fmt.Errorf("ANID cannot be empty for MagentaCLOUD")
		}
	case "hidrive_legacy":
		if cfg.RefreshToken == "" {
			return fmt.Errorf("refresh token cannot be empty for HiDrive Legacy")
		}
		if cfg.ClientID == "" {
			return fmt.Errorf("client ID cannot be empty for HiDrive Legacy")
		}
		if cfg.ClientSecret == "" {
			return fmt.Errorf("client secret cannot be empty for HiDrive Legacy")
		}
	case "dropbox":
		if cfg.RefreshToken == "" {
			return fmt.Errorf("refresh token cannot be empty for Dropbox")
		}
		if cfg.AppKey == "" {
			return fmt.Errorf("app key cannot be empty for Dropbox")
		}
		if cfg.AppSecret == "" {
			return fmt.Errorf("app secret cannot be empty for Dropbox")
		}
	default:
		return fmt.Errorf("unsupported service type: %s", cfg.ServiceType)
	}

	if cfg.TestFileSizeMB <= 0 {
		return fmt.Errorf("test file size must be positive, got %d", cfg.TestFileSizeMB)
	}
	if cfg.TestIntervalSec <= 0 {
		return fmt.Errorf("test interval must be positive, got %d", cfg.TestIntervalSec)
	}
	if cfg.TestChunkSizeMB <= 0 {
		return fmt.Errorf("chunk size must be positive, got %d", cfg.TestChunkSizeMB)
	}

	return nil
}
