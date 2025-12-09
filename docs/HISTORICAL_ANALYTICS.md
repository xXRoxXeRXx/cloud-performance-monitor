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
cloud_daily_average_upload_speed_mbytes_per_sec{service="nextcloud",instance="cloud.example.com"}
cloud_daily_average_download_speed_mbytes_per_sec{service="hidrive_legacy",instance="hidrive-legacy-main"}
cloud_daily_success_rate_percent{service="dropbox",instance="user@example.com"}
cloud_daily_test_count{service="hidrive",instance="storage.ionos.fr",type="upload"}
```

#### Monatliche Trends (30d)
```prometheus
cloud_monthly_average_upload_speed_mbytes_per_sec{service="nextcloud",instance="cloud.example.com"}
cloud_monthly_average_download_speed_mbytes_per_sec{service="hidrive_legacy",instance="hidrive-legacy-main"}
cloud_monthly_success_rate_percent{service="dropbox",instance="user@example.com"}
```

#### Performance-Ranking
```prometheus
cloud_service_upload_speed_ranking
cloud_service_download_speed_ranking
cloud_best_performing_service
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

- **Raw Metrics**: 2 Jahre / 730 Tage (konfiguriert)
- **Daily Averages**: 30 Tage (Recording Rules)
- **Monthly Averages**: 2 Jahre (Recording Rules)
- **Speicherlimit**: 15 GB (automatische Bereinigung)

### **Performance Impact:**

- **Minimal**: Recording Rules berechnen im Hintergrund
- **Efficient**: Pre-calculated Aggregations
- **Scalable**: Funktioniert mit 1-100+ Services

---

## 📅 Monthly Historical Analytics Dashboard

### **Neues Dashboard: "Cloud Performance - Monthly Historical Analytics"**

Ein dediziertes Dashboard für monatliche historische Performance-Analyse mit langfristigen Trends.

**Dashboard-UID:** `cloud-monthly-performance`  
**Standard-Zeitraum:** 90 Tage (erweiterbar auf 2 Jahre)

### **Dashboard-Panels:**

#### 📊 **Monthly Performance Overview**

- **📤 Monthly Average Upload Speed**: Balkendiagramm mit monatlichen Upload-Durchschnittswerten pro Service
- **📥 Monthly Average Download Speed**: Balkendiagramm mit monatlichen Download-Durchschnittswerten

#### 📈 **Historical Trend Analysis**

- **Daily Average Speed Trend**: Zeitreihen-Diagramm mit täglichen Upload/Download-Durchschnittswerten
- **Smooth Line Interpolation**: Geglättete Kurven für bessere Trendanalyse
- **Legend mit Statistiken**: Mean, Max, Min, Last Value

#### 🏆 **Service Comparison - Monthly Averages**

- **Upload Speed Ranking**: LCD-Balkengauge sortiert nach Performance
- **Download Speed Ranking**: LCD-Balkengauge sortiert nach Performance
- **Farbkodierte Thresholds**: Rot (<2 MB/s), Gelb (2-5 MB/s), Grün (>5 MB/s)

#### 📅 **Month-over-Month Comparison**

- **Current vs Previous Month Upload**: Balkendiagramm mit Vergleich aktueller vs. vorheriger Monat
- **Current vs Previous Month Download**: Balkendiagramm mit Vergleich aktueller vs. vorheriger Monat
- **Offset-basierte Vergleiche**: Verwendet `offset 30d` für präzise Vergleiche

#### 📉 **Monthly Duration Analysis**

- **Upload Duration Trend**: Zeitreihen mit Threshold-Linien (Gelb: 120s, Rot: 300s)
- **Download Duration Trend**: Zeitreihen mit Threshold-Linien (Gelb: 60s, Rot: 180s)
- **Performance-Degradation-Erkennung**: Visuelle Warnung bei Überschreitung

#### ✅ **Monthly Reliability Statistics**

- **Success Rate Gauge**: Prozentuale Erfolgsrate mit Thresholds (Rot <95%, Gelb 95-99%, Grün >99%)
- **Monthly Test Count**: Anzahl durchgeführter Tests mit Trend-Indikator
- **Monthly Errors by Type**: Fehleranzahl gruppiert nach Service und Error-Typ

#### 📊 **Historical Data Table**

- **Performance Summary Table**: Tabellarische Übersicht aller Services
- **Farbkodierte Zellen**: Automatische Farbgebung basierend auf Performance
- **Sortierbar**: Nach Service, Instance, Upload/Download Speed

### **Recording Rules für Monthly Analytics:**

```yaml
# Neue Recording Rules in prometheus/recording_rules.yml
groups:
  - name: monthly_historical_metrics
    interval: 3600s  # Jede Stunde berechnet
    rules:
      # Monatliche Upload-Metriken
      - record: cloud_monthly_upload_speed_avg
      - record: cloud_monthly_upload_speed_max
      - record: cloud_monthly_upload_speed_min
      - record: cloud_monthly_upload_duration_p95
      
      # Monatliche Download-Metriken
      - record: cloud_monthly_download_speed_avg
      - record: cloud_monthly_download_speed_max
      - record: cloud_monthly_download_speed_min
      - record: cloud_monthly_download_duration_p95
      
      # Wöchentliche Metriken (Trend-Vergleich)
      - record: cloud_weekly_upload_speed_avg
      - record: cloud_weekly_download_speed_avg
      
      # Zuverlässigkeits-Metriken
      - record: cloud_monthly_uptime_percent
      - record: cloud_monthly_error_count
      - record: cloud_monthly_total_tests
```

### **Beispiel-Queries für Monthly Analytics:**

```promql
# Monatlicher Durchschnitt Upload-Speed pro Service
avg_over_time(cloud_test_speed_mbytes_per_sec{type="upload"}[30d])

# Month-over-Month Vergleich
avg_over_time(cloud_test_speed_mbytes_per_sec{type="upload"}[30d])
  - 
avg_over_time(cloud_test_speed_mbytes_per_sec{type="upload"}[30d] offset 30d)

# Monatliche Erfolgsrate
(
  sum(sum_over_time(cloud_test_success{error_code="none"}[30d]))
  /
  sum(count_over_time(cloud_test_success[30d]))
) * 100

# Monatliche Fehler nach Typ
sum by (service, error_type) (increase(cloud_test_errors_total[30d]))
```

Die Historical Performance Analytics bieten jetzt vollständige Transparenz über die Langzeit-Performance aller überwachten Cloud-Services! 📊
