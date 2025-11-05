# Umfassendes Code Review - Cloud Performance Monitor
**Review-Datum**: November 3, 2025  
**Version**: 1.0.0  
**Reviewer**: AI Code Analysis

---

## 📋 Executive Summary

Das Cloud Performance Monitor Projekt ist ein **gut strukturiertes, produktionsreifes Monitoring-System** mit professioneller Architektur. Es zeigt starke Engineering-Praktiken mit Graceful Shutdown, strukturiertem Logging, Health Checks und umfassendem Error Handling.

### Gesamtbewertung: **8.5/10** 🌟

**Stärken:**
- ✅ Exzellente Architektur mit klarer Trennung der Verantwortlichkeiten
- ✅ Robustes Error Handling mit Retry-Logik und Circuit Breaker
- ✅ Professionelles Graceful Shutdown Management
- ✅ Umfassende Metriken und Observability
- ✅ Gute Test-Abdeckung mit strukturierten Tests

**Verbesserungspotenzial:**
- 🔧 Kontextmanagement und Request-Timeouts
- 🔧 Interface-Design für bessere Testbarkeit
- 🔧 Configuration Management und Validation
- 🔧 Einige Performance-Optimierungen
- 🔧 Security Hardening

---

## 🏗️ 1. Architektur & Design

### ✅ Gut gemacht

1. **Klare Package-Struktur**
   - Logische Trennung: `internal/agent`, `internal/nextcloud`, `internal/utils`
   - Gute Kapselung mit `internal/` Package
   - Service-spezifische Client-Implementierungen

2. **Dependency Management**
   ```go
   // go.mod - Minimale, aktuelle Dependencies
   go 1.22
   prometheus/client_golang v1.19.0
   google/uuid v1.6.0
   ```

3. **Graceful Shutdown Pattern**
   ```go
   // Exzellente Implementation
   shutdownManager := agent.NewShutdownManager(DefaultShutdownTimeout)
   shutdownManager.AddHook(httpManager.ShutdownHook())
   ```

### 🔧 Verbesserungsvorschläge

#### **KRITISCH: Interface-basiertes Design fehlt**

**Problem**: Clients sind konkrete Typen, keine Interfaces
```go
// Aktuell in cmd/agent/main.go:
clients := make(map[*agent.Config]interface{})  // ❌ interface{} ist type-unsafe
```

**Lösung**: Definiere Cloud Storage Interface
```go
// internal/storage/interface.go - NEU
package storage

import (
    "context"
    "io"
)

// CloudStorage definiert die gemeinsame Schnittstelle für alle Cloud-Provider
type CloudStorage interface {
    // EnsureDirectory erstellt ein Verzeichnis (idempotent)
    EnsureDirectory(ctx context.Context, dirPath string) error
    
    // UploadFile lädt eine Datei hoch mit Chunking-Support
    UploadFile(ctx context.Context, filePath string, reader io.Reader, size int64, chunkSize int64) error
    
    // DownloadFile lädt eine Datei herunter
    DownloadFile(ctx context.Context, filePath string) (io.ReadCloser, error)
    
    // DeleteFile löscht eine Datei oder Verzeichnis
    DeleteFile(ctx context.Context, filePath string) error
    
    // GetServiceInfo liefert Metadaten über den Service
    GetServiceInfo() ServiceInfo
}

type ServiceInfo struct {
    Type         string // "nextcloud", "hidrive", etc.
    InstanceName string
    BaseURL      string
}

// CloudStorageFactory erstellt Cloud Storage Clients
type CloudStorageFactory interface {
    CreateClient(cfg *agent.Config) (CloudStorage, error)
}
```

**Vorteile:**
- ✅ Type-safe client handling
- ✅ Einfacheres Mocking für Tests
- ✅ Pluggable architecture für neue Provider
- ✅ Bessere Code-Wiederverwendung

---

## 🛡️ 2. Error Handling & Robustheit

### ✅ Gut gemacht

1. **Umfassendes Retry-System**
   ```go
   // internal/utils/retry.go - Hervorragende Implementation
   type RetryConfig struct {
       MaxRetries      int
       InitialDelay    time.Duration
       MaxDelay        time.Duration
       BackoffFactor   float64
       RetryableErrors []string
       logger          ClientLogger
   }
   ```

2. **Circuit Breaker Pattern**
   ```go
   // internal/utils/circuit_breaker.go
   func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error
   ```

3. **Detailliertes Error-Code-Mapping**
   ```go
   // internal/agent/error_codes.go
   func ExtractErrorCode(err error, operation string) string
   ```

### 🔧 Verbesserungsvorschläge

#### **KRITISCH: Fehlende Context-Propagation**

**Problem**: Viele Funktionen nutzen keinen Context
```go
// internal/nextcloud/client.go
func (c *Client) EnsureDirectory(dirPath string) error {  // ❌ Kein context
    req, err := c.newRequest("MKCOL", fullPath, nil)
    // ...
    resp, err := httpRetry.DoWithRetryAndLog(req.Context(), ...)  // Context erst hier
}
```

**Lösung**: Context in alle Methoden einbauen
```go
// VORHER (❌)
func (c *Client) EnsureDirectory(dirPath string) error

// NACHHER (✅)
func (c *Client) EnsureDirectory(ctx context.Context, dirPath string) error {
    fullPath := path.Join("/remote.php/dav/files/", c.Username, dirPath)
    
    // Context mit Timeout für diese spezifische Operation
    opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    req, err := c.newRequest("MKCOL", fullPath, nil)
    if err != nil {
        return err
    }
    
    httpRetry := utils.NewHTTPRetryConfig()
    httpRetry.ClientLogger = c.logger
    
    resp, err := httpRetry.DoWithRetryAndLog(opCtx, c.HTTPClient, req, ...)
    // ...
}
```

#### **HOCH: Panic Recovery in Goroutinen fehlt**

**Problem**: Goroutinen ohne Panic Recovery
```go
// cmd/agent/main.go
for _, cfg := range allConfigs {
    wg.Add(1)
    go func(config *agent.Config) {
        defer wg.Done()  // ⚠️ Kein panic recovery
        agent.UpdateNetworkLatencyMetrics(shutdownManager.Context(), config, config.ServiceType)
    }(cfg)
}
```

**Lösung**: Panic Recovery hinzufügen
```go
// internal/utils/goroutine.go - NEU
package utils

import (
    "fmt"
    "runtime/debug"
)

// SafeGo führt eine Goroutine mit Panic Recovery aus
func SafeGo(name string, fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                stack := debug.Stack()
                Logger.ErrorWithFields("panic-recovery", name,
                    fmt.Sprintf("Recovered from panic: %v", r),
                    fmt.Errorf("%v", r),
                    string(stack))
            }
        }()
        fn()
    }()
}

// In cmd/agent/main.go verwenden:
for _, cfg := range allConfigs {
    wg.Add(1)
    config := cfg  // Capture variable
    utils.SafeGo(fmt.Sprintf("network-latency-%s", config.InstanceName), func() {
        defer wg.Done()
        agent.UpdateNetworkLatencyMetrics(shutdownManager.Context(), config, config.ServiceType)
    })
}
```

#### **MITTEL: Resource Leaks bei HTTP-Responses**

**Problem**: Response Bodies nicht immer geschlossen
```go
// internal/nextcloud/client.go - uploadChunks
resp, chunkErr = c.HTTPClient.Do(req)
if chunkErr != nil {
    // ...
    continue  // ❌ Response Body nicht geschlossen bei manchen Error-Pfaden
}
```

**Lösung**: Konsequentes Defer-Pattern
```go
resp, chunkErr = c.HTTPClient.Do(req)
if chunkErr != nil {
    // Bei Netzwerkfehler gibt es keine Response
    continue
}
defer resp.Body.Close()  // ✅ Immer schließen

if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
    break
} else {
    body, _ := io.ReadAll(resp.Body)
    // resp.Body.Close() wird durch defer aufgerufen
    continue
}
```

#### **MITTEL: Error Wrapping verbessern**

**Problem**: Kontextverlust bei Error Wrapping
```go
if err != nil {
    return fmt.Errorf("MKCOL request failed: %w", err)  // ❌ Wenig Kontext
}
```

**Lösung**: Mehr Kontext hinzufügen
```go
if err != nil {
    return fmt.Errorf("failed to create chunk directory at %s for instance %s: %w", 
        chunkDirURL, c.InstanceName, err)
}
```

---

## ⚡ 3. Performance & Skalierbarkeit

### ✅ Gut gemacht

1. **Memory-effizientes Streaming**
   ```go
   // internal/agent/tester.go
   type randomReader struct{}
   func (r *randomReader) Read(p []byte) (n int, err error) {
       return rand.Read(p)
   }
   reader := io.LimitReader(&randomReader{}, fileSize)  // ✅ Kein Memory-Alloc
   ```

2. **HTTP Connection Pooling**
   ```go
   // internal/hidrive/client.go
   t := http.DefaultTransport.(*http.Transport).Clone()
   t.MaxIdleConns = 100
   t.MaxConnsPerHost = 100
   ```

3. **Prometheus Metrics mit Labels**
   ```go
   TestDuration.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Set(...)
   ```

### 🔧 Verbesserungsvorschläge

#### **HOCH: Rate Limiting fehlt**

**Problem**: Keine Rate Limiting für API-Calls
```go
// Alle Instanzen starten gleichzeitig Tests
for _, cfg := range configs {
    startTest(cfg)  // ❌ Kann Provider überlasten
}
```

**Lösung**: Rate Limiter implementieren
```go
// internal/utils/ratelimit.go - NEU
package utils

import (
    "context"
    "golang.org/x/time/rate"
    "time"
)

// RateLimiter kontrolliert die Request-Rate pro Service
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
    }
}

// GetLimiter holt oder erstellt einen Limiter für einen Service
func (rl *RateLimiter) GetLimiter(service string, rps float64) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    if limiter, exists := rl.limiters[service]; exists {
        return limiter
    }
    
    // Erlaube `rps` Requests pro Sekunde mit Burst von 2x
    limiter := rate.NewLimiter(rate.Limit(rps), int(rps*2))
    rl.limiters[service] = limiter
    return limiter
}

// Wait wartet bis ein Request erlaubt ist
func (rl *RateLimiter) Wait(ctx context.Context, service string, rps float64) error {
    limiter := rl.GetLimiter(service, rps)
    return limiter.Wait(ctx)
}

// In agent verwenden:
var globalRateLimiter = utils.NewRateLimiter()

func runTest(cfg *Config) {
    // Rate Limit: Max 1 Request pro 10 Sekunden pro Instanz
    if err := globalRateLimiter.Wait(ctx, cfg.InstanceName, 0.1); err != nil {
        return err
    }
    // Test durchführen...
}
```

**Benötigte Dependency:**
```bash
go get golang.org/x/time/rate
```

#### **MITTEL: Chunk-Upload Parallelisierung**

**Problem**: Chunks werden sequenziell hochgeladen
```go
// internal/nextcloud/client.go - uploadChunks
for {
    bytesRead, readErr := reader.Read(chunk)
    // Upload chunk sequenziell...
}
```

**Lösung**: Worker Pool für parallele Uploads (optional, konfigurierbar)
```go
// internal/nextcloud/client.go
type ChunkUploader struct {
    client      *Client
    concurrency int  // Anzahl paralleler Worker
}

func (c *Client) uploadChunksParallel(chunkDir string, reader io.Reader, chunkSize int64, destinationURL string, concurrency int) error {
    type chunk struct {
        number int
        data   []byte
    }
    
    chunkChan := make(chan chunk, concurrency*2)
    errChan := make(chan error, concurrency)
    
    // Worker Pool
    var wg sync.WaitGroup
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for ch := range chunkChan {
                if err := c.uploadSingleChunk(chunkDir, ch.number, ch.data, destinationURL); err != nil {
                    errChan <- err
                    return
                }
            }
        }()
    }
    
    // Producer: Lese Chunks
    go func() {
        chunkNum := 1
        chunkBuf := make([]byte, chunkSize)
        for {
            n, err := reader.Read(chunkBuf)
            if n > 0 {
                dataCopy := make([]byte, n)
                copy(dataCopy, chunkBuf[:n])
                chunkChan <- chunk{number: chunkNum, data: dataCopy}
                chunkNum++
            }
            if err == io.EOF {
                break
            }
            if err != nil {
                errChan <- err
                break
            }
        }
        close(chunkChan)
    }()
    
    wg.Wait()
    close(errChan)
    
    // Check for errors
    if err := <-errChan; err != nil {
        return err
    }
    return nil
}
```

**Konfiguration in .env:**
```bash
# Anzahl paralleler Chunk-Uploads (Default: 1 = sequenziell)
CHUNK_UPLOAD_CONCURRENCY=3
```

#### **MITTEL: Metrics Cardinality**

**Problem**: Potenzielle hohe Cardinality bei Chunk-Metriken
```go
ChunkRetries.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "chunk_number")
```

**Lösung**: Aggregierte Metriken verwenden
```go
// Statt pro-Chunk Metriken:
ChunkRetries.WithLabelValues(service, instance, chunkNumber)  // ❌ Hohe Cardinality

// Besser:
TotalChunkRetries.WithLabelValues(service, instance).Inc()  // ✅ Niedrige Cardinality
ChunkRetryRate.WithLabelValues(service, instance).Set(retryRate)  // Rate statt absolute Zahlen
```

#### **NIEDRIG: Buffer Pool für Chunks**

**Problem**: Chunk-Buffers werden für jeden Upload neu allokiert
```go
chunk := make([]byte, chunkSize)  // Allokiert bei jedem Test
```

**Lösung**: sync.Pool für Chunk-Buffers
```go
// internal/utils/bufferpool.go - NEU
package utils

import "sync"

var chunkPool = sync.Pool{
    New: func() interface{} {
        b := make([]byte, 10*1024*1024)  // 10MB default
        return &b
    },
}

func GetChunkBuffer() *[]byte {
    return chunkPool.Get().(*[]byte)
}

func PutChunkBuffer(buf *[]byte) {
    chunkPool.Put(buf)
}

// Verwendung:
func (c *Client) uploadChunks(...) error {
    chunkBuf := utils.GetChunkBuffer()
    defer utils.PutChunkBuffer(chunkBuf)
    
    chunk := (*chunkBuf)[:chunkSize]  // Slice auf gewünschte Größe
    // ...
}
```

---

## 🔒 4. Security

### ✅ Gut gemacht

1. **Credentials aus Environment**
   ```go
   url := os.Getenv(urlKey)
   user := os.Getenv(userKey)
   ```

2. **TLS für SMTP**
   ```yaml
   SMTP_REQUIRE_TLS=true
   ```

3. **Docker Security - Minimal Port Exposure**
   ```yaml
   # Nur intern exposed, nicht nach außen
   expose:
     - "8080"
   ```

### 🔧 Verbesserungsvorschläge

#### **HOCH: Secrets Management**

**Problem**: Keine Secrets-Rotation oder externe Secret-Stores
```go
// .env - Credentials in plaintext
NC_INSTANCE_1_PASS=super-secret-password
```

**Lösung**: Docker Secrets oder Vault Integration
```go
// internal/agent/secrets.go - NEU
package agent

import (
    "fmt"
    "os"
    "strings"
)

// SecretProvider lädt Secrets aus verschiedenen Quellen
type SecretProvider interface {
    GetSecret(key string) (string, error)
}

// EnvSecretProvider lädt aus Environment
type EnvSecretProvider struct{}

func (e *EnvSecretProvider) GetSecret(key string) (string, error) {
    val := os.Getenv(key)
    if val == "" {
        return "", fmt.Errorf("secret %s not found", key)
    }
    return val, nil
}

// DockerSecretProvider lädt aus Docker Secrets (/run/secrets/)
type DockerSecretProvider struct {
    basePath string
}

func NewDockerSecretProvider() *DockerSecretProvider {
    return &DockerSecretProvider{basePath: "/run/secrets"}
}

func (d *DockerSecretProvider) GetSecret(key string) (string, error) {
    // Docker Secrets sind unter /run/secrets/<secret-name>
    secretName := strings.ToLower(strings.ReplaceAll(key, "_", "-"))
    data, err := os.ReadFile(fmt.Sprintf("%s/%s", d.basePath, secretName))
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(data)), nil
}

// ChainedSecretProvider versucht mehrere Provider
type ChainedSecretProvider struct {
    providers []SecretProvider
}

func (c *ChainedSecretProvider) GetSecret(key string) (string, error) {
    for _, provider := range c.providers {
        if val, err := provider.GetSecret(key); err == nil {
            return val, nil
        }
    }
    return "", fmt.Errorf("secret %s not found in any provider", key)
}

// Global Secret Provider
var DefaultSecretProvider SecretProvider = &ChainedSecretProvider{
    providers: []SecretProvider{
        NewDockerSecretProvider(),  // Zuerst Docker Secrets probieren
        &EnvSecretProvider{},       // Fallback auf Environment
    },
}

// Verwendung in config.go:
func getPassword(key string) (string, error) {
    return DefaultSecretProvider.GetSecret(key)
}
```

**Docker Secrets Setup:**
```yaml
# docker-compose.yml
services:
  monitor-agent:
    secrets:
      - nc_instance_1_pass
      - smtp_password
    # ...

secrets:
  nc_instance_1_pass:
    file: ./secrets/nc_instance_1_pass.txt
  smtp_password:
    file: ./secrets/smtp_password.txt
```

#### **MITTEL: Input Validation**

**Problem**: Fehlende Validierung von URL-Parametern
```go
// internal/agent/config.go
url := os.Getenv(urlKey)  // ❌ Keine URL-Validierung
```

**Lösung**: URL und Path Validation
```go
// internal/agent/validation.go
package agent

import (
    "fmt"
    "net/url"
    "regexp"
    "strings"
)

// ValidateURL prüft, ob eine URL gültig und sicher ist
func ValidateURL(urlStr string) error {
    if urlStr == "" {
        return fmt.Errorf("URL is empty")
    }
    
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    
    // Nur HTTPS erlauben
    if parsed.Scheme != "https" {
        return fmt.Errorf("only HTTPS URLs are allowed, got: %s", parsed.Scheme)
    }
    
    // Kein localhost in Production
    if strings.Contains(parsed.Host, "localhost") || strings.Contains(parsed.Host, "127.0.0.1") {
        return fmt.Errorf("localhost URLs not allowed in production")
    }
    
    return nil
}

// ValidateFilePath verhindert Path Traversal
func ValidateFilePath(filePath string) error {
    // Verhindere ../ und absolute Pfade
    if strings.Contains(filePath, "..") {
        return fmt.Errorf("path traversal detected: %s", filePath)
    }
    if strings.HasPrefix(filePath, "/") {
        return fmt.Errorf("absolute paths not allowed: %s", filePath)
    }
    
    // Nur alphanumerische Zeichen, -, _ und /
    validPath := regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`)
    if !validPath.MatchString(filePath) {
        return fmt.Errorf("invalid characters in path: %s", filePath)
    }
    
    return nil
}

// In config.go verwenden:
url := os.Getenv(urlKey)
if err := ValidateURL(url); err != nil {
    return nil, false, fmt.Errorf("invalid URL: %w", err)
}
```

#### **MITTEL: Rate Limiting für Health Endpoints**

**Problem**: Health Endpoints könnten für DDoS genutzt werden
```go
mux.HandleFunc("/health", healthChecker.HealthHandler())  // ❌ Kein Rate Limit
```

**Lösung**: Middleware für Rate Limiting
```go
// internal/utils/middleware.go - NEU
package utils

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rps      rate.Limit
    burst    int
}

func NewRateLimitMiddleware(rps float64, burst int) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        limiters: make(map[string]*rate.Limiter),
        rps:      rate.Limit(rps),
        burst:    burst,
    }
}

func (rl *RateLimitMiddleware) getLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rps, rl.burst)
        rl.limiters[ip] = limiter
    }
    return limiter
}

func (rl *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        limiter := rl.getLimiter(ip)
        
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// In main.go verwenden:
healthRateLimit := utils.NewRateLimitMiddleware(10, 20)  // 10 req/sec, burst 20
mux.Handle("/health", healthRateLimit.Handler(http.HandlerFunc(healthChecker.HealthHandler())))
```

---

## 📚 5. Code-Qualität & Wartbarkeit

### ✅ Gut gemacht

1. **Strukturiertes Logging**
   ```go
   Logger.InfoWithFields("component", "instance", "message", "", "")
   ```

2. **Health Checks**
   ```go
   healthChecker := agent.NewHealthChecker(Version)
   healthChecker.RegisterService(cfg.InstanceName)
   ```

3. **Comprehensive Tests**
   - Unit Tests für alle Packages
   - Integration Tests
   - Table-driven Tests

### 🔧 Verbesserungsvorschläge

#### **HOCH: Configuration Validation**

**Problem**: Konfiguration wird erst zur Laufzeit validiert
```go
fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
if fileSize == 0 {
    fileSize = DefaultFileSizeMB
}
// ❌ Erst später wird geprüft ob fileSize > 0
```

**Lösung**: Frühe Validierung mit strukturierter Config
```go
// internal/agent/config.go
type TestConfig struct {
    FileSizeMB   int           `validate:"required,min=1,max=10000"`
    IntervalSec  int           `validate:"required,min=10,max=86400"`
    ChunkSizeMB  int           `validate:"required,min=1,max=1000"`
}

type InstanceConfig struct {
    Name        string `validate:"required,min=1,max=100"`
    ServiceType string `validate:"required,oneof=nextcloud hidrive magentacloud hidrive_legacy dropbox"`
    URL         string `validate:"required,url,https"`
    Username    string `validate:"required_without=RefreshToken"`
    Password    string `validate:"required_without=RefreshToken"`
    // ...
}

func (c *Config) Validate() error {
    // Verwende github.com/go-playground/validator/v10
    validate := validator.New()
    return validate.Struct(c)
}

// Beim Laden validieren:
configs, err := LoadConfigs()
if err != nil {
    return nil, err
}
for i, cfg := range configs {
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("config %d validation failed: %w", i, err)
    }
}
```

**Benötigte Dependency:**
```bash
go get github.com/go-playground/validator/v10
```

#### **MITTEL: Konstanten zentralisieren**

**Problem**: Magic Numbers und Strings verteilt im Code
```go
time.Sleep(10 * time.Second)  // ❌ Was bedeutet 10?
case "nextcloud", "hidrive":  // ❌ Strings dupliziert
```

**Lösung**: Zentrale Constants
```go
// internal/agent/constants.go - NEU
package agent

import "time"

const (
    // Service Types
    ServiceTypeNextcloud      = "nextcloud"
    ServiceTypeHiDrive        = "hidrive"
    ServiceTypeMagentaCloud   = "magentacloud"
    ServiceTypeHiDriveLegacy  = "hidrive_legacy"
    ServiceTypeDropbox        = "dropbox"
    
    // Timeouts
    DefaultHTTPTimeout        = 300 * time.Second
    DefaultMoveTimeout        = 10 * time.Minute
    DefaultShutdownTimeout    = 30 * time.Second
    DefaultHealthCheckTimeout = 10 * time.Second
    
    // Delays
    ChunkUploadDelay          = 1 * time.Second
    MagentaCloudPostUploadDelay = 2 * time.Second
    MagentaCloudMKCOLDelay    = 10 * time.Second
    
    // Retry Settings
    DefaultMaxRetries         = 3
    DefaultInitialDelay       = 1 * time.Second
    DefaultMaxDelay           = 30 * time.Second
    DefaultBackoffFactor      = 2.0
    
    // Performance
    DefaultMaxIdleConns       = 100
    DefaultMaxConnsPerHost    = 100
    
    // Paths
    DefaultTestDirectory      = "performance_tests"
    DefaultSecretsPath        = "/run/secrets"
)

// Verwenden:
switch cfg.ServiceType {
case ServiceTypeNextcloud, ServiceTypeHiDrive:
    // ...
}

time.Sleep(ChunkUploadDelay)
```

#### **MITTEL: Logging-Levels konsistent**

**Problem**: Inkonsistente Verwendung von Log-Levels
```go
log.Printf("ERROR: Upload failed")  // ❌ log.Printf statt Logger.Error
fmt.Printf("[MagentaCloud] Starting MOVE")  // ❌ fmt.Printf statt Logger
```

**Lösung**: Nur strukturiertes Logging verwenden
```go
// Alle log.Printf und fmt.Printf ersetzen durch:
Logger.Info("Upload started", map[string]interface{}{
    "service":  cfg.ServiceType,
    "instance": cfg.InstanceName,
    "size":     fileSize,
})

Logger.Error("Upload failed", err, map[string]interface{}{
    "service":  cfg.ServiceType,
    "instance": cfg.InstanceName,
})
```

#### **NIEDRIG: Test Coverage erhöhen**

**Aktuell**: Gute Basis-Coverage
```bash
# Aktuelle Coverage prüfen:
go test -cover ./...
# internal/utils  53.8%
```

**Ziel**: 80%+ Coverage

**Fehlende Tests:**
```go
// internal/agent/network_monitoring.go - Nicht getestet
func UpdateNetworkLatencyMetrics(ctx context.Context, cfg *Config, service string)

// internal/agent/shutdown.go - Teilweise getestet
func (tm *TestManager) shutdownHook(ctx context.Context) error

// internal/hidrive/client.go - Upload-Chunks-Logik nur teilweise getestet
func (c *Client) uploadChunks(...)
```

**Test-Vorschläge:**
```go
// internal/agent/network_monitoring_test.go - NEU
package agent

import (
    "context"
    "testing"
    "time"
)

func TestUpdateNetworkLatencyMetrics(t *testing.T) {
    tests := []struct {
        name      string
        cfg       *Config
        timeout   time.Duration
        expectErr bool
    }{
        {
            name: "successful latency check",
            cfg: &Config{
                InstanceName: "test-instance",
                URL:          "https://httpbin.org",  // Test-Server
                ServiceType:  "nextcloud",
            },
            timeout:   5 * time.Second,
            expectErr: false,
        },
        {
            name: "timeout on slow server",
            cfg: &Config{
                InstanceName: "slow-instance",
                URL:          "https://httpstat.us/200?sleep=10000",
                ServiceType:  "nextcloud",
            },
            timeout:   1 * time.Second,
            expectErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()
            
            UpdateNetworkLatencyMetrics(ctx, tt.cfg, tt.cfg.ServiceType)
            // Assertions...
        })
    }
}
```

---

## 🔧 6. Docker & Deployment

### ✅ Gut gemacht

1. **Multi-stage Build**
   ```dockerfile
   FROM golang:1.22-alpine AS builder
   FROM alpine:latest
   ```

2. **Health Checks**
   ```yaml
   healthcheck:
     test: ["CMD", "curl", "-f", "http://localhost:8080/health/live"]
   ```

3. **Minimal Port Exposure**
   ```yaml
   expose:  # Nur intern
     - "8080"
   ```

### 🔧 Verbesserungsvorschläge

#### **MITTEL: Image Size Optimierung**

**Aktuell**: Alpine-based (~50MB)

**Optimierung möglich**:
```dockerfile
# Dockerfile - Optimiert
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache Dependencies separat
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build mit Optimierungen
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \  # Strip debug info
    -trimpath \                   # Remove file paths
    -o /agent ./cmd/agent

# Multi-platform support (optional)
FROM --platform=$BUILDPLATFORM alpine:latest

# Nur nötige Tools
RUN apk --no-cache add \
    ca-certificates \
    curl \
    tzdata \
    && rm -rf /var/cache/apk/*

# Non-root user
RUN addgroup -g 1000 monitor && \
    adduser -D -u 1000 -G monitor monitor

WORKDIR /app
COPY --from=builder /agent /app/agent

# Permissions
RUN chown -R monitor:monitor /app

USER monitor

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health/live || exit 1

EXPOSE 8080
CMD ["/app/agent"]
```

**Ergebnis**: ~30MB statt ~50MB, bessere Security durch non-root user

#### **NIEDRIG: Build-Argumente für Versionierung**

```dockerfile
# Dockerfile
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN go build \
    -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /agent ./cmd/agent

# In cmd/agent/main.go:
var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)

func main() {
    Logger.InfoWithFields("monitor-agent", Version, 
        fmt.Sprintf("Build: %s, Commit: %s", BuildTime, GitCommit), "", "")
    // ...
}

# Build:
docker build \
    --build-arg VERSION=1.0.0 \
    --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
    --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
    -t monitor-agent:1.0.0 .
```

---

## 📊 7. Monitoring & Observability

### ✅ Gut gemacht

1. **Umfassende Metriken**
   - Test Duration, Speed, Success
   - Chunk-Statistiken
   - Network Latency
   - Circuit Breaker State

2. **Prometheus Alerts**
   ```yaml
   # prometheus/alert_rules.yml
   - alert: ServiceDown
   - alert: CriticalUploadDuration
   ```

3. **Grafana Dashboards**
   - Enhanced Dashboard mit 4 Bereichen
   - Service Selector

### 🔧 Verbesserungsvorschläge

#### **MITTEL: Distributed Tracing**

**Problem**: Keine End-to-End Request-Tracing
```go
// Schwierig zu debuggen welcher Chunk bei welchem Test fehlschlägt
```

**Lösung**: OpenTelemetry Integration
```go
// internal/utils/tracing.go - NEU
package utils

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("cloud-performance-monitor")

// StartSpan startet einen neuen Trace-Span
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
    return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// In Nextcloud Client verwenden:
func (c *Client) UploadFile(ctx context.Context, filePath string, ...) error {
    ctx, span := utils.StartSpan(ctx, "nextcloud.UploadFile",
        attribute.String("file.path", filePath),
        attribute.Int64("file.size", size),
        attribute.String("instance", c.InstanceName),
    )
    defer span.End()
    
    // Chunks upload
    ctx, chunkSpan := utils.StartSpan(ctx, "nextcloud.uploadChunks")
    defer chunkSpan.End()
    
    // MOVE operation
    ctx, moveSpan := utils.StartSpan(ctx, "nextcloud.MOVE")
    defer moveSpan.End()
    
    // ...
}
```

**Setup:**
```go
// cmd/agent/main.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(os.Getenv("JAEGER_ENDPOINT")),
    ))
    if err != nil {
        return nil, err
    }
    
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("cloud-performance-monitor"),
            semconv.ServiceVersionKey.String(Version),
        )),
    )
    
    otel.SetTracerProvider(tp)
    return tp, nil
}

func main() {
    tp, err := initTracer()
    if err != nil {
        Logger.Warn("Failed to initialize tracing", err)
    } else {
        defer tp.Shutdown(context.Background())
    }
    // ...
}
```

**Docker Compose:**
```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # Collector
    environment:
      - COLLECTOR_ZIPKIN_HTTP_PORT=9411
    networks:
      - monitor-net

  monitor-agent:
    # ...
    environment:
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

#### **NIEDRIG: Custom Metrics**

**Vorschlag**: Zusätzliche Business-Metriken
```go
// internal/agent/metrics.go
var (
    // Cost-Tracking (falls Provider Kosten pro API-Call haben)
    EstimatedCosts = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloud_estimated_costs_usd",
            Help: "Estimated costs based on API usage",
        },
        []string{"service", "instance"},
    )
    
    // Data Transfer
    TotalDataTransferred = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cloud_data_transferred_bytes_total",
            Help: "Total bytes transferred (upload + download)",
        },
        []string{"service", "instance", "direction"},  // direction: upload/download
    )
    
    // SLA Compliance
    SLACompliance = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloud_sla_compliance_percent",
            Help: "SLA compliance percentage (rolling 24h)",
        },
        []string{"service", "instance"},
    )
)
```

---

## 🎯 8. Prioritäten-Matrix

### 🔴 **KRITISCH** (Sofort umsetzen)

1. **Context-Propagation in alle Client-Methoden**
   - Betrifft: `internal/nextcloud`, `internal/hidrive`, `internal/magentacloud`, `internal/hidrive_legacy`, `internal/dropbox`
   - Aufwand: 2-3 Tage
   - Impact: Hoch (Timeout-Control, Graceful Cancellation)

2. **Interface-basiertes Design für CloudStorage**
   - Betrifft: `internal/storage` (neu), `cmd/agent/main.go`
   - Aufwand: 1-2 Tage
   - Impact: Hoch (Testbarkeit, Erweiterbarkeit)

3. **Panic Recovery in allen Goroutinen**
   - Betrifft: `cmd/agent/main.go`, `internal/utils` (neu: goroutine.go)
   - Aufwand: 1 Tag
   - Impact: Hoch (Stabilität)

### 🟡 **HOCH** (Nächste Sprint)

4. **Rate Limiting implementieren**
   - Betrifft: `internal/utils` (neu: ratelimit.go), Client-Aufrufe
   - Aufwand: 2 Tage
   - Impact: Mittel-Hoch (Provider-Schutz)

5. **Secrets Management (Docker Secrets)**
   - Betrifft: `internal/agent/secrets.go` (neu), `docker-compose.yml`
   - Aufwand: 1-2 Tage
   - Impact: Hoch (Security)

6. **Configuration Validation**
   - Betrifft: `internal/agent/config.go`, `internal/agent/validation.go` (neu)
   - Aufwand: 1 Tag
   - Impact: Mittel (Fehlerprävention)

### 🟢 **MITTEL** (Backlog)

7. **Response Body Leak-Fixes**
   - Betrifft: Alle Client-Packages
   - Aufwand: 1 Tag
   - Impact: Mittel (Resource Management)

8. **Chunk-Upload Parallelisierung**
   - Betrifft: Client UploadFile-Methoden
   - Aufwand: 2-3 Tage
   - Impact: Mittel (Performance)

9. **Distributed Tracing**
   - Betrifft: Übergreifend, neue Infrastruktur
   - Aufwand: 3-4 Tage
   - Impact: Mittel (Debugging)

10. **Docker Image Optimierung**
    - Betrifft: `Dockerfile`
    - Aufwand: 0.5 Tage
    - Impact: Niedrig-Mittel (Size, Security)

### ⚪ **NIEDRIG** (Nice-to-have)

11. **Buffer Pool für Chunks**
    - Aufwand: 0.5 Tage
    - Impact: Niedrig (Micro-Optimierung)

12. **Test Coverage auf 80%+**
    - Aufwand: 3-4 Tage
    - Impact: Niedrig-Mittel (Qualitätssicherung)

13. **Konstanten zentralisieren**
    - Aufwand: 1 Tag
    - Impact: Niedrig (Code-Qualität)

---

## 📈 9. Implementierungsplan (4 Sprints)

### **Sprint 1 (Woche 1-2): Kritische Fixes**
- ✅ Context-Propagation
- ✅ Interface-Design
- ✅ Panic Recovery
- **Ziel**: Robustheit & Testbarkeit

### **Sprint 2 (Woche 3-4): Security & Configuration**
- ✅ Rate Limiting
- ✅ Secrets Management
- ✅ Configuration Validation
- ✅ Input Validation
- **Ziel**: Production-Ready Security

### **Sprint 3 (Woche 5-6): Performance & Observability**
- ✅ Resource Leak Fixes
- ✅ Chunk Parallelisierung (optional)
- ✅ Distributed Tracing
- ✅ Custom Metrics
- **Ziel**: Skalierbarkeit

### **Sprint 4 (Woche 7-8): Polishing**
- ✅ Docker Optimierung
- ✅ Test Coverage
- ✅ Code-Qualität
- ✅ Dokumentation
- **Ziel**: Code Excellence

---

## 🎓 10. Best Practices für neue Features

### **Code-Review Checklist**

Für jedes neue Feature prüfen:

1. **Context Handling**
   ```go
   ✅ func DoSomething(ctx context.Context, ...) error
   ❌ func DoSomething(...) error
   ```

2. **Error Wrapping**
   ```go
   ✅ return fmt.Errorf("operation failed for %s: %w", resource, err)
   ❌ return err
   ```

3. **Resource Cleanup**
   ```go
   ✅ resp, err := client.Do(req)
      if err != nil { return err }
      defer resp.Body.Close()
   ```

4. **Logging**
   ```go
   ✅ Logger.InfoWithFields("component", "instance", "message", "", "")
   ❌ fmt.Printf("message")
   ```

5. **Testing**
   ```go
   ✅ Mindestens Unit-Test für jeden neuen Code-Pfad
   ✅ Table-driven Tests für Varianten
   ✅ Mock für externe Dependencies
   ```

6. **Metrics**
   ```go
   ✅ Neue Operationen als Metriken erfassen
   ✅ Niedrige Cardinality (keine IDs in Labels)
   ```

---

## 📝 Zusammenfassung

### **Stärken** 👍
- Solide Architektur mit klarer Struktur
- Exzellentes Shutdown-Management
- Gutes Error-Handling mit Retry/Circuit Breaker
- Umfassende Observability
- Docker-ready mit Health Checks

### **Quick Wins** 🚀 (< 1 Tag Aufwand)
1. Panic Recovery hinzufügen
2. Docker non-root user
3. Konstanten zentralisieren
4. Resource Leak Fixes
5. URL-Validierung

### **High Impact Improvements** 💎 (1-3 Tage Aufwand)
1. Context-Propagation
2. Interface-Design
3. Secrets Management
4. Rate Limiting
5. Configuration Validation

### **Gesamtaufwand für alle Verbesserungen**
- Kritisch: ~5 Tage
- Hoch: ~7 Tage
- Mittel: ~10 Tage
- Niedrig: ~5 Tage
- **Total**: ~27 Tage (ca. 5-6 Wochen für 1 Entwickler)

---

## 🔗 Nächste Schritte

1. **Review dieses Dokuments** mit dem Team
2. **Priorisieren** basierend auf Business-Anforderungen
3. **Issues anlegen** für priorisierte Items
4. **Sprint Planning** für die nächsten 2 Monate
5. **Implementation** starten mit kritischen Fixes

---

**Ende des Code Reviews**  
*Für Fragen oder Diskussion der Vorschläge stehe ich zur Verfügung.*
