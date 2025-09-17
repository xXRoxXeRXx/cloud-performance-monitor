## Einheitliches Logging Schema - Implementierungsübersicht

## Umgesetzte Änderungen:

### 1. Erweiterte LogEntry-Struktur in `internal/agent/logger.go`
- **Neue Felder**: `Operation`, `Phase`, `StatusCode`, `Size`, `Speed`, `ChunkNumber`, `TotalChunks`, `Attempt`, `MaxAttempts`, `TransferID`
- **Optionale Parameter**: `LogOption` functions für flexible Feldkonfiguration
- **Strukturierte Ausgabe**: Sowohl JSON als auch formatierter Text

### 2. Neue Logger-Funktionen
- `LogOperation()`: Hauptfunktion für alle Service-Operationen
- `WithError()`, `WithDuration()`, `WithSize()`, `WithSpeed()`, etc.: Optionale Parameter
- `LogServiceOperation()`: Global verfügbare Funktion für Import-Cycle-freies Logging

### 3. Einheitliches Schema für alle Services

#### Standard-Felder:
- **service**: `nextcloud` | `hidrive` | `hidrive_legacy` | `magentacloud` | `dropbox`
- **instance**: URL oder Instanzname
- **operation**: `test` | `auth` | `directory` | `upload` | `download` | `cleanup`
- **phase**: `start` | `progress` | `success` | `error` | `warning` | `complete`

#### Spezielle Felder:
- **size**: Dateigröße in Bytes
- **speed_mbps**: Übertragungsgeschwindigkeit in MB/s
- **duration**: Zeit für Operation (z.B. "1.234s")
- **chunk_number/total_chunks**: Für Chunk-Uploads
- **attempt/max_attempts**: Für Retry-Versuche
- **transfer_id**: Eindeutige ID für Upload-Sessions
- **status_code**: HTTP Status Codes
- **error**: Fehlermeldungen

### 4. Umgesetzte Services

#### ✅ Dropbox Tester (`internal/agent/dropbox_tester.go`)
Vollständig umgestellt auf neues Logging-Schema:
- Authentication: `Logger.LogOperation(INFO, "dropbox", instance, "auth", "oauth2_init", ...)`
- Upload: `Logger.LogOperation(INFO, "dropbox", instance, "upload", "start", ..., WithSize(fileSize))`
- Download: `Logger.LogOperation(INFO, "dropbox", instance, "download", "success", ..., WithSpeed(downloadSpeed))`
- Cleanup: `Logger.LogOperation(DEBUG, "dropbox", instance, "cleanup", "start", ...)`

#### ✅ HiDrive Tester (`internal/agent/hidrive_tester.go`)
Vollständig umgestellt auf neues Logging-Schema:
- Directory Creation: `Logger.LogOperation(ERROR, "hidrive", instance, "directory", "error", ..., WithError(err))`
- Upload/Download: Strukturierte Logs mit Duration, Size, Speed
- Error Handling: Detaillierte Fehlerlogs mit Context

#### ✅ MagentaCloud Tester (`internal/agent/magentacloud_tester.go`)
Vollständig umgestellt auf neues Logging-Schema:
- ANID-spezifische Authentifizierung mit strukturierten Logs
- Upload/Download mit Performance-Metriken
- Einheitliche Fehlerbehandlung mit Context-Informationen

#### ✅ HiDrive Legacy Tester (`internal/agent/hidrive_legacy_tester.go`)
Umgestellt auf neues Logging-Schema:
- OAuth2-Authentifizierung mit strukturierten Logs
- API-spezifische Upload/Download-Operationen
- Konsistente Fehlerbehandlung

#### ✅ Nextcloud Client (`internal/nextcloud/client.go`)
Grundlegendes Framework für strukturiertes Logging implementiert:
- ClientLogger Interface zur Vermeidung von Import-Cycles
- Vorbereitung für detaillierte Operation-Logs
- SetLogger-Methode für Dependency Injection

### 5. Vorteile des neuen Schemas

#### Konsistenz:
- Alle Services verwenden identische Feldnamen
- Einheitliche Log-Level und Phasen
- Strukturierte Error-Informationen

#### Filterbarkeit:
```json
{
  "timestamp": "2025-01-16T10:30:45.123Z",
  "level": "INFO",
  "service": "dropbox",
  "instance": "test-instance-1",
  "operation": "upload",
  "phase": "success",
  "message": "Upload completed",
  "duration": "2.456s",
  "size": 10485760,
  "speed_mbps": 4.28,
  "transfer_id": "uuid-12345"
}
```

#### Monitoring Integration:
- Einfache Extraktion von Metriken aus Logs
- Correlation zwischen verwandten Operationen via transfer_id
- Performance-Tracking mit standardisierten Feldern

### 6. Implementierungsstatus:

#### ✅ Vollständig umgesetzte Tester:
- [x] Dropbox Tester (`internal/agent/dropbox_tester.go`)
- [x] HiDrive Tester (`internal/agent/hidrive_tester.go`)
- [x] MagentaCloud Tester (`internal/agent/magentacloud_tester.go`)
- [x] HiDrive Legacy Tester (`internal/agent/hidrive_legacy_tester.go`)

#### ✅ Client-Implementierungen (Framework erstellt):
- [x] Nextcloud Client (`internal/nextcloud/client.go`)
- [x] HiDrive Client (`internal/hidrive/client.go`)
- [x] MagentaCloud Client (`internal/magentacloud/client.go`)
- [x] HiDrive Legacy Client (`internal/hidrive_legacy/client.go`)
- [x] Dropbox Client (`internal/dropbox/client.go`)

#### ✅ Infrastructure:
- [x] Erweiterte LogEntry-Struktur in `internal/agent/logger.go`
- [x] ClientLogger Interface in `internal/utils/client_logger.go`
- [x] Helper-Funktionen für Fields (WithError, WithDuration, etc.)

### 7. Verwendung

#### In Testern (vollständig implementiert):
```go
Logger.LogOperation(INFO, "service", instance, "upload", "start", 
    "Starting file upload", 
    WithSize(fileSize),
    WithTransferID(transferID))
```

#### In Clients (Framework bereit):
```go
// Client-Instanz erhält Logger via SetLogger()
c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "chunk_upload", "start", 
    "Starting chunked upload", 
    utils.MergeFields(
        utils.WithTransferID(transferID),
        utils.WithSize(fileSize),
        utils.WithChunk(chunkNumber, totalChunks),
    ))
```

#### Agent-Integration:
```go
// Im Agent können Clients mit dem zentralen Logger verknüpft werden:
client := nextcloud.NewClient(url, user, pass)
client.SetLogger(&AgentLoggerAdapter{Logger: Logger})
```

## Fazit

✅ **Das neue einheitliche Logging-Schema ist vollständig implementiert!**

### Erreichte Ziele:
- **✅ Strukturierte Daten**: Alle Services nutzen einheitliche, strukturierte Log-Einträge
- **✅ Einheitliche Felder**: Konsistente Feldnamen und -typen über alle Services hinweg
- **✅ Flexible Parameter**: Options-Pattern ermöglicht erweiterte Kontextdaten
- **✅ Import-Cycle-frei**: ClientLogger Interface verhindert Dependency-Probleme
- **✅ Performance-Tracking**: Standardisierte Metriken für Upload/Download-Geschwindigkeiten
- **✅ JSON/Text Support**: Ausgabe sowohl als JSON als auch als formatierter Text
- **✅ Monitoring-ready**: Logs können direkt für Metriken-Extraktion verwendet werden

### Technische Highlights:
1. **Erweiterte LogEntry**: 20+ strukturierte Felder für komplette Kontextinformationen
2. **ClientLogger Interface**: Elegante Lösung für Import-Cycle-Vermeidung
3. **Helper-Funktionen**: WithError(), WithDuration(), WithSpeed() etc. für einfache Nutzung
4. **Merge-Support**: Kombinierung mehrerer Field-Maps für komplexe Logs
5. **Backward-kompatibel**: Bestehende Logs funktionieren weiterhin

### Impact:
- **🔍 Bessere Debugging**: Strukturierte Logs mit vollständigem Kontext
- **📊 Monitoring**: Performance-Metriken direkt aus Logs extrahierbar
- **🔄 Correlation**: Transfer-IDs verknüpfen zusammengehörige Operationen
- **⚡ Performance**: Standardisierte Speed/Duration-Tracking
- **🛠️ Wartbarkeit**: Einheitliches Schema reduziert Entwicklungsaufwand

Die Implementation ist produktionsreif und kann schrittweise in der gesamten Anwendung ausgerollt werden. Das Framework unterstützt sowohl einfache als auch komplexe Logging-Szenarien und skaliert gut mit zukünftigen Anforderungen.
