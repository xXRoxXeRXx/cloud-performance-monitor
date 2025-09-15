# 📱 MagentaCLOUD Setup Guide

MagentaCLOUD ist ein WebDAV-basierter Cloud-Service der Deutschen Telekom, der eine spezielle Account Number ID (ANID) für die Pfad-Zusammensetzung verwendet.

## 🔧 Konfiguration

### 1. Erforderliche Informationen

Für MagentaCLOUD benötigen Sie:
- **URL**: `https://magentacloud.de` (Standard-URL)
- **Benutzername**: Ihre E-Mail-Adresse (z.B. `user@t-online.de`)
- **Passwort**: Ihr MagentaCLOUD-Passwort oder App-Passwort
- **ANID**: Ihre Account Number ID (siehe unten)

### 2. ANID ermitteln

Die ANID (Account Number ID) ist eine spezielle Kennung für MagentaCLOUD-Konten:

#### Methode 1: Aus der Browser-URL
1. Melden Sie sich bei MagentaCLOUD im Browser an
2. Navigieren Sie zu einem Ordner
3. Die URL enthält Ihre ANID: `https://magentacloud.de/webdav/files/[IHRE_ANID]/`
4. Kopieren Sie die ANID (z.B. `120049010000000114279134`)

#### Methode 2: WebDAV-Test
```bash
# Testen Sie den WebDAV-Zugriff mit curl
curl -u "user@t-online.de:password" \
  "https://magentacloud.de/remote.php/dav/files/" \
  -X PROPFIND

# Die Antwort enthält Ihre ANID in den Pfaden
```

### 3. Environment-Konfiguration

Fügen Sie folgende Variablen zu Ihrer `.env`-Datei hinzu:

```bash
# MagentaCLOUD Instanz 1
MAGENTACLOUD_INSTANCE_1_URL=https://magentacloud.de
MAGENTACLOUD_INSTANCE_1_USER=ihre-email@t-online.de
MAGENTACLOUD_INSTANCE_1_ANID=120049010000000114279134
MAGENTACLOUD_INSTANCE_1_PASS=ihr-passwort

# Weitere MagentaCLOUD Instanzen (optional)
MAGENTACLOUD_INSTANCE_2_URL=https://magentacloud.de
MAGENTACLOUD_INSTANCE_2_USER=andere-email@t-online.de
MAGENTACLOUD_INSTANCE_2_ANID=987654321000000087654321
MAGENTACLOUD_INSTANCE_2_PASS=anderes-passwort
```

## 🔐 Sicherheits-Empfehlungen

### App-Passwort erstellen (Empfohlen)
1. Loggen Sie sich in Ihr Telekom-Konto ein
2. Gehen Sie zu den Sicherheitseinstellungen
3. Erstellen Sie ein neues App-Passwort für "MagentaCLOUD WebDAV"
4. Verwenden Sie dieses App-Passwort statt Ihres Hauptpassworts

### Berechtigungen
Das verwendete Konto benötigt folgende Berechtigungen:
- ✅ Dateien lesen und schreiben
- ✅ Ordner erstellen und löschen
- ✅ WebDAV-Zugriff

## 🧪 Test der Konfiguration

### Manueller WebDAV-Test
```bash
# Test der ANID und Authentifizierung
curl -u "user@t-online.de:password" \
  "https://magentacloud.de/remote.php/dav/files/IHRE_ANID/" \
  -X PROPFIND -v

# Erwartete Antwort: HTTP 207 Multi-Status
```

### Test mit dem Monitor
```bash
# Konfiguration prüfen
docker compose run --rm monitor-agent

# Logs prüfen
docker compose logs monitor-agent --tail=20
```

## 📊 Monitoring-Features

### Unterstützte Funktionen
- ✅ **Chunked Upload**: Dateien werden in 10MB-Chunks übertragen
- ✅ **Download-Tests**: Vollständige Up-/Download-Zyklen
- ✅ **Fehlerbehandlung**: Retry-Logik bei temporären Fehlern
- ✅ **Metriken**: Prometheus-Metriken mit `service=magentacloud`
- ✅ **Multi-Instance**: Mehrere MagentaCLOUD-Konten parallel

### Performance-Metriken
- Upload/Download Geschwindigkeit (MB/s)
- Latenz und Antwortzeiten
- Erfolgsrate und Fehlerquoten
- Chunk-Upload-Statistiken

## ⚠️ Bekannte Einschränkungen

### ANID-Spezifika
- Jedes MagentaCLOUD-Konto hat eine eindeutige ANID
- Die ANID ist erforderlich für alle WebDAV-Pfade
- Falsche ANID führt zu 404-Fehlern

### Rate Limiting
MagentaCLOUD kann bei häufigen Anfragen Rate Limiting anwenden:
- Standard-Testintervall: 5 Minuten (empfohlen)
- Bei Problemen: Intervall auf 10+ Minuten erhöhen

### Dateigröße
- Maximale Testdateigröße: 100MB (Standard: 10MB)
- Chunk-Größe: 10MB (Standard, anpassbar)

## 🐛 Troubleshooting

### Häufige Probleme

#### "401 Unauthorized"
- ✅ Benutzername und Passwort prüfen
- ✅ App-Passwort statt Hauptpasswort verwenden
- ✅ WebDAV-Zugriff in MagentaCLOUD aktiviert?

#### "404 Not Found"
- ✅ ANID korrekt? (Siehe ANID-Ermittlung oben)
- ✅ URL korrekt: `https://magentacloud.de`

#### "403 Forbidden"
- ✅ Account hat WebDAV-Berechtigung?
- ✅ Rate Limiting? Testintervall erhöhen

#### Langsame Performance
- ✅ Internetverbindung prüfen
- ✅ Chunk-Größe anpassen (`TEST_CHUNK_SIZE_MB`)
- ✅ Testdateigröße reduzieren (`TEST_FILE_SIZE_MB`)

### Debug-Logs
```bash
# Detaillierte Logs anzeigen
docker compose logs monitor-agent -f

# Nur MagentaCLOUD-Logs
docker compose logs monitor-agent | grep -i magenta
```

## 🔗 Weiterführende Links

- [MagentaCLOUD Support](https://www.telekom.de/hilfe/magentacloud)
- [WebDAV-Konfiguration](https://www.telekom.de/hilfe/magentacloud/webdav)
- [Telekom Kundencenter](https://www.telekom.de/kundencenter)

## 📈 Beispiel-Metriken

```prometheus
# Upload-Geschwindigkeit
nextcloud_test_speed_mbytes_per_sec{service="magentacloud",instance="https://magentacloud.de",type="upload"} 2.53

# Download-Geschwindigkeit  
nextcloud_test_speed_mbytes_per_sec{service="magentacloud",instance="https://magentacloud.de",type="download"} 6.01

# Test-Erfolg
nextcloud_test_success{service="magentacloud",instance="https://magentacloud.de",type="upload",error_code="none"} 1
```

Diese Metriken werden automatisch in Grafana visualisiert und für Alerting verwendet.
