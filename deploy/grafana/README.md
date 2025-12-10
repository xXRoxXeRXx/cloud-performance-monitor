# Grafana Dashboards# Grafana Dashboards - Cloud Performance Monitor



## 📊 Available Dashboards (3)Dieses Verzeichnis enthält alle Grafana-Dashboards für das Cloud Performance Monitor System.



### 1. Daily Performance (`daily-performance-dashboard.json`)## 📊 Verfügbare Dashboards

- Current upload/download speeds

- Real-time performance metrics### 1. Main Dashboard (`dashboard.json`)

- 24-hour view**Zweck**: Haupt-Performance-Übersicht

- Service-Verfügbarkeit und Uptime

### 2. Monthly Performance (`monthly-performance-dashboard.json`)- Upload/Download Performance-Metriken

- 30-day trends and averages- Netzwerk-Latenz Monitoring

- Long-term performance analysis- Test-Duration Trends

- SLA compliance monitoring- Service-Vergleiche



### 3. Errors (`errors-dashboard.json`)### 2. Enhanced Dashboard (`enhanced-dashboard.json`)

- Error tracking and categorization**Zweck**: Erweiterte Analysen und historische Trends

- Success rate monitoring- Detaillierte Performance-Analysen

- Error code analysis- Historische Datenanalyse

- Trend-Analysen

## 🔍 Filters- Detaillierte Service-Aufschlüsselungen



All dashboards support:### 3. Enhanced Alt (`dashboard-enhanced.json`)

- **Service**: nextcloud, hidrive, magentacloud, dropbox, hidrive_legacy**Zweck**: Alternative erweiterte Ansicht

- **Instance**: Specific URLs/names- Zusätzliche Visualisierungen

- **Time Range**: Adjustable- Erweiterte Metriken



## 📈 Key Metrics### 4. Error Analysis Dashboard (`errors-dashboard.json`)

**Zweck**: Fehlerüberwachung und Troubleshooting

```promql- Echtzeit Fehler-Tracking

# Duration- Fehler-Kategorisierung und Trends

cloud_test_duration_seconds{service="...", type="upload|download"}- Success Rate Monitoring

- Error Code Analyse

# Speed

cloud_test_speed_mbytes_per_sec{service="...", type="upload|download"}**Fehler-Kategorien**:

- **HTTP 5xx Errors**: Server-seitige Probleme (rot)

# Success- **HTTP 4xx Errors**: Client-seitige Probleme (orange)

cloud_test_success{service="...", error_code="..."}- **Network Errors**: Verbindungsprobleme (lila)

```- **Authentication Errors**: Auth/Token-Probleme (gelb)

- **File Operation Errors**: Upload/Download-Probleme (blau)

## 🚀 Access- **WebDAV Errors**: Protokoll-spezifische Probleme (grün)



- **URL**: http://localhost:3003### 5. Daily Performance Dashboard (`daily-performance-dashboard.json`)

- **Login**: admin / admin**Zweck**: Tägliche Rohdaten ohne Aggregation

- Aktuelle Performance-Werte

Dashboards are auto-provisioned on startup via `provisioning/`.- Rohdaten-Visualisierung

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