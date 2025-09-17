# E-Mail-Konfiguration für Alertmanager

## Übersicht

Der Alertmanager wurde erfolgreich mit E-Mail-Benachrichtigungen konfiguriert. Alle E-Mail-Einstellungen werden über Umgebungsvariablen aus der `.env` Datei gesteuert.

## Konfiguration

### Umgebungsvariablen (.env)

```bash
# SMTP-Server Konfiguration
SMTP_SMARTHOST=smtp.example.com:465
SMTP_FROM=admin@example.com
SMTP_AUTH_USERNAME=admin@example.com
SMTP_AUTH_PASSWORD=your-password
SMTP_REQUIRE_TLS=true

# E-Mail-Empfänger für verschiedene Alert-Typen
EMAIL_ADMIN=admin@example.com        # Kritische Alerts
EMAIL_DEVOPS=devops@example.com       # Performance Alerts
EMAIL_NETWORK=network@example.com      # Netzwerk Alerts
EMAIL_MANAGEMENT=management@example.com   # SLA Verletzungen
```

### Alert-Routing

**1. Kritische Alerts (`severity: critical`)**
- Empfänger: `EMAIL_ADMIN`
- Intervall: Sofortige Benachrichtigung (10s)
- Wiederholung: Alle 5 Minuten
- Kanäle: E-Mail + Webhook

**2. Performance Alerts (`category: performance`)**
- Empfänger: `EMAIL_DEVOPS`
- Intervall: Gebündelt alle 10 Minuten
- Wiederholung: Alle 4 Stunden
- Kanäle: E-Mail + Webhook

**3. Netzwerk Alerts (`category: network`)**
- Empfänger: `EMAIL_NETWORK`
- Intervall: Gebündelt alle 15 Minuten
- Wiederholung: Alle 6 Stunden
- Kanäle: E-Mail + Webhook

**4. SLA Verletzungen (`category: sla`)**
- Empfänger: `EMAIL_MANAGEMENT` + `EMAIL_DEVOPS`
- Intervall: 5 Minuten Wartezeit
- Wiederholung: Alle 2 Stunden
- Kanäle: E-Mail + Webhook

## E-Mail Provider Beispiele

### Gmail
```bash
SMTP_SMARTHOST=smtp.gmail.com:587
SMTP_AUTH_USERNAME=your-email@gmail.com
SMTP_AUTH_PASSWORD=your-app-password  # App-Passwort erforderlich!
SMTP_REQUIRE_TLS=true
```

### Outlook/Hotmail
```bash
SMTP_SMARTHOST=smtp-mail.outlook.com:587
SMTP_AUTH_USERNAME=your-email@outlook.com
SMTP_AUTH_PASSWORD=your-password
SMTP_REQUIRE_TLS=true
```

### Yahoo
```bash
SMTP_SMARTHOST=smtp.mail.yahoo.com:587
SMTP_AUTH_USERNAME=your-email@yahoo.com
SMTP_AUTH_PASSWORD=your-app-password
SMTP_REQUIRE_TLS=true
```

### Strato (aktuell konfiguriert)
```bash
SMTP_SMARTHOST=smtp.example.com:465
SMTP_AUTH_USERNAME=your-email@your-domain.com
SMTP_AUTH_PASSWORD=your-password
SMTP_REQUIRE_TLS=true
```

## E-Mail-Vorlagen

### Kritische Alerts
```
Subject: 🚨 CRITICAL Alert - Nextcloud Performance Monitor

🚨 CRITICAL ALERT DETECTED 🚨

Alert: Service Down
Description: Nextcloud instance is not responding
Instance: cloud.example.com
Service: nextcloud
Severity: critical

Time: 2025-09-13 21:30:00

Please take immediate action to resolve this issue.
```

### Performance Alerts
```
Subject: ⚠️ Performance Alert - Nextcloud Performance Monitor

⚠️ PERFORMANCE ALERT ⚠️

Alert: High Upload Duration
Description: Upload takes longer than 30 seconds
Instance: cloud.example.com
Service: nextcloud
Time: 2025-09-13 21:30:00
```

### Netzwerk Alerts
```
Subject: 🌐 Network Alert - Nextcloud Performance Monitor

🌐 NETWORK ALERT 🌐

Alert: High Network Latency
Description: Network latency exceeds 1000ms
Instance: cloud.example.com
Time: 2025-09-13 21:30:00

Current Latency: 1500ms
```

### SLA Verletzungen
```
Subject: 📊 SLA Violation - Nextcloud Performance Monitor

📊 SLA VIOLATION DETECTED 📊

SLA Violation for Service: nextcloud

Alert: Service Availability Below SLA
Description: Service uptime is below 99%
Service: nextcloud
Time: 2025-09-13 21:30:00

Please review the service performance and take corrective action.
```

## Testen der E-Mail-Konfiguration

### 1. Container-Logs prüfen
```bash
docker compose logs alertmanager
```

### 2. Alertmanager-UI verwenden
- URL: http://localhost:9093
- Überprüfung der Receiver-Konfiguration
- Status der E-Mail-Verbindung

### 3. Test-Alert senden
```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "critical",
      "instance": "test.example.com",
      "service": "test"
    },
    "annotations": {
      "summary": "Test Alert for Email Configuration",
      "description": "This is a test alert to verify email notifications"
    }
  }]'
```

## Troubleshooting

### SMTP-Authentifizierung fehlgeschlagen
- Überprüfen Sie Benutzername und Passwort
- Für Gmail: App-Passwort verwenden
- Für 2FA-Konten: App-spezifische Passwörter erforderlich

### TLS-Verbindungsfehler
- Port 587 für STARTTLS
- Port 465 für SSL/TLS
- `SMTP_REQUIRE_TLS=true` für verschlüsselte Verbindungen

### E-Mails werden nicht gesendet
1. Container-Logs prüfen: `docker compose logs alertmanager`
2. Firewall-Regeln für SMTP-Ports überprüfen
3. SMTP-Server-Erreichbarkeit testen: `telnet smtp.server.com 587`

### Konfigurationsfehler
```bash
# Konfiguration validieren
docker exec alertmanager cat /etc/alertmanager/alertmanager.yml

# Neustart nach Änderungen
docker compose restart alertmanager
```

## Sicherheitshinweise

1. **Passwort-Schutz**: Niemals Passwörter in Git committen
2. **App-Passwörter**: Für Gmail und andere 2FA-Konten verwenden
3. **TLS-Verschlüsselung**: Immer `SMTP_REQUIRE_TLS=true` verwenden
4. **Berechtigungen**: E-Mail-Konten nur mit notwendigen Berechtigungen

## Monitoring

- **Webhook-Logger**: Alle Alerts werden zusätzlich an `http://webhook-logger:8080` gesendet
- **Prometheus**: Alert-Metriken in Prometheus verfügbar
- **Grafana**: Dashboard für Alert-Statistiken

## Weiterführende Konfiguration

### Slack-Integration (optional)
Für Slack-Benachrichtigungen können Sie die auskommentierte Slack-Konfiguration in `alertmanager.yml.template` aktivieren.

### Teams-Integration (optional)
Microsoft Teams-Webhooks können als zusätzliche Webhook-URLs konfiguriert werden.

### PagerDuty-Integration (optional)
Für 24/7-Support können PagerDuty-Integrationen hinzugefügt werden.
