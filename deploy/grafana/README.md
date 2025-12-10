# Grafana Dashboards - Cloud Performance Monitor

Dieses Verzeichnis enthält alle Grafana-Dashboards für das Cloud Performance Monitor System.

## 📊 Verfügbare Dashboards

### 1. Main Dashboard (`dashboard.json`)
**Zweck**: Haupt-Performance-Übersicht
- Service-Verfügbarkeit und Uptime
- Upload/Download Performance-Metriken
- Netzwerk-Latenz Monitoring
- Test-Duration Trends
- Service-Vergleiche

### 2. Enhanced Dashboard (`enhanced-dashboard.json`)
**Zweck**: Erweiterte Analysen und historische Trends
- Detaillierte Performance-Analysen
- Historische Datenanalyse
- Trend-Analysen
- Detaillierte Service-Aufschlüsselungen

### 3. Enhanced Alt (`dashboard-enhanced.json`)
**Zweck**: Alternative erweiterte Ansicht
- Zusätzliche Visualisierungen
- Erweiterte Metriken

### 4. Error Analysis Dashboard (`errors-dashboard.json`)
**Zweck**: Fehlerüberwachung und Troubleshooting
- Echtzeit Fehler-Tracking
- Fehler-Kategorisierung und Trends
- Success Rate Monitoring
- Error Code Analyse

**Fehler-Kategorien**:
- **HTTP 5xx Errors**: Server-seitige Probleme (rot)
- **HTTP 4xx Errors**: Client-seitige Probleme (orange)
- **Network Errors**: Verbindungsprobleme (lila)
- **Authentication Errors**: Auth/Token-Probleme (gelb)
- **File Operation Errors**: Upload/Download-Probleme (blau)
- **WebDAV Errors**: Protokoll-spezifische Probleme (grün)

### 5. Daily Performance Dashboard (`daily-performance-dashboard.json`)
**Zweck**: Tägliche Rohdaten ohne Aggregation
- Aktuelle Performance-Werte
- Rohdaten-Visualisierung
- Keine zeitliche Aggregation
- Ideal für kurzfristige Analysen

### 6. Monthly Performance Dashboard (`monthly-performance-dashboard.json`)
**Zweck**: Monatliche Trends und Durchschnitte
- 30-Tage-Durchschnitte
- Langzeit-Performance-Trends
- SLA-Compliance Monitoring
- Kapazitätsplanung

## � Filter-Optionen

Alle Dashboards bieten:
- **Service Filter**: Nextcloud, HiDrive, MagentaCLOUD, Dropbox, HiDrive Legacy
- **Instance Filter**: Spezifische URLs/Namen der Instanzen
- **Time Range**: Anpassbarer Zeitraum

## � Key Metrics

### Performance Metriken
```promql
# Upload/Download Dauer
cloud_test_duration_seconds{service="...", instance="...", type="upload|download"}

# Übertragungsgeschwindigkeit
cloud_test_speed_mbytes_per_sec{service="...", instance="...", type="upload|download"}

# Netzwerk-Latenz
cloud_network_latency_seconds{service="...", instance="..."}
```

### Error Tracking
```promql
# Erfolgsrate
cloud_test_success{service="...", instance="...", error_code="..."}

# Fehler-Zähler
cloud_test_errors_total{service="...", instance="...", error_code="..."}
```

## 🚀 Setup

### Automatischer Import
Dashboards werden automatisch beim Start importiert:
```bash
docker compose up -d
```

### Zugriff
- **Grafana URL**: http://localhost:3003
- **Default Login**: admin / admin

## 📁 Verzeichnisstruktur

```
deploy/grafana/
├── dashboard.json                      # Main Dashboard
├── enhanced-dashboard.json             # Enhanced Analytics
├── dashboard-enhanced.json             # Enhanced Alt
├── errors-dashboard.json               # Error Analysis
├── daily-performance-dashboard.json    # Daily Performance
├── monthly-performance-dashboard.json  # Monthly Trends
├── Dockerfile                          # Grafana Container
├── README.md                           # Diese Datei
└── provisioning/
    ├── dashboards/
    │   └── dashboard.yaml              # Dashboard Provisioning
    └── datasources/
        └── prometheus.yaml             # Prometheus Datasource
```

## 📚 Weiterführende Dokumentation

- [Email Configuration](../../docs/EMAIL_CONFIGURATION.md)
- [Port Security](../../docs/PORT_SECURITY.md)
- [Alert Error Codes](../../docs/ALERT_ERROR_CODES.md)