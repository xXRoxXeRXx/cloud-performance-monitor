# 📊 Historical Performance Analytics

## Übersicht

Das Enhanced Analytics System bietet umfassende historische Performance-Analyse mit automatisch berechneten Tages- und Monatsdurchschnitten für alle überwachten Cloud-Services.

## 🎯 Neue Features

### **📈 Automatische Durchschnittsberechnung**
- **Tägliche Durchschnitte** (24 Stunden): Upload/Download Geschwindigkeit, Erfolgsrate
- **Monatliche Durchschnitte** (30 Tage): Langzeit-Performance-Trends
- **Service-übergreifende Vergleiche**: Ranking und Best-Performer-Analyse
- **Performance-Trend-Erkennung**: Automatische Degradation Detection

### **📊 Erweiterte Metriken**

#### Tägliche Performance (24h)
```prometheus
nextcloud_daily_average_upload_speed_mbytes_per_sec{service="nextcloud",instance="cloud.example.com"}
nextcloud_daily_average_download_speed_mbytes_per_sec{service="hidrive_legacy",instance="hidrive-legacy-main"}
nextcloud_daily_success_rate_percent{service="dropbox",instance="user@example.com"}
nextcloud_daily_test_count{service="hidrive",instance="storage.ionos.fr",type="upload"}
```

#### Monatliche Trends (30d)
```prometheus
nextcloud_monthly_average_upload_speed_mbytes_per_sec{service="nextcloud",instance="cloud.example.com"}
nextcloud_monthly_average_download_speed_mbytes_per_sec{service="hidrive_legacy",instance="hidrive-legacy-main"}
nextcloud_monthly_success_rate_percent{service="dropbox",instance="user@example.com"}
```

#### Performance-Ranking
```prometheus
nextcloud_service_upload_speed_ranking
nextcloud_service_download_speed_ranking
nextcloud_best_performing_service
```

## 🎛️ Enhanced Grafana Dashboard

### **Neue Dashboard: "Enhanced Analytics"**
Zusätzlich zum Standard-Dashboard steht ein erweitertes Analytics-Dashboard zur Verfügung:

**URL:** http://localhost:3003  
**Dashboard:** "Nextcloud Performance Monitor - Enhanced Analytics"

### **Dashboard-Bereiche:**

#### 📊 **Current Performance**
- Aktuelle Upload/Download Geschwindigkeiten
- 24h Erfolgsrate
- Anzahl Tests heute
- Farbkodierte Thresholds (Rot/Gelb/Grün)

#### 📈 **Daily Averages (24 Hours)**
- **Bar Charts**: Tagesvergleich zwischen Services
- **Service Ranking**: Upload/Download Performance Ranking
- **Horizontal Bar Gauges**: Visuelle Service-Vergleiche

#### 📅 **Monthly Trends (30 Days)**
- **Monatsdurchschnitte**: Upload/Download Trends
- **Time Series**: Historische Performance-Entwicklung
- **Trend-Analyse**: Langzeit-Performance-Vergleiche

#### 🏆 **Service Performance Ranking**
- **Performance-Tabelle**: Sortiert nach Geschwindigkeit
- **Real-time vs. Durchschnitt**: Current vs. Daily Average Vergleich
- **Best Performer**: Automatische Identifikation der schnellsten Services

## ⚙️ Prometheus Recording Rules

Das System verwendet automatische Recording Rules für effiziente Berechnung:

### **Calculation Intervals:**
- **Performance Averages**: Alle 5 Minuten (`300s`)
- **Trend Analysis**: Alle 30 Minuten (`1800s`)
- **Service Comparisons**: Alle 10 Minuten (`600s`)

### **Automatic Alerts:**
- **Performance Degradation**: Alarm wenn Current < 80% des 24h Durchschnitts
- **Service Ranking**: Automatische Best/Worst Performer Erkennung

## 📈 Beispiel-Queries

### Täglicher Service-Vergleich
```promql
# Beste Upload-Performance heute
sort_desc(
  nextcloud_daily_average_upload_speed_mbytes_per_sec
)

# Service mit höchster Erfolgsrate  
sort_desc(
  nextcloud_daily_success_rate_percent
)
```

### Trend-Analyse
```promql
# Performance-Trend (Current vs. Daily Average)
(
  rate(nextcloud_test_speed_mbytes_per_sec{type="upload"}[1h]) /
  avg_over_time(nextcloud_test_speed_mbytes_per_sec{type="upload"}[24h])
)

# Performance Degradation Detection
(
  avg_over_time(nextcloud_test_speed_mbytes_per_sec[1h]) /
  avg_over_time(nextcloud_test_speed_mbytes_per_sec[24h])
) < 0.8
```

## 🚀 Deployment

Die Enhanced Analytics sind automatisch aktiv nach:

```bash
# System neu bauen mit Enhanced Features
docker compose down
docker compose build --no-cache
docker compose up -d

# Grafana öffnen
open http://localhost:3003
# Wähle: "Nextcloud Performance Monitor - Enhanced Analytics"
```

## 📊 Business Value

### **Für IT-Management:**
- **SLA-Monitoring**: Monatliche Performance-Trends
- **Service-Vergleiche**: ROI-Analyse verschiedener Cloud-Provider
- **Capacity Planning**: Historische Daten für Ressourcen-Planung

### **Für Operations:**
- **Performance Degradation**: Frühwarnung bei Leistungseinbußen
- **Best Practice**: Identifikation der effizientesten Services
- **Trend-Analyse**: Proaktive Problem-Erkennung

### **Für Business:**
- **Cost Optimization**: Performance/Preis-Verhältnis-Analyse
- **Service Selection**: Datenbasierte Provider-Entscheidungen
- **KPI Reporting**: Automatisierte Performance-Reports

## 🔧 Technische Details

### **Data Retention:**
- **Raw Metrics**: 15 Tage (Prometheus Standard)
- **Daily Averages**: 30 Tage (Recording Rules)
- **Monthly Averages**: 1 Jahr (Recording Rules)

### **Performance Impact:**
- **Minimal**: Recording Rules berechnen im Hintergrund
- **Efficient**: Pre-calculated Aggregations
- **Scalable**: Funktioniert mit 1-100+ Services

Die Historical Performance Analytics bieten jetzt vollständige Transparenz über die Langzeit-Performance aller überwachten Cloud-Services! 📊
