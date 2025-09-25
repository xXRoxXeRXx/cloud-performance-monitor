# Alert Error Codes Reference

Diese Dokumentation beschreibt alle Error Codes, die in den Alert-E-Mails erscheinen können und ihre Bedeutung.

## HTTP Status Error Codes

| Error Code | Bedeutung | Aktion |
|------------|-----------|---------|
| `http_404_not_found` | Datei oder Endpoint nicht gefunden | Prüfe URL und Pfade |
| `http_401_unauthorized` | Authentifizierung fehlgeschlagen | Prüfe Credentials |
| `http_403_forbidden` | Zugriff verweigert | Prüfe Berechtigungen |
| `http_500_internal_error` | Server-interner Fehler | Kontaktiere Admin |
| `http_502_bad_gateway` | Gateway-Fehler | Netzwerk/Proxy-Problem |
| `http_503_service_unavailable` | Service nicht verfügbar | Service ist down |
| `http_504_gateway_timeout` | Gateway-Timeout | Netzwerk/Performance-Problem |
| `http_507_insufficient_storage` | Nicht genügend Speicherplatz | Speicher freigeben |

## Network Error Codes

| Error Code | Bedeutung | Aktion |
|------------|-----------|---------|
| `connection_refused` | Verbindung abgelehnt | Service/Port prüfen |
| `network_timeout` | Netzwerk-Timeout | Netzwerk-Latenz prüfen |
| `dns_resolution_failed` | DNS-Auflösung fehlgeschlagen | DNS-Konfiguration prüfen |
| `tls_handshake_failed` | SSL/TLS-Fehler | Zertifikate prüfen |

## Operation Error Codes

| Error Code | Bedeutung | Aktion |
|------------|-----------|---------|
| `upload_failed` | Upload fehlgeschlagen | Upload-Prozess prüfen |
| `download_failed` | Download fehlgeschlagen | Download-Prozess prüfen |
| `delete_failed` | Löschvorgang fehlgeschlagen | Berechtigungen prüfen |
| `size_mismatch` | Dateigröße stimmt nicht überein | Übertragung unterbrochen |
| `chunk_assembly_failed` | Chunk-Zusammenfügung fehlgeschlagen | WebDAV-Konfiguration prüfen |

## WebDAV Specific Error Codes

| Error Code | Bedeutung | Aktion |
|------------|-----------|---------|
| `webdav_error` | Allgemeiner WebDAV-Fehler | WebDAV-Logs prüfen |
| `propfind_failed` | PROPFIND-Request fehlgeschlagen | WebDAV-Zugriff prüfen |
| `mkcol_failed` | Verzeichnis-Erstellung fehlgeschlagen | Berechtigungen prüfen |

## Authentication Error Codes

| Error Code | Bedeutung | Aktion |
|------------|-----------|---------|
| `auth_failed` | Authentifizierung fehlgeschlagen | Credentials prüfen |
| `token_expired` | OAuth-Token abgelaufen | Token erneuern |
| `token_refresh_failed` | Token-Erneuerung fehlgeschlagen | OAuth-Konfiguration prüfen |

## Special Error Codes

| Error Code | Bedeutung | Aktion |
|------------|-----------|---------|
| `none` | Kein Fehler aufgetreten | Normal - Erfolgreicher Test |
| `unknown_error` | Unbekannter Fehler | Logs detailliert prüfen |
| `sla_violation` | SLA-Verletzung | Performance analysieren |
| `circuit_breaker_open` | Circuit Breaker geöffnet | Service-Health prüfen |

## Error Code Priorität in Alerts

1. **HTTP Status Codes** - Höchste Priorität (spezifische Server-Responses)
2. **Network Error Codes** - Hohe Priorität (Infrastruktur-Probleme)
3. **Operation Error Codes** - Mittlere Priorität (Anwendungs-Logik)
4. **Generic Error Codes** - Niedrigste Priorität (Fallback)

## Troubleshooting Workflow

1. **Error Code identifizieren** - Aus der Alert-E-Mail
2. **Kategorie bestimmen** - HTTP/Network/Operation/Auth
3. **Runbook konsultieren** - Spezifische Aktionen ausführen
4. **Grafana Dashboard prüfen** - Trend-Analyse durchführen
5. **Service Logs prüfen** - Detaillierte Fehler-Analyse

## Alert E-Mail Format

```
🔍 Error Analysis:
• Error Code: http_504_gateway_timeout
• Test Type: upload
• Category: network

🌐 Common Network Error Codes:
• http_504_timeout: Gateway timeout
• connection_refused: Service unavailable
• network_timeout: Network connectivity issue
```

## Integration mit Grafana

Alle Error Codes werden auch im Grafana Error Dashboard angezeigt:
- **URL**: http://localhost:3003/d/cloud-error-analysis/
- **Filter**: Nach Service und Error Code filtern
- **Trends**: Zeitreihen-Analyse der Fehler-Patterns

## Weitere Dokumentation

- [GitHub Wiki Runbooks](https://github.com/xXRoxXeRXx/cloud-performance-monitor/wiki/)
- [Error Code Reference](https://github.com/xXRoxXeRXx/cloud-performance-monitor/wiki/Error-Code-Reference)
- [Grafana Dashboard Guide](./GRAFANA_DASHBOARDS.md)
