# Cloud Performance Monitor - Alert Overview

## Alarm-Kategorien und Übersicht

### 🔴 **KRITISCHE ALARME (severity: critical)**

| Alert Name | Beschreibung | Trigger | Dauer | E-Mail Ziel |
|------------|-------------|---------|-------|-------------|
| **ServiceDown** | Monitor Agent ist offline | `up{job="monitor-agent"} == 0` | 1m | Admin, DevOps |
| **CriticalUploadDuration** | Upload dauert über 10 Minuten | `cloud_test_duration_seconds{type="upload"} > 600` | 1m | Admin, DevOps |
| **ServiceTestFailure** | Test-Fehler oder keine erfolgreichen Tests in 20min | `cloud_test_success{error_code!="none"} == 0` oder `absent_over_time(cloud_test_success{error_code="none"}[20m])` | 30s | Admin, DevOps |
| **ServiceUnavailable** | 503/502/500 Server-Fehler | `cloud_test_success{error_code=~"503|502|500"} == 0` | 30s | Admin, DevOps |
| **CriticalErrorRate** | Fehlerrate über 10% in 5min | `rate(cloud_test_errors_total[5m]) > 0.1` | 30s | Admin, DevOps |
| **HTTPServerError** | HTTP 50x Server-Fehler | `cloud_test_success{error_code=~"http_50[0-9]"} == 0` | 30s | Admin, DevOps |
| **CircuitBreakerOpen** | Circuit Breaker ist geöffnet | `nextcloud_circuit_breaker_state > 0` | 0s (sofort) | Admin, DevOps |
| **SLAViolation95Percent** | 24h Erfolgsrate unter 95% | `success_rate < 0.95` | 2m | Admin, Management |

### ⚠️ **WARNUNGEN (severity: warning)**

| Alert Name | Beschreibung | Trigger | Dauer | E-Mail Ziel |
|------------|-------------|---------|-------|-------------|
| **SlowUploadSpeed** | Upload-Geschwindigkeit unter 1 MB/s | `cloud_test_speed_mbytes_per_sec{type="upload"} < 1` | 5m | DevOps |
| **RepeatedServiceFailures** | Mehr als 3 Fehler in 15min | `increase(cloud_test_errors_total[15m]) > 3` | 2m | DevOps |
| **HighErrorRate** | Fehlerrate über 5% in 10min | `rate(cloud_test_errors_total[10m]) > 0.05` | 2m | DevOps |
| **DirectoryCreationFailed** | Verzeichnis-Erstellung fehlgeschlagen | `cloud_test_success{error_code="directory_creation"} == 0` | 1m | DevOps |
| **HTTPConflictError** | HTTP 409 Conflict-Fehler | `cloud_test_success{error_code="http_409_conflict"} == 0` | 1m | DevOps |
| **HTTPPreconditionFailed** | HTTP 412 Precondition Failed | `cloud_test_success{error_code="http_412_precondition_failed"} == 0` | 1m | DevOps |
| **HTTPRateLimitError** | HTTP 429 Rate Limit erreicht | `cloud_test_success{error_code="http_429_rate_limited"} == 0` | 1m | DevOps |
| **HTTPServiceUnavailable** | HTTP 503 Service Unavailable | `cloud_test_success{error_code="http_503_unavailable"} == 0` | 1m | DevOps |
| **RepeatedServiceUnavailableErrors** | Wiederholte 503-Fehler (>2 in 15min) | `increase(cloud_test_errors_total{error_type="http_503_unavailable"}[15m]) > 2` | 5m | DevOps |
| **DownloadIncompleteError** | Download unvollständig | `cloud_test_success{error_code="download_failed"} == 0` | 1m | DevOps |
| **RepeatedConflictErrors** | Wiederholte 409-Fehler (>3 in 20min) | `increase(cloud_test_errors_total{error_type="http_409_conflict"}[20m]) > 3` | 2m | DevOps |
| **ClientTimeoutErrors** | Client-Timeout-Fehler | `cloud_test_success{error_code="network_timeout"} == 0` | 1m | Network Team |
| **RepeatedTimeoutErrors** | Wiederholte Timeout-Fehler (>2 in 20min) | `increase(cloud_test_errors_total{error_type="network_timeout"}[20m]) > 2` | 5m | Network Team |
| **HighNetworkLatency** | Hohe Netzwerk-Latenz | `nextcloud_network_latency_seconds > 2` | 5m | Network Team |
| **ConnectionTimeouts** | Verbindungs-Timeouts | `rate(nextcloud_connection_timeouts_total[10m]) > 0.05` | 3m | Network Team |
| **SlowChunkUploads** | Langsame Chunk-Uploads | `avg_over_time(nextcloud_chunk_upload_duration_seconds[5m]) > 10` | 3m | DevOps |
| **HighChunkRetryRate** | Hohe Chunk-Retry-Rate | `chunk_retry_rate > 15%` | 5m | DevOps |
| **SLAViolation99Percent** | 24h Erfolgsrate unter 99% | `success_rate < 0.99` | 5m | Management |
| **TooManyAlerts** | Zu viele aktive Alarme | `count(ALERTS{alertstate="firing"}) > 5` | 5m | Admin, DevOps |

### 📊 **NETZWERK ALARME (severity: warning)**

| Alert Name | Beschreibung | Trigger | Dauer | E-Mail Ziel |
|------------|-------------|---------|-------|-------------|
| **HighNetworkLatency** | Hohe Netzwerk-Latenz | `nextcloud_network_latency_seconds > 2` | 5m | Network Team |
| **ConnectionTimeouts** | Verbindungs-Timeouts | `rate(nextcloud_connection_timeouts_total[10m]) > 0.05` | 3m | Network Team |
| **ClientTimeoutErrors** | Client-Timeout-Fehler | `cloud_test_success{error_code="network_timeout"} == 0` | 1m | Network Team |
| **RepeatedTimeoutErrors** | Wiederholte Timeout-Fehler (>2 in 20min) | `increase(cloud_test_errors_total{error_type="network_timeout"}[20m]) > 2` | 5m | Network Team |

### 🔧 **PERFORMANCE & RELIABILITY ALARME**

| Alert Name | Beschreibung | Trigger | Dauer | E-Mail Ziel |
|------------|-------------|---------|-------|-------------|
| **SlowChunkUploads** | Langsame Chunk-Uploads | `avg_over_time(nextcloud_chunk_upload_duration_seconds[5m]) > 10` | 3m | DevOps |
| **HighChunkRetryRate** | Hohe Chunk-Retry-Rate | `chunk_retry_rate > 15%` | 5m | DevOps |
| **CircuitBreakerOpen** | Circuit Breaker ist geöffnet | `nextcloud_circuit_breaker_state > 0` | 0s (sofort) | Admin, DevOps |

### 📊 **SLA & META ALARME**

| Alert Name | Beschreibung | Trigger | Dauer | E-Mail Ziel |
|------------|-------------|---------|-------|-------------|
| **SLAViolation99Percent** | 24h Erfolgsrate unter 99% | `success_rate < 0.99` | 5m | Management |
| **SLAViolation95Percent** | 24h Erfolgsrate unter 95% | `success_rate < 0.95` | 2m | Admin, Management |
| **TooManyAlerts** | Zu viele aktive Alarme (>5) | `count(ALERTS{alertstate="firing"}) > 5` | 5m | Admin, DevOps |

## E-Mail Zielgruppen

### 🔴 **Admin + DevOps** (Kritische Alarme)
- `ADMIN_EMAIL` + `DEVOPS_EMAIL`
- Sofortige Reaktion erforderlich
- Service-Ausfälle, kritische Fehler

### ⚠️ **DevOps** (Warnungen)
- `DEVOPS_EMAIL`
- Überwachung und Optimierung
- Performance-Probleme, wiederholte Fehler

### 🌐 **Network Team**
- `NETWORK_EMAIL`
- Netzwerk-spezifische Probleme
- Timeouts, Latenz, Verbindungsprobleme

### 📈 **Management** (SLA-Berichte)
- `MANAGEMENT_EMAIL`
- Monatliche/wöchentliche Berichte
- Trend-Analyse

## Alarm-Schwellwerte

### Performance-Schwellwerte
- **Kritischer Upload**: > 10 Minuten (600s)
- **Langsame Upload-Geschwindigkeit**: < 1 MB/s
- **Hohe Netzwerk-Latenz**: > 2 Sekunden

### Fehlerrate-Schwellwerte
- **Kritische Fehlerrate**: > 10% in 5 Minuten
- **Hohe Fehlerrate**: > 5% in 10 Minuten
- **Connection Timeouts**: > 5% in 10 Minuten

### Wiederholungsschwellwerte
- **Service-Fehler**: > 3 in 15 Minuten
- **503-Fehler**: > 2 in 15 Minuten
- **409-Konflikte**: > 3 in 20 Minuten
- **Timeouts**: > 2 in 20 Minuten

## Entfernte Alarme
- ~~**HighUploadDuration**~~ - Entfernt (war: Upload > 5 Minuten als Warning)

## Alarm-Kategorien nach Severity

- **🔴 Critical (8 Alarme)**: Sofortige Aufmerksamkeit, Service-beeinträchtigend
- **⚠️ Warning (17 Alarme)**: Überwachung erforderlich, potentielle Probleme
- **📊 Info (0 Alarme)**: Informationszwecke, keine Aktion erforderlich

## Alarm-Gruppen

### nextcloud_performance_alerts (25 Alarme)
- Service-Verfügbarkeit, Performance, HTTP-Fehler, Netzwerk-Probleme

### nextcloud_sla_alerts (2 Alarme)
- SLA-Überwachung (99% und 95% Schwellwerte)

### nextcloud_meta_alerts (1 Alarm)
- Meta-Überwachung des Alarm-Systems selbst

## **TOTAL: 28 Alarme**

## Status: ✅ Aktiv | 🔄 In Überprüfung | ❌ Deaktiviert
Alle Alarme sind aktuell **✅ Aktiv** nach den Error-State-Reset-Fixes.