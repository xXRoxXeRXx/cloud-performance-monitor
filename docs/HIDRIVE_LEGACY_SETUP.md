# HiDrive Legacy HTTP REST API Setup

Diese Anleitung erklärt, wie du die HiDrive Legacy HTTP REST API in den Nextcloud Performance Monitor integrierst.

## Überblick

Die HiDrive Legacy Integration nutzt die native HiDrive HTTP REST API anstelle von WebDAV. Dies bietet folgende Vorteile:

- **Bessere Performance**: Native API ist oft schneller als WebDAV
- **Mehr Features**: Zugriff auf HiDrive-spezifische Funktionen
- **OAuth2 mit Refresh Tokens**: Sichere, langlebige Authentifizierung
- **Automatische Token-Erneuerung**: Tokens werden automatisch erneuert
- **Detaillierte Fehlerbehandlung**: Spezifische HiDrive API Error Codes

## Voraussetzungen

1. **HiDrive Account**: Aktiver HiDrive Account bei STRATO
2. **Developer Access**: Zugang zur HiDrive Developer API
3. **OAuth2 App**: Registrierte OAuth2-Anwendung

## Schritt 1: HiDrive Developer App erstellen

### 1.1 HiDrive Developer Portal
1. Gehe zu: https://developer.hidrive.com/
2. Melde dich mit deinem HiDrive Account an
3. Navigiere zu "Apps" → "Create App"

### 1.2 App-Konfiguration
```
App Name: Nextcloud Performance Monitor
App Type: Web Application
Redirect URI: http://localhost:8080/callback (für lokale Tests)
Scope: owner,rw
```

### 1.3 Credentials erhalten
Nach der App-Erstellung erhältst du:
- **Client ID**: z.B. `your-app-client-id`
- **Client Secret**: z.B. `your-app-client-secret`

## Schritt 2: OAuth2 Refresh Token generieren

### Option A: PowerShell Script (Windows)

```powershell
# HiDrive OAuth2 Token Exchange mit Refresh Token
$clientId = "your-app-client-id"
$clientSecret = "your-app-client-secret"

# Schritt 1: Authorization URL generieren (wichtig: offline_access für Refresh Token)
$authUrl = "https://www.hidrive.strato.com/oauth2/authorize?client_id=$clientId&redirect_uri=http://localhost:8080/callback&response_type=code&scope=owner,rw&access_type=offline"
Write-Host "Öffne diese URL im Browser:" -ForegroundColor Green
Write-Host $authUrl

# Schritt 2: Authorization Code aus Callback URL extrahieren
$authCode = Read-Host "Gib den Authorization Code aus der Callback URL ein"

# Schritt 3: Refresh Token tauschen
$tokenUrl = "https://www.hidrive.strato.com/oauth2/token"
$body = @{
    grant_type = "authorization_code"
    client_id = $clientId
    client_secret = $clientSecret
    redirect_uri = "http://localhost:8080/callback"
    code = $authCode
}

try {
    $response = Invoke-RestMethod -Uri $tokenUrl -Method POST -Body $body -ContentType "application/x-www-form-urlencoded"
    
    Write-Host "SUCCESS! OAuth2 Tokens erhalten:" -ForegroundColor Green
    Write-Host "Access Token: $($response.access_token)" -ForegroundColor Yellow
    Write-Host "Token Type: $($response.token_type)"
    Write-Host "Expires in: $($response.expires_in) seconds"
    
    if ($response.refresh_token) {
        Write-Host "Refresh Token: $($response.refresh_token)" -ForegroundColor Cyan
        Write-Host "WICHTIG: Der Refresh Token ist langlebig und sollte für die Konfiguration verwendet werden!" -ForegroundColor Green
    } else {
        Write-Host "WARNUNG: Kein Refresh Token erhalten. Stelle sicher, dass access_type=offline gesetzt ist." -ForegroundColor Red
    }
} catch {
    Write-Host "ERROR: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}
```

### Option B: curl (Linux/macOS)

```bash
# 1. Authorization URL öffnen
CLIENT_ID="your-app-client-id"
CLIENT_SECRET="your-app-client-secret"

echo "Öffne diese URL im Browser:"
echo "https://www.hidrive.strato.com/oauth2/authorize?client_id=$CLIENT_ID&redirect_uri=http://localhost:8080/callback&response_type=code&scope=owner,rw"

# 2. Authorization Code eingeben
read -p "Authorization Code aus Callback URL: " AUTH_CODE

# 3. Access Token tauschen
curl -X POST "https://www.hidrive.strato.com/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "redirect_uri=http://localhost:8080/callback" \
  -d "code=$AUTH_CODE"
```

### Option C: Python Script

```python
import requests
import urllib.parse

# Konfiguration
client_id = "your-app-client-id"
client_secret = "your-app-client-secret"
redirect_uri = "http://localhost:8080/callback"

# 1. Authorization URL
auth_url = f"https://www.hidrive.strato.com/oauth2/authorize?client_id={client_id}&redirect_uri={redirect_uri}&response_type=code&scope=owner,rw"
print(f"Öffne diese URL im Browser:\n{auth_url}")

# 2. Authorization Code
auth_code = input("Authorization Code aus Callback URL: ")

# 3. Token Exchange
token_url = "https://www.hidrive.strato.com/oauth2/token"
data = {
    'grant_type': 'authorization_code',
    'client_id': client_id,
    'client_secret': client_secret,
    'redirect_uri': redirect_uri,
    'code': auth_code
}

response = requests.post(token_url, data=data)
if response.status_code == 200:
    token_data = response.json()
    print(f"SUCCESS! Access Token: {token_data['access_token']}")
    if 'refresh_token' in token_data:
        print(f"Refresh Token: {token_data['refresh_token']}")
else:
    print(f"ERROR: {response.status_code} - {response.text}")
```

## Schritt 3: Konfiguration

### 3.1 .env Datei erstellen
Kopiere die `.env.hidrive_legacy.example` und benenne sie zu `.env`:

```bash
cp .env.hidrive_legacy.example .env
```

### 3.2 Credentials eintragen
```env
# HiDrive Legacy Instanz mit Refresh Token (empfohlen)
HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN=your-generated-refresh-token
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=your-app-client-id
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=your-app-client-secret
HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-legacy-main
```

### 3.3 Test-Parameter (optional)
```env
TEST_FILE_SIZE_MB=10          # Testdatei-Größe
TEST_INTERVAL_SECONDS=300     # Test-Intervall (5 Minuten)
TEST_CHUNK_SIZE_MB=10         # Chunk-Größe für Uploads
```

## Schritt 4: System starten

### 4.1 Docker Compose
```bash
# System komplett starten
docker compose up -d

# Nur Monitor-Agent starten (für Tests)
docker compose up monitor-agent -d

# Logs verfolgen
docker logs monitor-agent -f
```

### 4.2 Erfolg prüfen
Erwartete Log-Ausgaben:
```
[Config] Validated instance: hidrive-legacy-main, ServiceType: hidrive_legacy, URL: https://api.hidrive.strato.com
[HiDrive Legacy] Using OAuth2 client with refresh token for hidrive-legacy-main
[HiDrive Legacy] Access token refreshed successfully (expires in 3600 seconds)
[HiDrive Legacy] >>> RunHiDriveLegacyTest betreten für hidrive-legacy-main
[HiDrive Legacy] Upload finished in 2.34s (4.26 MB/s)
[HiDrive Legacy] Download finished in 1.87s (5.35 MB/s)
Test completed for hidrive-legacy-main (hidrive_legacy) in 4.21s
```

## Schritt 5: Monitoring & Metriken

### 5.1 Prometheus Metriken
HiDrive Legacy exportiert Metriken mit `service="hidrive_legacy"` Label:

```promql
# Upload-Performance
nextcloud_test_speed_mbytes_per_sec{service="hidrive_legacy",type="upload"}

# Download-Performance
nextcloud_test_speed_mbytes_per_sec{service="hidrive_legacy",type="download"}

# Test-Erfolg
nextcloud_test_success{service="hidrive_legacy"}

# Test-Dauer
nextcloud_test_duration_seconds{service="hidrive_legacy"}
```

### 5.2 Grafana Dashboard
- **URL**: http://localhost:3003
- **Login**: admin/admin
- **Service-Filter**: Wähle "hidrive_legacy" im Service-Dropdown

## Troubleshooting

### Häufige Fehler

#### 1. "Invalid Client ID"
```
ERROR: OAuth2 token exchange failed: invalid_client
```
**Lösung**: Client ID und Client Secret prüfen

#### 2. "Refresh Token Invalid"
```
ERROR: Token refresh failed: 400 Bad Request
```
**Lösung**: Neuen Refresh Token generieren (siehe Schritt 2) - Refresh Tokens können ablaufen

#### 3. "Insufficient Scope"
```
ERROR: API request failed: 403 Forbidden
```
**Lösung**: App-Berechtigung auf "owner,rw" setzen

#### 4. "Connection Timeout"
```
ERROR: Failed to connect to HiDrive API
```
**Lösung**: Netzwerk-Konnektivität und Firewall prüfen

### Debug-Modus
Für detaillierte Logs:
```bash
# Debug-Logs aktivieren
docker logs monitor-agent -f | grep -i hidrive

# API-Requests debuggen
docker exec monitor-agent curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.hidrive.strato.com/2.1/user/me
```

### Token-Management
HiDrive Legacy nutzt OAuth2 Refresh Tokens für langlebige Authentifizierung:

1. **Refresh Tokens**: Laufen normalerweise nicht ab und können für automatische Token-Erneuerung verwendet werden
2. **Access Tokens**: Haben kurze Lebensdauer (1-4 Stunden) und werden automatisch erneuert
3. **Automatische Erneuerung**: Das System erneuert Access Tokens automatisch bei 401-Fehlern
4. **Produktionssicherheit**: Keine manuelle Token-Verwaltung notwendig

## Multi-Service Integration

Du kannst HiDrive Legacy mit anderen Services kombinieren:

```env
# Nextcloud
NC_INSTANCE_1_URL=https://cloud.company.com
NC_INSTANCE_1_USER=monitor_user
NC_INSTANCE_1_PASS=password

# HiDrive WebDAV
HIDRIVE_INSTANCE_1_URL=https://webdav.hidrive.com
HIDRIVE_INSTANCE_1_USER=webdav_user
HIDRIVE_INSTANCE_1_PASS=webdav_password

# HiDrive Legacy API
HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN=your-refresh-token
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=your-client-id
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=your-client-secret

# Dropbox OAuth2
DROPBOX_INSTANCE_1_REFRESH_TOKEN=your-refresh-token
DROPBOX_INSTANCE_1_APP_KEY=your-app-key
DROPBOX_INSTANCE_1_APP_SECRET=your-app-secret
```

Das System testet alle konfigurierten Services sequenziell und exportiert Service-spezifische Metriken.

## Support

Bei Problemen:
1. **Logs prüfen**: `docker logs monitor-agent`
2. **API-Dokumentation**: https://developer.hidrive.com/
3. **GitHub Issues**: https://github.com/MarcelWMeyer/nextcloud-performance-monitor/issues
