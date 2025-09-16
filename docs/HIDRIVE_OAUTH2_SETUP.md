# 🔑 HiDrive Legacy OAuth2 Setup Guide

Dieser Guide erklärt, wie du die HiDrive Legacy OAuth2 Integration für das Performance Monitoring einrichtest.

## 📋 Überblick

Die **HiDrive Legacy** Integration nutzt die native HiDrive REST API v2.1 mit OAuth2-Authentifizierung. Diese bietet im Vergleich zur WebDAV-Integration optimierte Performance und erweiterte Funktionen.

### ✨ Vorteile der OAuth2 Integration:
- **Bessere Performance**: Optimierte multipart uploads mit automatischer Chunk-Verarbeitung
- **Sichere Authentifizierung**: OAuth2 mit automatischer Token-Erneuerung
- **Erweiterte Funktionen**: Native API-Features wie Pfad-Normalisierung
- **Robustheit**: Automatische Fehlerbehandlung und Retry-Logic

## 🚀 Setup-Schritte

### 1. OAuth2 Application erstellen

1. Gehe zur **HiDrive Developer Console**: https://developer.hidrive.com/
2. Klicke auf **"Create Application"**
3. Fülle die erforderlichen Felder aus:
   - **Application Name**: `Performance Monitor`
   - **Description**: `Monitoring application for performance testing`
   - **Redirect URI**: `http://localhost:8080/callback` (für lokales Setup)

### 2. Client Credentials erhalten

Nach der App-Erstellung erhältst du:
- **Client ID**: z.B. `YOUR_CLIENT_ID_HERE`
- **Client Secret**: z.B. `YOUR_CLIENT_SECRET_HERE`

⚠️ **Wichtig**: Client Secret sicher aufbewahren!

### 3. Refresh Token generieren

#### Option A: Manual Authorization (Empfohlen)

1. Öffne diese URL in deinem Browser (ersetze `CLIENT_ID`):
```
https://api.hidrive.strato.com/2.1/oauth2/authorize?response_type=code&client_id=CLIENT_ID&redirect_uri=http://localhost:8080/callback&scope=rw
```

2. Melde dich mit deinen HiDrive-Zugangsdaten an
3. Kopiere den `code` Parameter aus der Redirect-URL
4. Tausche den Code gegen Refresh Token:

```bash
curl -X POST "https://api.hidrive.strato.com/2.1/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "client_id=CLIENT_ID" \
  -d "client_secret=CLIENT_SECRET" \
  -d "code=AUTHORIZATION_CODE" \
  -d "redirect_uri=http://localhost:8080/callback"
```

#### Option B: Direct Token (Für Tests)

Wenn du bereits einen Access Token hast, nutze ihn direkt:

```bash
curl -X POST "https://api.hidrive.strato.com/2.1/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "client_id=CLIENT_ID" \
  -d "client_secret=CLIENT_SECRET" \
  -d "refresh_token=EXISTING_REFRESH_TOKEN"
```

### 4. Environment Configuration

Füge die Konfiguration zu deiner `.env` Datei hinzu:

```bash
# HiDrive Legacy (OAuth2) Configuration
HIDRIVE_LEGACY_INSTANCE_1_URL=https://api.hidrive.strato.com/2.1
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=YOUR_CLIENT_ID_HERE
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=YOUR_CLIENT_SECRET_HERE
HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN=YOUR_REFRESH_TOKEN_HERE
HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-legacy-main
```

### 5. Testen der Konfiguration

1. Starte den Monitor Agent:
```bash
docker compose up -d monitor-agent
```

2. Überprüfe die Logs:
```bash
docker compose logs monitor-agent | grep "HiDrive Legacy"
```

Erfolgreiche Ausgabe:
```
HiDrive Legacy: OAuth2 access token generated successfully
HiDrive Legacy: Upload completed for hidrive-legacy-main: 7.66 MB/s
HiDrive Legacy: Download completed for hidrive-legacy-main: 18.09 MB/s
```

## 🔧 Konfiguration

### Multiple Instanzen

Du kannst mehrere HiDrive Legacy Instanzen konfigurieren:

```bash
# Erste Instanz
HIDRIVE_LEGACY_INSTANCE_1_URL=https://api.hidrive.strato.com/2.1
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=your-client-id-1
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=your-client-secret-1
HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN=your-refresh-token-1
HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-production

# Zweite Instanz
HIDRIVE_LEGACY_INSTANCE_2_URL=https://api.hidrive.strato.com/2.1
HIDRIVE_LEGACY_INSTANCE_2_CLIENT_ID=your-client-id-2
HIDRIVE_LEGACY_INSTANCE_2_CLIENT_SECRET=your-client-secret-2
HIDRIVE_LEGACY_INSTANCE_2_REFRESH_TOKEN=your-refresh-token-2
HIDRIVE_LEGACY_INSTANCE_2_NAME=hidrive-backup
```

### Erweiterte Konfiguration

Alle Standard-Test-Parameter gelten auch für HiDrive Legacy:

```bash
# Test-Konfiguration
TEST_FILE_SIZE_MB=100                # Testdatei-Größe
TEST_INTERVAL_SECONDS=300            # Test-Intervall
TEST_CHUNK_SIZE_MB=10                # Chunk-Größe für uploads
```

## 📊 Monitoring & Metrics

### Verfügbare Metriken

```prometheus
# Performance Metriken
cloud_test_duration_seconds{service="hidrive_legacy",type="upload|download"}
cloud_test_speed_mbytes_per_sec{service="hidrive_legacy",type="upload|download"}
cloud_test_success{service="hidrive_legacy"}
cloud_test_errors_total{service="hidrive_legacy"}
```

### Grafana Dashboard

Das HiDrive Legacy Service wird automatisch im Grafana Dashboard angezeigt:

1. Öffne Grafana: http://localhost:3003
2. Wähle den **Service Filter** aus
3. Wähle **hidrive_legacy** für spezifische Metriken

## 🐛 Troubleshooting

### Häufige Probleme

**Token ungültig/abgelaufen:**
```bash
# Logs prüfen
docker compose logs monitor-agent | grep "HiDrive Legacy"

# Mögliche Ausgabe
HiDrive Legacy: OAuth2 token refresh failed: invalid_grant
```

**Lösung**: Refresh Token erneuern mit OAuth2 Flow

**Upload-Fehler:**
```bash
# Debugging aktivieren
docker compose logs monitor-agent | grep "DEBUG"
```

**Pfad-Probleme:**
```bash
# Home Directory überprüfen
HiDrive Legacy: User home directory: root/users/myserver
HiDrive Legacy: DEBUG UploadFile - cleanHomePath after TrimPrefix: "/users/myserver"
```

### Debug-Modus

Für detaillierte Logs, nutze Debug-Level:

```bash
# In der Go-Anwendung ist Debug bereits aktiviert
# Logs zeigen vollständige Request/Response Details
docker compose logs -f monitor-agent
```

## 🔐 Security Best Practices

1. **Client Secret sicher aufbewahren**
   - Nutze Environment Variables
   - Niemals in Code committen
   - Regelmäßig rotieren

2. **Refresh Token Protection**
   - Sichere Speicherung
   - Automatische Erneuerung implementiert
   - Monitoring auf Token-Fehler

3. **Access Control**
   - Minimale Scope-Berechtigungen (`rw`)
   - Separate Apps für verschiedene Umgebungen
   - Regelmäßige Audit der App-Berechtigungen

## 📚 API Referenz

- **HiDrive API Dokumentation**: https://api.hidrive.strato.com/2.1/static/apidoc/
- **OAuth2 Endpoints**: https://api.hidrive.strato.com/2.1/oauth2/
- **Developer Console**: https://developer.hidrive.com/

## ✅ Checkliste

- [ ] OAuth2 App in HiDrive Developer Console erstellt
- [ ] Client ID und Secret erhalten
- [ ] Refresh Token via OAuth2 Flow generiert
- [ ] Environment Variables in `.env` konfiguriert
- [ ] Monitor Agent erfolgreich gestartet
- [ ] Logs zeigen erfolgreiche OAuth2-Authentifizierung
- [ ] Performance Tests laufen erfolgreich
- [ ] Metriken in Grafana sichtbar

---

🎯 **Ready to monitor!** Deine HiDrive Legacy OAuth2 Integration ist jetzt bereit für kontinuierliches Performance Monitoring.
