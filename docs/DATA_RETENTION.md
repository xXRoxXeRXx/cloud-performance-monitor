# Data Retention - Cloud Performance Monitor# Datenspeicherung & Retention - Cloud Performance Monitor



## Default Configuration## 📊 Aktuelle Speicherdauer



- **Retention**: 15 days (Prometheus default)### **Rohmetriken (Raw Metrics)**

- **Storage**: Docker volume `prometheus-data`- **Default**: 15 Tage (Prometheus Standard)

- **Scrape Interval**: 15 seconds- **Speicherort**: Docker Volume `prometheus-data`

- **Granularität**: Volle Details alle 15 Sekunden

## Configure Retention

### **Aggregierte Metriken (Recording Rules)**

### Option 1: Extended Retention (Recommended)- **Daily Averages**: Basierend auf letzten 24h (verfügbar für 15 Tage)

- **Monthly Averages**: Basierend auf letzten 30 Tagen (verfügbar für 15 Tage)

For **90 days** data retention:

## ⚙️ Retention konfigurieren

```yaml

# docker-compose.yml### **Option 1: Längere Speicherdauer (Empfohlen)**

services:

  prometheus:Für **90 Tage** Datenspeicherung:

    command:

      - '--config.file=/etc/prometheus/prometheus.yml'```yaml

      - '--storage.tsdb.path=/prometheus'# docker-compose.yml

      - '--storage.tsdb.retention.time=90d'services:

      - '--storage.tsdb.retention.size=50GB'  prometheus:

      - '--web.enable-lifecycle'    command:

```      - '--config.file=/etc/prometheus/prometheus.yml'

      - '--storage.tsdb.path=/prometheus'

### Option 2: Size-Based Only      - '--storage.tsdb.retention.time=90d'  # 90 Tage

      - '--storage.tsdb.retention.size=50GB'  # Max 50GB

```yaml      - '--web.enable-lifecycle'

command:      - '--web.enable-admin-api'

  - '--storage.tsdb.retention.size=100GB'```

  # No time limit - data kept until disk is full

```**Speicherplatz-Kalkulation:**

- **15 Sekunden Interval** = 4 Messungen/Minute

### Option 3: Minimal (Homelab)- **10 Services** mit je ~20 Metriken = 200 Metriken

- **Prometheus Kompression** ~1.5 Bytes/Sample

```yaml- **90 Tage**: ca. **15-20 GB**

command:

  - '--storage.tsdb.retention.time=30d'### **Option 2: Unbegrenzte Dauer (Nur Size-Limit)**

  - '--storage.tsdb.retention.size=10GB'

``````yaml

command:

## Storage Calculation  - '--storage.tsdb.retention.size=100GB'  # Nur Size-basiert

  # Kein time-limit = unbegrenzt bis Plattenplatz voll

``````

Assumptions:

- 5 services, 15s scrape interval, ~25 metrics/service### **Option 3: Kurze Retention + Langzeit-Speicher**



Daily samples = (86400s / 15s) × 5 × 25 = 720,000 samplesFür **sehr lange Speicherung** (Jahre):

Daily storage = ~1 MB (compressed)

```yaml

Realistic estimates (with TSDB overhead):services:

- 30 days: ~500 MB - 1 GB  # ... existing services ...

- 90 days: ~2-5 GB  

- 365 days: ~10-20 GB  thanos-sidecar:

```    image: quay.io/thanos/thanos:v0.35.0

    container_name: thanos-sidecar

## Backup & Restore    command:

      - 'sidecar'

```bash      - '--tsdb.path=/prometheus'

# Backup      - '--prometheus.url=http://prometheus:9090'

docker run --rm \      - '--objstore.config-file=/etc/thanos/bucket.yml'

  -v prometheus-data:/data \      - '--grpc-address=0.0.0.0:10901'

  -v $(pwd)/backup:/backup \      - '--http-address=0.0.0.0:10902'

  alpine tar czf /backup/prometheus-$(date +%Y%m%d).tar.gz -C /data .    volumes:

      - prometheus-data:/prometheus:ro

# Restore      - ./thanos/bucket.yml:/etc/thanos/bucket.yml

docker run --rm \    networks:

  -v prometheus-data:/data \      - monitor-net

  -v $(pwd)/backup:/backup \

  alpine tar xzf /backup/prometheus-20251210.tar.gz -C /data  thanos-store:

```    image: quay.io/thanos/thanos:v0.35.0

    container_name: thanos-store

## Quick Setup    command:

      - 'store'

```bash      - '--objstore.config-file=/etc/thanos/bucket.yml'

# 1. Edit docker-compose.yml (add retention flags)      - '--grpc-address=0.0.0.0:10901'

# 2. Restart stack      - '--http-address=0.0.0.0:10902'

docker compose down && docker compose up -d    volumes:

      - ./thanos/bucket.yml:/etc/thanos/bucket.yml

# 3. Verify    networks:

docker exec prometheus promtool tsdb analyze /prometheus      - monitor-net

``````



## References**Thanos Bucket Config** (`thanos/bucket.yml`):

```yaml

- [Prometheus Storage](https://prometheus.io/docs/prometheus/latest/storage/)type: S3

config:
  bucket: "cloud-monitor-metrics"
  endpoint: "s3.amazonaws.com"
  access_key: "YOUR_ACCESS_KEY"
  secret_key: "YOUR_SECRET_KEY"
```

**Vorteile:**
- ✅ Prometheus: 15-30 Tage (schnell)
- ✅ Thanos S3: Unbegrenzt (günstig)
- ✅ Globale Abfragen über alle Zeiträume

## 📈 Empfohlene Konfiguration

### **Für Produktions-Monitoring:**

```yaml
# docker-compose.yml
services:
  prometheus:
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=90d'      # 90 Tage Raw Metrics
      - '--storage.tsdb.retention.size=50GB'     # Max 50GB
      - '--storage.tsdb.wal-compression'          # WAL Kompression aktiviert
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
```

### **Für kleine Installationen (Homelab):**

```yaml
command:
  - '--storage.tsdb.retention.time=30d'   # 30 Tage ausreichend
  - '--storage.tsdb.retention.size=10GB'  # 10GB Limit
```

### **Für Enterprise (mit Thanos):**

```yaml
# Prometheus: 15d Hot Data
# Thanos: Unbegrenzt Cold Storage in S3/MinIO
command:
  - '--storage.tsdb.retention.time=15d'
  - '--storage.tsdb.min-block-duration=2h'  # Für Thanos
  - '--storage.tsdb.max-block-duration=2h'  # Für Thanos
```

## 🔍 Daten-Aufbewahrungsrichtlinien

### **Gesetzliche Anforderungen (DSGVO/GDPR)**

Performance-Metriken sind normalerweise **keine personenbezogenen Daten**, aber:

- **Service-URLs** könnten internen Infos enthalten
- **Empfehlung**: 90 Tage für Debugging, dann löschen
- **Backup**: Für Compliance-Audits separate Langzeit-Archivierung

### **Best Practices**

| Zeitraum | Verwendungszweck | Empfehlung |
|----------|------------------|------------|
| **0-7 Tage** | Akutes Incident-Response | Rohmetriken mit voller Granularität |
| **7-30 Tage** | Trend-Analyse, Debugging | Rohmetriken + Recording Rules |
| **30-90 Tage** | Quartals-Reports, SLA-Validierung | Rohmetriken (optional) + Recording Rules |
| **90+ Tage** | Jahres-Analysen, Compliance | Nur Recording Rules oder Thanos |
| **1+ Jahre** | Historische Trends | Thanos/Cold Storage mit Downsampling |

## 📊 Speicherplatz-Anforderungen

### **Kalkulation für Ihr Setup:**

```
Annahmen:
- 5 Services (Nextcloud, HiDrive, MagentaCLOUD, Dropbox, HiDrive Legacy)
- 15s Scrape Interval
- ~25 Metriken pro Service
- Prometheus Kompression: 1.5 Bytes/Sample

Berechnung:
- Samples/Tag = (86400s / 15s) * 5 Services * 25 Metriken = 720.000 Samples/Tag
- Storage/Tag = 720.000 * 1.5 Bytes = 1.08 MB/Tag (mit Kompression)
- 30 Tage = ~32 MB
- 90 Tage = ~97 MB
- 365 Tage = ~394 MB

ABER: Prometheus TSDB hat Overhead (Indizes, WAL):
- Real 30d: ~500 MB - 1 GB
- Real 90d: ~2-5 GB
- Real 365d: ~10-20 GB
```

### **Empfohlene Volume-Größen:**

```yaml
# docker-compose.yml - Volume mit Größenlimit
volumes:
  prometheus-data:
    driver: local
    driver_opts:
      type: none
      o: bind,size=50G  # Max 50GB
      device: /mnt/prometheus-data
```

## 🛠️ Wartung & Cleanup

### **Manuelle Daten-Löschung**

```bash
# Zugriff auf Prometheus Container
docker exec -it prometheus sh

# TSDB Status prüfen
promtool tsdb analyze /prometheus

# Alte Blöcke löschen (Vorsicht!)
rm -rf /prometheus/01HXXXXXX*  # Block-ID
```

### **Automatisches Cleanup via Retention**

Prometheus löscht automatisch alte Daten basierend auf:
- `--storage.tsdb.retention.time`: Zeitbasiert
- `--storage.tsdb.retention.size`: Speicherbasiert

**Welche Regel greift zuerst, wird angewendet!**

### **Backup-Strategie**

```bash
# Volume-Backup erstellen
docker run --rm \
  -v prometheus-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/prometheus-backup-$(date +%Y%m%d).tar.gz -C /data .

# Restore
docker run --rm \
  -v prometheus-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar xzf /backup/prometheus-backup-20251105.tar.gz -C /data
```

## 📋 Checkliste: Retention anpassen

- [ ] Retention-Zeit in `docker-compose.yml` hinzufügen
- [ ] Volume-Größe kalkulieren und provisionieren
- [ ] Backup-Strategie definieren
- [ ] Monitoring-Metriken für Speicherplatz einrichten
- [ ] Dokumentation für Team aktualisieren
- [ ] Alert bei 80% Speicher-Auslastung konfigurieren

## 🚀 Schnell-Setup

**Für 90 Tage Retention:**

```bash
# 1. docker-compose.yml anpassen (siehe oben)

# 2. Stack neu starten
docker compose down
docker compose up -d

# 3. Retention verifizieren
docker exec prometheus promtool tsdb analyze /prometheus

# 4. Speicherplatz überwachen
du -sh /var/lib/docker/volumes/cloud-performance-monitor_prometheus-data
```

## 📚 Weitere Informationen

- **Prometheus Storage**: https://prometheus.io/docs/prometheus/latest/storage/
- **Thanos Guide**: https://thanos.io/
- **Recording Rules**: `prometheus/recording_rules.yml`
- **TSDB Format**: https://ganeshvernekar.com/blog/prometheus-tsdb-the-head-block/
