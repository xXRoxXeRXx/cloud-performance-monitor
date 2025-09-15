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

// ServiceConfig defines the configuration pattern for a service type
type ServiceConfig struct {
	ServiceType     string
	Prefix          string
	URLKey          string
	UserKey         string
	PassKey         string
	ANIDKey         string // For MagentaCLOUD
	RefreshTokenKey string // For OAuth2 services
	ClientIDKey     string // For OAuth2 services
	ClientSecretKey string // For OAuth2 services
	AppKeyKey       string // For Dropbox
	AppSecretKey    string // For Dropbox
	NameKey         string // For named instances
	DefaultURL      string // For services with fixed URLs
}

// LoadConfigs loads configurations for all specified Nextcloud, HiDrive, HiDrive Legacy, Dropbox, and MagentaCLOUD instances
func LoadConfigs() ([]*Config, error) {
	var configs []*Config

	// Define service configurations
	serviceConfigs := []ServiceConfig{
		{
			ServiceType: "nextcloud",
			Prefix:      "NC_INSTANCE",
			URLKey:      "URL",
			UserKey:     "USER",
			PassKey:     "PASS",
		},
		{
			ServiceType: "hidrive",
			Prefix:      "HIDRIVE_INSTANCE",
			URLKey:      "URL",
			UserKey:     "USER",
			PassKey:     "PASS",
		},
		{
			ServiceType:     "hidrive_legacy",
			Prefix:          "HIDRIVE_LEGACY_INSTANCE",
			RefreshTokenKey: "REFRESH_TOKEN",
			ClientIDKey:     "CLIENT_ID",
			ClientSecretKey: "CLIENT_SECRET",
			NameKey:         "NAME",
			DefaultURL:      "https://api.hidrive.strato.com",
		},
		{
			ServiceType:     "dropbox",
			Prefix:          "DROPBOX_INSTANCE",
			RefreshTokenKey: "REFRESH_TOKEN",
			AppKeyKey:       "APP_KEY",
			AppSecretKey:    "APP_SECRET",
			NameKey:         "NAME",
			DefaultURL:      "https://api.dropboxapi.com",
		},
		{
			ServiceType: "magentacloud",
			Prefix:      "MAGENTACLOUD_INSTANCE",
			URLKey:      "URL",
			UserKey:     "USER",
			PassKey:     "PASS",
			ANIDKey:     "ANID",
		},
	}

	// Load configurations for each service type
	for _, svc := range serviceConfigs {
		serviceConfigs, err := loadServiceConfigs(svc)
		if err != nil {
			return nil, err
		}
		configs = append(configs, serviceConfigs...)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("error: no instances configured. Please set NC_INSTANCE_1_..., HIDRIVE_INSTANCE_1_..., HIDRIVE_LEGACY_INSTANCE_1_..., DROPBOX_INSTANCE_1_REFRESH_TOKEN (with OAuth2 credentials), or MAGENTACLOUD_INSTANCE_1_... (with ANID)")
	}

	// Validate all configurations
	for i, cfg := range configs {
		if err := validateConfig(cfg); err != nil {
			return nil, fmt.Errorf("configuration validation failed for instance %d: %w", i+1, err)
		}
	}
	return configs, nil
}

// loadServiceConfigs loads configurations for a specific service type
func loadServiceConfigs(svc ServiceConfig) ([]*Config, error) {
	var configs []*Config

	i := 1
	for {
		config, found, err := loadSingleServiceConfig(svc, i)
		if err != nil {
			return nil, err
		}
		if !found {
			break
		}
		configs = append(configs, config)
		i++
	}

	return configs, nil
}

// loadSingleServiceConfig loads a single configuration instance for a service
func loadSingleServiceConfig(svc ServiceConfig, index int) (*Config, bool, error) {
	// Load common test parameters
	fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
	if fileSize == 0 {
		fileSize = DefaultFileSizeMB
	}
	if fileSize <= 0 {
		return nil, false, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
	}

	interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
	if interval == 0 {
		interval = DefaultIntervalSec
	}
	if interval <= 0 {
		return nil, false, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
	}

	chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
	if chunkSize == 0 {
		chunkSize = DefaultChunkSizeMB
	}
	if chunkSize <= 0 {
		return nil, false, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
	}

	// Load service-specific parameters
	switch svc.ServiceType {
	case "nextcloud", "hidrive":
		return loadWebDAVConfig(svc, index, fileSize, interval, chunkSize)
	case "hidrive_legacy":
		return loadHiDriveLegacyConfig(svc, index, fileSize, interval, chunkSize)
	case "dropbox":
		return loadDropboxConfig(svc, index, fileSize, interval, chunkSize)
	case "magentacloud":
		return loadMagentaCloudConfig(svc, index, fileSize, interval, chunkSize)
	default:
		return nil, false, fmt.Errorf("unknown service type: %s", svc.ServiceType)
	}
}

// loadWebDAVConfig loads configuration for WebDAV-based services (Nextcloud, HiDrive)
func loadWebDAVConfig(svc ServiceConfig, index, fileSize, interval, chunkSize int) (*Config, bool, error) {
	urlKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.URLKey)
	userKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.UserKey)
	passKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.PassKey)

	url := os.Getenv(urlKey)
	if index == 1 && url == "" {
		// First instance not configured, continue to next service type
		return nil, false, nil
	}
	if url == "" {
		return nil, false, nil
	}

	user := os.Getenv(userKey)
	pass := os.Getenv(passKey)
	if user == "" || pass == "" {
		return nil, false, fmt.Errorf("error: %s and %s must be set for instance %d", userKey, passKey, index)
	}

	config := &Config{
		InstanceName:    url,
		ServiceType:     svc.ServiceType,
		URL:             url,
		Username:        user,
		Password:        pass,
		TestFileSizeMB:  fileSize,
		TestIntervalSec: interval,
		TestChunkSizeMB: chunkSize,
	}
	return config, true, nil
}

// loadHiDriveLegacyConfig loads configuration for HiDrive Legacy (OAuth2)
func loadHiDriveLegacyConfig(svc ServiceConfig, index, fileSize, interval, chunkSize int) (*Config, bool, error) {
	refreshTokenKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.RefreshTokenKey)
	clientIDKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.ClientIDKey)
	clientSecretKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.ClientSecretKey)
	nameKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.NameKey)

	refreshToken := os.Getenv(refreshTokenKey)
	clientID := os.Getenv(clientIDKey)
	clientSecret := os.Getenv(clientSecretKey)

	if index == 1 && (refreshToken == "" || clientID == "" || clientSecret == "") {
		return nil, false, nil
	}
	if refreshToken == "" || clientID == "" || clientSecret == "" {
		return nil, false, nil
	}

	instanceName := os.Getenv(nameKey)
	if instanceName == "" {
		instanceName = fmt.Sprintf("hidrive-legacy-instance-%d", index)
	}

	config := &Config{
		InstanceName:    instanceName,
		ServiceType:     svc.ServiceType,
		URL:             svc.DefaultURL,
		RefreshToken:    refreshToken,
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		TestFileSizeMB:  fileSize,
		TestIntervalSec: interval,
		TestChunkSizeMB: chunkSize,
	}
	return config, true, nil
}

// loadDropboxConfig loads configuration for Dropbox (OAuth2)
func loadDropboxConfig(svc ServiceConfig, index, fileSize, interval, chunkSize int) (*Config, bool, error) {
	refreshTokenKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.RefreshTokenKey)
	appKeyKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.AppKeyKey)
	appSecretKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.AppSecretKey)
	nameKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.NameKey)

	refreshToken := os.Getenv(refreshTokenKey)
	appKey := os.Getenv(appKeyKey)
	appSecret := os.Getenv(appSecretKey)

	if index == 1 && (refreshToken == "" || appKey == "" || appSecret == "") {
		return nil, false, nil
	}
	if refreshToken == "" || appKey == "" || appSecret == "" {
		return nil, false, nil
	}

	instanceName := os.Getenv(nameKey)
	if instanceName == "" {
		instanceName = fmt.Sprintf("dropbox-instance-%d", index)
	}

	config := &Config{
		InstanceName:    instanceName,
		ServiceType:     svc.ServiceType,
		URL:             svc.DefaultURL,
		RefreshToken:    refreshToken,
		AppKey:          appKey,
		AppSecret:       appSecret,
		TestFileSizeMB:  fileSize,
		TestIntervalSec: interval,
		TestChunkSizeMB: chunkSize,
	}
	return config, true, nil
}

// loadMagentaCloudConfig loads configuration for MagentaCLOUD (WebDAV with ANID)
func loadMagentaCloudConfig(svc ServiceConfig, index, fileSize, interval, chunkSize int) (*Config, bool, error) {
	urlKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.URLKey)
	userKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.UserKey)
	passKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.PassKey)
	anidKey := fmt.Sprintf("%s_%d_%s", svc.Prefix, index, svc.ANIDKey)

	url := os.Getenv(urlKey)
	user := os.Getenv(userKey)
	pass := os.Getenv(passKey)
	anid := os.Getenv(anidKey)

	if index == 1 && (url == "" || user == "" || pass == "" || anid == "") {
		return nil, false, nil
	}
	if url == "" || user == "" || pass == "" || anid == "" {
		return nil, false, nil
	}

	config := &Config{
		InstanceName:    url,
		ServiceType:     svc.ServiceType,
		URL:             url,
		Username:        user,
		Password:        pass,
		ANID:            anid,
		TestFileSizeMB:  fileSize,
		TestIntervalSec: interval,
		TestChunkSizeMB: chunkSize,
	}
	return config, true, nil
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
