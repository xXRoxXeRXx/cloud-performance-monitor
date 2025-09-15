# HiDrive Legacy HTTP API Integration

Diese Dokumentation erklärt die Integration des HiDrive Legacy HTTP REST API für das Multi-Cloud Performance Monitoring System.

## Überblick

HiDrive Legacy nutzt die HTTP REST API von HiDrive anstelle von WebDAV. Dies bietet folgende Vorteile:
- OAuth2-Authentifizierung
- Multipart-Upload für große Dateien  
- Chunked Upload Support
- REST API Endpoints für alle Operationen

## OAuth2 Setup

### 1. HiDrive Developer App erstellen

1. Besuche die [HiDrive Developer Console](https://developer.hidrive.com/)
2. Erstelle eine neue OAuth2 Application
3. Notiere dir:
   - **Client ID** (`client_id`)
   - **Client Secret** (`client_secret`)
   - **Redirect URI** (für OAuth2 Flow)

### 2. Access Token erhalten

#### Option A: OAuth2 Authorization Code Flow

```bash
# 1. Benutzer-Autorisierung (im Browser öffnen)
https://api.hidrive.strato.com/oauth2/authorize?client_id=YOUR_CLIENT_ID&response_type=code&redirect_uri=YOUR_REDIRECT_URI&scope=rw

# 2. Authorization Code zu Access Token tauschen
curl -X POST https://api.hidrive.strato.com/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&code=AUTHORIZATION_CODE&redirect_uri=YOUR_REDIRECT_URI"
```

#### Option B: Client Credentials Flow (wenn verfügbar)

```bash
curl -X POST https://api.hidrive.strato.com/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&scope=rw"
```

### 3. Konfiguration

Füge die HiDrive Legacy Konfiguration zu deiner `.env` Datei hinzu:

```bash
# HiDrive Legacy HTTP API
HIDRIVE_LEGACY_INSTANCE_1_TOKEN=your-access-token-here
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=your-client-id-here          # Optional
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=your-client-secret-here  # Optional  
HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-legacy-main              # Optional

# Weitere Instanzen
HIDRIVE_LEGACY_INSTANCE_2_TOKEN=another-access-token
HIDRIVE_LEGACY_INSTANCE_2_NAME=hidrive-legacy-backup
```

## API Endpoints

Das System nutzt folgende HiDrive API Endpoints:

| Operation | Method | Endpoint | Beschreibung |
|-----------|--------|----------|-------------|
| Upload (Single) | POST | `/2.1/file` | Einfacher Datei-Upload mit multipart/form-data |
| Upload (Empty File) | POST | `/2.1/file` | Leere Datei für Chunked Upload erstellen |
| Upload (Chunk) | PATCH | `/2.1/file?path=X&offset=Y` | Chunk an spezifische Position schreiben |
| Download | GET | `/2.1/file/data?path=X` | Datei herunterladen |
| Delete | DELETE | `/2.1/file?path=X` | Datei löschen |
| Directory Check | GET | `/2.1/dir?path=X` | Verzeichnis prüfen |
| Directory Create | POST | `/2.1/dir` | Verzeichnis erstellen |
| Connection Test | GET | `/2.1/app/me?fields=id,name` | Verbindung und Token testen |

## Upload-Strategien

### Small Files (≤ 50MB)
- Direkter POST Upload via `/file` mit multipart/form-data
- Vollständiger Datei-Upload in einem Request

### Large Files (> 50MB) - Chunked Upload
Das HiDrive API verwendet ein **zweistufiges Chunked Upload-Verfahren**:

1. **POST /file** - Erstellt eine leere Datei mit multipart/form-data
2. **PATCH /file?offset=X** - Schreibt Chunks sequenziell mit Offset

#### Technischer Ablauf:

```bash
# Schritt 1: Leere Datei erstellen
POST /2.1/file
Content-Type: multipart/form-data
Body:
- file: [leer]
- path: /performance_tests/testfile.tmp

# Schritt 2: Chunks mit PATCH hochladen
PATCH /2.1/file?path=/performance_tests/testfile.tmp&offset=0
Content-Type: application/octet-stream
Body: [Chunk 1 Daten - 50MB]

PATCH /2.1/file?path=/performance_tests/testfile.tmp&offset=52428800
Content-Type: application/octet-stream  
Body: [Chunk 2 Daten - 50MB]
```

#### Vorteile:
- **Echtes Streaming**: Keine komplette Datei im Speicher
- **Robust**: Offset-basierte Chunk-Reihenfolge
- **Effizient**: Direktes Schreiben an spezifische Dateipositionen
- **Standard-konform**: Nutzt HTTP PATCH für partielle Updates

## Performance Monitoring

### Metrics

Das System erfasst folgende Metriken für HiDrive Legacy:

```prometheus
# Test-Dauer
hidrive_legacy_test_duration_seconds{service="hidrive_legacy",instance="hidrive-legacy-main",type="upload",status="success"}

# Upload/Download-Geschwindigkeit  
hidrive_legacy_test_speed_mbytes_per_sec{service="hidrive_legacy",instance="hidrive-legacy-main",type="upload",status="success"}

# Standard Nextcloud-Metriken (wiederverwendet)
nextcloud_test_success{service="hidrive_legacy",instance="hidrive-legacy-main",type="upload",error_code="none"}
nextcloud_test_errors_total{service="hidrive_legacy",instance="hidrive-legacy-main",type="upload",error_type="none"}
```

### Grafana Dashboard

Das bestehende Grafana Dashboard unterstützt HiDrive Legacy über den Service-Selector:
- Service-Filter: `hidrive_legacy`
- Alle Standard-Panels funktionieren
- Upload/Download-Geschwindigkeiten
- Fehlerrate und Latenz-Metriken

## Troubleshooting

### Häufige Probleme

1. **"Invalid access token"**
   - Token ist abgelaufen → Neuen Token generieren
   - Token hat falsche Scopes → `rw` Scope benötigt

2. **"Upload failed with 413"**
   - Datei zu groß für Single Upload → Chunked Upload wird automatisch verwendet
   - Chunk-Größe anpassen: `TEST_CHUNK_SIZE_MB=2`

3. **"Directory not found"**  
   - Performance-Test-Verzeichnis wird automatisch erstellt
   - Prüfe Berechtigungen des Access Tokens

### Debug-Logs

```bash
# Container-Logs anzeigen
docker logs monitor-agent

# Spezifische HiDrive Legacy Logs filtern
docker logs monitor-agent 2>&1 | grep "HiDrive Legacy"

# Beispiel-Ausgabe bei Chunked Upload:
# 2025/09/14 15:30:00 [HiDrive Legacy] Starting chunked upload for /performance_tests/testfile.tmp (size: 104857600 bytes, chunk size: 52428800)
# 2025/09/14 15:30:01 [HiDrive Legacy] Empty file created: /performance_tests/testfile.tmp  
# 2025/09/14 15:30:03 [HiDrive Legacy] Uploaded chunk 1/2 (offset: 0, size: 52428800 bytes)
# 2025/09/14 15:30:05 [HiDrive Legacy] Uploaded chunk 2/2 (offset: 52428800, size: 52428800 bytes)
# 2025/09/14 15:30:05 [HiDrive Legacy] Chunked upload completed for /performance_tests/testfile.tmp (2 chunks)
```

### API-Test

Teste die HiDrive API direkt:

```bash
# User-Info abrufen
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  https://api.hidrive.strato.com/2.1/app/me?fields=id,name

# Datei-Upload testen
curl -X POST https://api.hidrive.strato.com/2.1/file \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@testfile.txt" \
  -F "path=/testfile.txt"

# Chunk-Upload testen (PATCH)
curl -X PATCH "https://api.hidrive.strato.com/2.1/file?path=/testfile.txt&offset=0" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @chunk1.bin
```

### API-Limits und Beschränkungen

- **Max. Datei-Größe**: Abhängig von HiDrive-Plan
- **Max. Chunk-Größe**: 2GB (PATCH-Limit)
- **Rate Limiting**: API-spezifische Limits beachten
- **Concurrent Uploads**: Sequenzieller Upload empfohlen

## Integration Details

### Code-Struktur

```
internal/
├── agent/
│   ├── config.go              # HiDrive Legacy Konfiguration
│   ├── hidrive_legacy_tester.go  # Performance-Tests
│   └── metrics.go             # Metriken-Definition
├── hidrive_legacy/
│   └── client.go              # HTTP API Client
cmd/agent/main.go              # Service-Integration
```

### Client-Implementation

- **OAuth2**: Access Token im Authorization Header
- **Streaming Upload**: Chunks werden direkt vom Reader gelesen (kein Memory Overflow)
- **Offset-basiertes Chunking**: PATCH mit präzisem Offset für jeden Chunk
- **Error Handling**: Robuste Fehlerbehandlung pro Chunk und Operation
- **HTTP-Compliance**: Korrekte Content-Type Header für multipart und octet-stream
- **Progress Tracking**: Detailliertes Logging für Upload-Fortschritt

### Technische Besonderheiten

1. **Zwei-Phasen Upload**:
   - Phase 1: POST erstellt leere Datei
   - Phase 2: PATCH schreibt Chunks sequenziell

2. **Content-Type Handling**:
   - POST: `multipart/form-data`
   - PATCH: `application/octet-stream`

3. **Chunk-Streaming**:
   ```go
   // Kein io.ReadAll() - direktes Streaming
   chunkData := make([]byte, chunkSize)
   bytesRead, err := io.ReadFull(reader, chunkData)
   ```

4. **Offset-Berechnung**:
   ```go
   offset := chunkNumber * chunkSize
   // PATCH /file?path=X&offset={offset}
   ```

## API-Referenz

Vollständige Dokumentation: https://api.hidrive.strato.com/2.1/static/apidoc/

### Wichtige Endpunkte für Performance-Tests:

- **POST /2.1/file**: Datei-Upload (single/empty)
- **PATCH /2.1/file**: Chunk-Upload mit Offset
- **GET /2.1/file/data**: Datei-Download
- **DELETE /2.1/file**: Datei-Löschung
- **GET /2.1/app/me**: Verbindungstest
- **Chunked Upload**: Binäre Chunks mit Content-Range Headers
- **Error Handling**: HTTP Status Codes + JSON Error Response
- **Retry Logic**: Exponential Backoff für fehlgeschlagene Requests

## Migration von WebDAV

Falls du von HiDrive WebDAV zu HiDrive Legacy HTTP API wechseln möchtest:

1. **Parallel betreiben**: Beide Services können gleichzeitig laufen
2. **Konfiguration kopieren**: 
   ```bash
   # Bestehend (WebDAV)
   HIDRIVE_INSTANCE_1_URL=https://webdav.hidrive.com
   HIDRIVE_INSTANCE_1_USER=username
   HIDRIVE_INSTANCE_1_PASS=password
   
   # Neu (HTTP API)
   HIDRIVE_LEGACY_INSTANCE_1_TOKEN=oauth2-access-token
   HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-legacy-main
   ```
3. **Monitoring vergleichen**: Beide Services in Grafana Dashboard überwachen
4. **Migration**: WebDAV Konfiguration entfernen wenn Legacy-Version stabil läuft

## Weitere Ressourcen

- [HiDrive API Dokumentation](https://developer.hidrive.com/http-api-reference/)
- [OAuth2 Setup Guide](https://developer.hidrive.com/get-started/)
- [HiDrive Developer Console](https://developer.hidrive.com/)
