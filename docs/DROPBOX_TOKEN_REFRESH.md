# Dropbox OAuth2 Integration

## OAuth2-Only Implementation ✅ VOLLSTÄNDIG IMPLEMENTIERT

Die Dropbox-Integration verwendet **ausschließlich OAuth2 mit Refresh Tokens** für automatische Token-Erneuerung. Legacy Access Tokens werden nicht mehr unterstützt.

### Vorteile:
- **Automatische Token-Erneuerung**: Keine manuellen Eingriffe nötig
- **Längere Gültigkeit**: Refresh Tokens laufen nicht ab  
- **Produktionsbereit**: Für 24/7 Monitoring geeignet
- **Vereinfachte Architektur**: Ein einheitlicher OAuth2-Client

## Setup

### 1. Dropbox App erstellen
1. Gehe zu: https://www.dropbox.com/developers/apps
2. Klicke **"Create app"**
3. Wähle **"Dropbox API"**
4. Wähle **"Full Dropbox"** access
5. Gib deiner App einen Namen (z.B. "Performance Monitor")

### 2. OAuth2 Credentials erhalten
1. In der App-Konfiguration unter **"Settings"**:
   - Notiere **App Key** und **App Secret**
2. Unter **"OAuth 2"**:
   - Setze eine **Redirect URI** (z.B. `http://localhost:8080/callback`)
3. **Authorization Code Flow** durchführen:
   ```bash
   # Schritt 1: Authorization URL besuchen
   https://www.dropbox.com/oauth2/authorize?client_id=YOUR_APP_KEY&response_type=code&redirect_uri=YOUR_REDIRECT_URI
   
   # Schritt 2: Authorization Code aus Callback URL extrahieren
   # Schritt 3: Access + Refresh Token erhalten
   curl -X POST https://api.dropboxapi.com/oauth2/token \
     -d grant_type=authorization_code \
     -d code=YOUR_AUTHORIZATION_CODE \
     -d redirect_uri=YOUR_REDIRECT_URI \
     -u YOUR_APP_KEY:YOUR_APP_SECRET
   ```

### 3. .env Konfiguration
```bash
# OAuth2 Credentials (einzige unterstützte Methode)
DROPBOX_INSTANCE_1_REFRESH_TOKEN=your_refresh_token_here
DROPBOX_INSTANCE_1_APP_KEY=your_app_key_here
DROPBOX_INSTANCE_1_APP_SECRET=your_app_secret_here
DROPBOX_INSTANCE_1_NAME=dropbox-main

# Weitere Instanzen
DROPBOX_INSTANCE_2_REFRESH_TOKEN=your_refresh_token_2
DROPBOX_INSTANCE_2_APP_KEY=your_app_key_2
DROPBOX_INSTANCE_2_APP_SECRET=your_app_secret_2
DROPBOX_INSTANCE_2_NAME=dropbox-backup
```

## Automatische Token-Erneuerung

### Arbeitsweise:
1. **Client-Initialisierung**: System erstellt OAuth2-Client mit Refresh Token
2. **Initial Token Generation**: Erster Access Token wird automatisch generiert
3. **API-Aufrufe**: System verwendet aktuellen Access Token
4. **401 Unauthorized**: System erkennt abgelaufenen Token automatisch  
5. **Automatischer Refresh**: Neuer Access Token wird über Refresh Token angefordert
6. **Retry**: API-Aufruf wird mit neuem Token wiederholt
7. **Logging**: Alle Aktionen werden geloggt

### Log-Beispiele:
```
[Dropbox] Using OAuth2 client with refresh token for dropbox-main
[Dropbox] OAuth2 access token generated successfully for dropbox-main
[Dropbox] Access token expired, attempting refresh...  
[Dropbox] Access token refreshed successfully (expires in 14400 seconds)
[Dropbox] Retrying request with refreshed token...
```

## Konfigurationsvalidierung

Das System validiert automatisch, dass alle erforderlichen OAuth2-Felder vorhanden sind:

```bash
# ✅ Gültig - alle OAuth2-Felder vorhanden
DROPBOX_INSTANCE_1_REFRESH_TOKEN=xxxxx
DROPBOX_INSTANCE_1_APP_KEY=xxxxx  
DROPBOX_INSTANCE_1_APP_SECRET=xxxxx

# ❌ Ungültig - unvollständige OAuth2-Konfiguration
DROPBOX_INSTANCE_1_REFRESH_TOKEN=xxxxx
# Fehlende APP_KEY und APP_SECRET führt zu Konfigurationsfehler
```

## Troubleshooting

### Problem: "refresh token, app key, or app secret not available"
**Lösung**: Alle drei OAuth2-Felder müssen konfiguriert sein:
- `DROPBOX_INSTANCE_X_REFRESH_TOKEN`
- `DROPBOX_INSTANCE_X_APP_KEY`  
- `DROPBOX_INSTANCE_X_APP_SECRET`

### Problem: "invalid_grant" bei Refresh
- **Refresh Token abgelaufen**: Neue OAuth2-Autorisierung erforderlich
- **App-Credentials falsch**: App Key/Secret überprüfen

### Problem: "invalid_client" 
- **App Key/Secret falsch**: Credentials in Dropbox Developer Console überprüfen
- **App permissions**: Sicherstellen, dass "Full Dropbox" access gewährt wurde

### Problem: "Failed to generate initial access token"
- **Refresh Token ungültig**: OAuth2-Flow neu durchführen
- **Network Issues**: Internetverbindung und Firewall prüfen

## Monitoring & Alerting

### Prometheus Metrics:
```
# Token Refresh Events
dropbox_token_refresh_total{instance="dropbox-main",status="success"} 1
dropbox_token_refresh_total{instance="dropbox-main",status="failed"} 0

# Auth Errors
dropbox_auth_errors_total{instance="dropbox-main",error_type="oauth2_failed"} 0
dropbox_auth_errors_total{instance="dropbox-main",error_type="invalid_grant"} 0
```

### Debug-Befehle:
```bash
# OAuth2-spezifische Logs anzeigen
docker logs monitor-agent 2>&1 | grep -E "(OAuth2|refresh|token)"

# Dropbox-spezifische Fehler
docker logs monitor-agent 2>&1 | grep -i "dropbox.*error"

# Erfolgreiche Dropbox-Tests  
docker logs monitor-agent 2>&1 | grep "Dropbox.*completed"

# Token-Refresh-Events
docker logs monitor-agent 2>&1 | grep "access token refreshed"
```

## Migration von Legacy Tokens

### Wichtiger Hinweis: 
**Legacy Access Tokens (`DROPBOX_INSTANCE_X_TOKEN`) werden nicht mehr unterstützt!**

Falls du bisher Legacy Tokens verwendet hast:

1. **OAuth2-Credentials beschaffen** (siehe Setup oben)
2. **`.env` aktualisieren**:
   ```bash
   # Alt (nicht mehr unterstützt):
   # DROPBOX_INSTANCE_1_TOKEN=sl.xxxxx

   # Neu (erforderlich):
   DROPBOX_INSTANCE_1_REFRESH_TOKEN=xxxxx
   DROPBOX_INSTANCE_1_APP_KEY=xxxxx
   DROPBOX_INSTANCE_1_APP_SECRET=xxxxx
   ```
3. **Container neustarten**: `docker compose restart monitor-agent`

## API-Limits & Best Practices

### Rate Limiting:
- **Dropbox API**: 120 requests/minute per app
- **Automatisches Retry**: Bei 429-Fehlern mit exponential backoff
- **Token Refresh**: Maximal 1x pro Minute pro Instance

### Security:
- **App Secret**: Niemals in Logs oder öffentlichen Repos
- **Refresh Token**: Sichere Speicherung, regelmäßige Rotation
- **Least Privilege**: Nur benötigte Scopes verwenden

## Status

- **✅ OAuth2 Implementation**: Vollständig implementiert und getestet
- **✅ Automatische Token-Erneuerung**: Produktionsbereit
- **✅ Legacy Token Entfernung**: Vereinfachte, einheitliche Architektur
- **✅ Umfassende Validierung**: Fehlerbehandlung und Logging
- **� Live**: Bereit für Production-Deployment
