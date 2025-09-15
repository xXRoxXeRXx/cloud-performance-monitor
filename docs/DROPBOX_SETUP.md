# Dropbox Integration Setup Guide

## Dropbox App-Konfiguration

### 1. Dropbox App erstellen

1. Besuche die [Dropbox App Console](https://www.dropbox.com/developers/apps)
2. Klicke auf "Create app"
3. Wähle die folgenden Einstellungen:
   - **Choose an API**: Dropbox API
   - **Choose the type of access**: Full Dropbox (für vollständigen Zugriff)
   - **Name your app**: `cloud-performance-monitor` (oder ein anderer Name)

### 2. Access Token generieren

1. Gehe zu deiner erstellten App in der Console
2. Navigiere zum "Settings" Tab
3. Scrolle zu "OAuth 2" Sektion
4. Klicke auf "Generate access token"
5. Kopiere den generierten Token (beginnt normalerweise mit `sl.`)

**⚠️ Wichtig**: Der Access Token gewährt vollen Zugriff auf deinen Dropbox-Account. Behandle ihn wie ein Passwort!

### 3. Token-Berechtigungen (Scopes)

Stelle sicher, dass deine App die folgenden Berechtigungen hat:
- `files.metadata.read` - Lesen von Datei-Metadaten
- `files.content.read` - Lesen von Dateiinhalten
- `files.content.write` - Schreiben von Dateiinhalten

Diese werden automatisch bei "Full Dropbox" Access gesetzt.

## Umgebungsvariablen-Konfiguration

Füge folgende Variablen zu deiner `.env` Datei hinzu:

```bash
# === Dropbox Instanzen ===
# Erste Dropbox-Instanz
DROPBOX_INSTANCE_1_TOKEN=sl.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
DROPBOX_INSTANCE_1_NAME=dropbox-main

# Weitere Instanzen (optional)
DROPBOX_INSTANCE_2_TOKEN=sl.yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
DROPBOX_INSTANCE_2_NAME=dropbox-backup
```

### Parameter-Erklärung

- **DROPBOX_INSTANCE_X_TOKEN**: Der Access Token aus der Dropbox App Console
- **DROPBOX_INSTANCE_X_NAME**: Optional. Name für Monitoring (Standard: `dropbox-instance-X`)

## Test der Konfiguration

1. Führe den Monitor aus:
```bash
make run
```

2. Überprüfe die Logs für Dropbox-Aktivität:
```bash
docker logs monitor-agent | grep Dropbox
```

3. Überprüfe Metrics in Prometheus (http://localhost:9090):
```
nextcloud_test_success{service="dropbox"}
nextcloud_test_duration_seconds{service="dropbox"}
nextcloud_test_speed_mbytes_per_sec{service="dropbox"}
```

## Sicherheitshinweise

### Token-Sicherheit
- **Nie** Access Tokens in öffentliche Repositories committen
- Verwende `.env`-Dateien für lokale Entwicklung
- In Produktion: Environment Variables oder Secrets Management
- Rotiere Tokens regelmäßig (alle 6-12 Monate)

### Produktions-Deployment
```bash
# Docker-Umgebung
docker run -e DROPBOX_INSTANCE_1_TOKEN=sl.xxx... nextcloud-monitor-agent

# Kubernetes Secret
kubectl create secret generic dropbox-tokens \
  --from-literal=token1=sl.xxx...
```

### Überwachung und Limits

Die Dropbox API hat folgende Rate Limits:
- **Upload/Download**: 300 requests/hour für einzelne Benutzer
- **Metadata**: 1000 requests/hour

Der Monitor berücksichtigt diese Limits automatisch.

## Fehlerbehebung

### Häufige Fehler

**401 Unauthorized**
- Token ist ungültig oder abgelaufen
- App-Berechtigungen sind nicht ausreichend
- Lösung: Neuen Token generieren

**403 Forbidden**
- Rate Limit erreicht
- Lösung: Warten bis zum Reset oder Intervall vergrößern

**400 Bad Request**
- Pfad-Format falsch (muss mit "/" beginnen)
- Lösung: Überprüfe Pfad-Konfiguration

### Debug-Modus

Aktiviere detaillierte Logs:
```bash
docker logs -f monitor-agent | grep -E "(Dropbox|ERROR)"
```

### Test-Verbindung

Teste deine Dropbox-Verbindung:
```bash
curl -X POST https://api.dropboxapi.com/2/users/get_current_account \
  --header "Authorization: Bearer sl.your-token-here"
```

## Monitoring-Dashboard

Nach erfolgreicher Konfiguration siehst du in Grafana (http://localhost:3000):
- Upload/Download-Geschwindigkeiten für Dropbox
- Test-Erfolgsraten
- Chunk-Upload-Statistiken
- Circuit Breaker Status

Die Dropbox-Metriken erscheinen mit dem Label `service="dropbox"`.
