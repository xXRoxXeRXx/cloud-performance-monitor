# Email Configuration for Alertmanager

## Configuration

Add these variables to your `.env` file:

```bash
# SMTP Server
SMTP_SMARTHOST=smtp.gmail.com:587
SMTP_FROM=alerts@your-domain.com
SMTP_AUTH_USERNAME=alerts@your-domain.com
SMTP_AUTH_PASSWORD=your-app-password
SMTP_REQUIRE_TLS=true

# Alert Recipient
EMAIL_ADMIN=admin@your-domain.com
```

## Email Provider Examples

### Gmail
```bash
SMTP_SMARTHOST=smtp.gmail.com:587
SMTP_AUTH_USERNAME=your-email@gmail.com
SMTP_AUTH_PASSWORD=your-app-password  # Requires App Password!
SMTP_REQUIRE_TLS=true
```

> **Note**: Gmail requires an App Password. Create one at: https://myaccount.google.com/apppasswords

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

### Strato/Generic SMTP
```bash
SMTP_SMARTHOST=smtp.strato.de:465
SMTP_AUTH_USERNAME=your-email@strato.de
SMTP_AUTH_PASSWORD=your-password
SMTP_REQUIRE_TLS=true
```

## Alert Routing

All alerts are sent to `EMAIL_ADMIN` with these rules:

| Severity | Group Wait | Repeat Interval |
|----------|------------|-----------------|
| Critical | 10 seconds | 1 hour |
| Warning | 30 seconds | 4 hours |

## Testing

```bash
# Start the stack
make dev

# Check alertmanager logs
docker compose logs alertmanager

# View alert status
docker exec alertmanager wget -qO- http://localhost:9093/api/v1/status
```

## Troubleshooting

**Emails not arriving?**
1. Check alertmanager logs: `docker compose logs alertmanager`
2. Verify SMTP settings in `.env`
3. For Gmail: Ensure App Password is used (not regular password)
4. Check spam folder

**Connection refused?**
- Verify SMTP port (usually 587 for TLS, 465 for SSL)
- Check firewall settings
