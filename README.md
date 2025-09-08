
# Nextcloud & HiDrive Performance Monitor

[![Build Status](https://img.shields.io/github/actions/workflow/status/MarcelWMeyer/nextcloud-performance-monitor/docker-image.yml?branch=main)](https://github.com/MarcelWMeyer/nextcloud-performance-monitor/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


Ein einfacher, containerisierter Agent zur Überwachung der Upload- und Download-Leistung von Nextcloud- und HiDrive-Instanzen über WebDAV, mit Visualisierung in Grafana.

## Features
- Überwache mehrere Nextcloud- und HiDrive-Instanzen mit einem einzigen Agenten.
- Messe reale Leistung durch synthetische Upload- (mit Chunking) und Download-Tests.
- Biete Metriken im Prometheus-Format (Geschwindigkeit in MB/s, Dauer, Erfolgsrate) – jetzt mit `service`-Label (z.B. nextcloud/hidrive).
- Starte den gesamten Stack (Agent, Prometheus, Grafana) einfach über Docker Compose.
- Enthält ein vorkonfiguriertes, anpassbares Grafana-Dashboard mit Service-Selector für sofortige Visualisierung.



## Tests & Continuous Integration
Das Projekt enthält Go-Unit-Tests (z.B. für die Konfiguration). Tests können lokal ausgeführt werden mit:
```bash
go test -v -cover ./...
```
Die GitHub Actions Pipeline nutzt Caching für Go-Module und Docker-Layer, Coverage-Report und baut/pusht Images automatisch bei neuen Tags.

## Voraussetzungen
Bevor Sie beginnen, stellen Sie sicher, dass Folgendes vorhanden ist:
- Ein Server mit Shell-Zugang (SSH), auf dem Sie den Monitoring-Stack ausführen können.
- **Docker** und **Docker Compose** müssen auf diesem Server installiert sein.
- Ein dedizierter **Nextcloud-Benutzer** für jede zu testende Instanz mit Lese-/Schreibberechtigungen. Es wird dringend empfohlen, ein "App-Passwort" für diesen Benutzer zu erstellen und zu verwenden.

## Schnellstart-Anleitung

### Schritt 1: Projekt herunterladen
Klonen Sie das Projekt-Repository von GitHub auf Ihren Server und navigieren Sie in das Verzeichnis.
```bash
git clone https://github.com/MarcelWMeyer/nextcloud-performance-monitor.git
cd nextcloud-performance-monitor
```

### Schritt 2: Prometheus-Konfiguration erstellen
Prometheus benötigt eine Konfigurationsdatei, um zu wissen, wo Metriken zu finden sind.

Erstellen Sie das Verzeichnis:
```bash
mkdir -p prometheus
```

Erstellen Sie die Konfigurationsdatei `prometheus/prometheus.yml` mit folgendem Inhalt:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'nextcloud-monitor-agent'
    static_configs:
      - targets: ['monitor-agent:8080']
```

### Schritt 3: Instanzen konfigurieren
Kopieren Sie die Beispiel-Konfigurationsdatei. Diese Datei wird verwendet, um Ihre Instanz-Anmeldedaten an den Agenten zu übergeben.
```bash
cp .env.example .env
```


Öffnen Sie die `.env`-Datei mit einem Texteditor (z.B. `nano .env`). Fügen Sie einen nummerierten Block für jede Instanz hinzu, die Sie überwachen möchten:

Beispiel `.env`-Datei für Nextcloud und HiDrive:
```bash
# Nextcloud Instanz 1
NC_INSTANCE_1_URL=https://cloud.company-a.com
NC_INSTANCE_1_USER=monitor_user_a
NC_INSTANCE_1_PASS=super-secret-password-a

# HiDrive Instanz 1
HIDRIVE_INSTANCE_1_URL=https://storage.ionos.fr
HIDRIVE_INSTANCE_1_USER=monitor_user_hidrive
HIDRIVE_INSTANCE_1_PASS=super-secret-password-hidrive
```

### Schritt 4: Monitoring-Stack starten
Starten Sie den gesamten Stack (Agent, Prometheus, Grafana) mit einem einzigen Befehl:
```bash
docker compose up -d
```

Sie können den Status der Container mit `docker compose ps` überprüfen.

### Schritt 5: Grafana einrichten
**Anmelden:** Öffnen Sie Ihren Webbrowser und navigieren Sie zu `http://<ihre-server-ip>:3000`. Melden Sie sich mit den Standard-Grafana-Anmeldedaten an (Benutzer: `admin`, Passwort: `admin`). Sie werden aufgefordert, das Passwort zu ändern.

**Datenquelle hinzufügen:** Gehen Sie im Menü zu **Connections > Data sources > Add new data source**.

Wählen Sie **Prometheus**.

Geben Sie für die "Prometheus server URL" `http://prometheus:9090` ein.

Klicken Sie auf **Save & Test**. Sie sollten eine Erfolgsmeldung sehen.

**Dashboard importieren:** Gehen Sie zu **Dashboards > New > Import**.

Laden Sie die Datei `deploy/grafana/dashboard.json` aus dem Projektverzeichnis hoch.

Wählen Sie im nächsten Schritt die Prometheus-Datenquelle aus, die Sie gerade erstellt haben.

Klicken Sie auf **Import**.

### Schritt 6: Ergebnisse anzeigen

Ihr Dashboard ist jetzt live! Nach einigen Minuten sollten die ersten Datenpunkte erscheinen. Verwenden Sie das Dropdown-Menü "service" oben, um zwischen Nextcloud- und HiDrive-Instanzen zu filtern.
## Prometheus-Metriken (Beispiel)

Die wichtigsten Metriken enthalten jetzt das Label `service`:

```
nextcloud_test_duration_seconds{service="nextcloud",instance="https://cloud.company-a.com",type="upload"} 2.5
nextcloud_test_duration_seconds{service="hidrive",instance="https://storage.ionos.fr",type="upload"} 12.3
```


## Fehlerbehebung
- **Fehler bei `docker compose up`:** Stellen Sie sicher, dass die Dateien `prometheus/prometheus.yml` und `.env` vor dem Ausführen des Befehls existieren.
- **Keine Daten in Grafana?** Überprüfen Sie die Agent-Logs mit `docker compose logs monitor-agent`. Suchen Sie nach Verbindungs- oder Authentifizierungsfehlern im Zusammenhang mit Ihrer Nextcloud-Instanz.
- **"Data source error" in Grafana?** Stellen Sie sicher, dass alle Container laufen (`docker compose ps`) und dass Sie die korrekte URL (`http://prometheus:9090`) beim Einrichten der Datenquelle in Schritt 5 verwendet haben.

## Lizenz
Dieses Projekt ist unter der MIT-Lizenz lizenziert.
