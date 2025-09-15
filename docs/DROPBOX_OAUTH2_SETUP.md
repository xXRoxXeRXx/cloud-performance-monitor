# Dropbox OAuth2 Refresh Token Setup - Schritt für Schritt

## Übersicht
Um einen Dropbox Refresh Token zu erhalten, musst du den OAuth2 Authorization Code Flow durchführen. Hier ist eine praktische Anleitung:

## Schritt 1: Dropbox App erstellen/konfigurieren

### 1.1 App erstellen (falls noch nicht vorhanden)
1. Gehe zu: https://www.dropbox.com/developers/apps
2. Klicke **"Create app"**
3. Wähle **"Dropbox API"**
4. Wähle **"Full Dropbox"** access
5. Gib einen Namen ein (z.B. "Performance Monitor")

### 1.2 App konfigurieren
1. In deiner App unter **"Settings"**
2. Notiere dir:
   - **App key** (wird als `DROPBOX_INSTANCE_1_APP_KEY` benötigt)
   - **App secret** (wird als `DROPBOX_INSTANCE_1_APP_SECRET` benötigt)
3. Unter **"OAuth 2"** → **"Redirect URIs"**:
   - Füge hinzu: `http://localhost:8080/callback`
   - Klicke **"Add"**

## Schritt 2: Authorization Code bekommen

### 2.1 Authorization URL erstellen
Ersetze `YOUR_APP_KEY` mit deinem echten App Key:

```
https://www.dropbox.com/oauth2/authorize?client_id=YOUR_APP_KEY&response_type=code&redirect_uri=http://localhost:8080/callback&token_access_type=offline
```

**Wichtig**: `token_access_type=offline` sorgt dafür, dass ein Refresh Token zurückgegeben wird!

### 2.2 Authorization durchführen
1. Öffne die URL in deinem Browser
2. Logge dich in Dropbox ein (falls nötig)
3. Klicke **"Allow"** um der App Zugriff zu gewähren
4. Du wirst zu `http://localhost:8080/callback?code=AUTHORIZATION_CODE` weitergeleitet
5. **Kopiere den `code` Parameter** aus der URL (das ist dein Authorization Code)

Beispiel URL nach Redirect:
```
http://localhost:8080/callback?code=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnop
```
→ Authorization Code: `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnop`

## Schritt 3: Access Token + Refresh Token erhalten

### 3.1 Mit curl (Linux/macOS/WSL)
```bash
curl -X POST https://api.dropboxapi.com/oauth2/token \
  -d grant_type=authorization_code \
  -d code=YOUR_AUTHORIZATION_CODE \
  -d redirect_uri=http://localhost:8080/callback \
  -u YOUR_APP_KEY:YOUR_APP_SECRET
```

### 3.2 Mit PowerShell (Windows)
```powershell
$appKey = "YOUR_APP_KEY"
$appSecret = "YOUR_APP_SECRET"
$authCode = "YOUR_AUTHORIZATION_CODE"

$credentials = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${appKey}:${appSecret}"))

$body = @{
    grant_type = "authorization_code"
    code = $authCode
    redirect_uri = "http://localhost:8080/callback"
}

$headers = @{
    Authorization = "Basic $credentials"
    "Content-Type" = "application/x-www-form-urlencoded"
}

$response = Invoke-RestMethod -Uri "https://api.dropboxapi.com/oauth2/token" -Method Post -Body $body -Headers $headers
$response | ConvertTo-Json -Depth 3
```

### 3.3 Mit Python (falls verfügbar)
```python
import requests
import base64

app_key = "YOUR_APP_KEY"
app_secret = "YOUR_APP_SECRET"
auth_code = "YOUR_AUTHORIZATION_CODE"

# Basic Authentication Header
credentials = base64.b64encode(f"{app_key}:{app_secret}".encode()).decode()

headers = {
    "Authorization": f"Basic {credentials}",
    "Content-Type": "application/x-www-form-urlencoded"
}

data = {
    "grant_type": "authorization_code",
    "code": auth_code,
    "redirect_uri": "http://localhost:8080/callback"
}

response = requests.post("https://api.dropboxapi.com/oauth2/token", headers=headers, data=data)
print(response.json())
```

## Schritt 4: Response verstehen

Die Antwort sollte so aussehen:
```json
{
  "access_token": "sl.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "token_type": "bearer",
  "expires_in": 14400,
  "refresh_token": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "scope": "account_info.read files.content.read files.content.write files.metadata.read files.metadata.write",
  "uid": "123456789",
  "account_id": "dbid:xxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

**Wichtige Werte für deine .env:**
- `refresh_token` → `DROPBOX_INSTANCE_1_REFRESH_TOKEN`
- `access_token` → Wird automatisch generiert, nicht in .env nötig

## Schritt 5: .env Konfiguration

Erstelle/aktualisiere deine `.env` Datei:
```bash
# Dropbox OAuth2 Configuration
DROPBOX_INSTANCE_1_REFRESH_TOKEN=your_refresh_token_from_step_4
DROPBOX_INSTANCE_1_APP_KEY=your_app_key_from_step_1
DROPBOX_INSTANCE_1_APP_SECRET=your_app_secret_from_step_1
DROPBOX_INSTANCE_1_NAME=dropbox-main

# Test Configuration
TEST_FILE_SIZE_MB=10
TEST_INTERVAL_SECONDS=300
TEST_CHUNK_SIZE_MB=5
```

## Schritt 6: Testen

```bash
# Container neustarten
docker compose restart monitor-agent

# Logs prüfen
docker logs monitor-agent --tail 30

# Nach Dropbox-spezifischen Logs suchen
docker logs monitor-agent 2>&1 | findstr /i "dropbox"
```

## Troubleshooting

### Problem: "invalid_grant" Fehler
- **Authorization Code bereits verwendet**: Codes können nur einmal verwendet werden
- **Code abgelaufen**: Authorization Codes sind nur ~10 Minuten gültig
- **Lösung**: Neuen Authorization Code anfordern (ab Schritt 2.1)

### Problem: "invalid_client" Fehler  
- **App Key/Secret falsch**: Werte aus Dropbox Developer Console überprüfen
- **Basic Auth fehlerhaft**: Sicherstellen, dass `app_key:app_secret` korrekt base64-encodiert ist

### Problem: Kein "refresh_token" in Response
- **`token_access_type=offline` fehlt**: In Authorization URL hinzufügen
- **App-Permissions**: Sicherstellen, dass "Full Dropbox" access gewährt wurde

### Problem: Browser zeigt "Site can't be reached"
- **Das ist normal!** Der localhost:8080 Server läuft nicht
- **Wichtig**: Nur den `code` Parameter aus der URL kopieren
- **URL-Beispiel**: `http://localhost:8080/callback?code=HIER_IST_DER_CODE`

## Sicherheitshinweise

- **App Secret**: Niemals öffentlich teilen oder in Git committen
- **Refresh Token**: Sicher speichern, ermöglicht dauerhaften Zugriff
- **Authorization Code**: Nur einmal verwendbar, schnell verarbeiten
- **Access Token**: Wird automatisch generiert und erneuert

## Nächste Schritte

Nach erfolgreicher Konfiguration:
1. System startet automatisch mit OAuth2
2. Access Tokens werden automatisch erneuert
3. Keine manuellen Token-Updates mehr nötig
4. 24/7 Monitoring ohne Unterbrechungen
