# 🔒 Port Security Configuration

## 🌐 Exposed Ports (External Access)

| Port | Service | Purpose | Security Level |
|------|---------|---------|----------------|
| 3003 | Grafana | Main Dashboard UI | ✅ **Safe** - Login required |
| 8080 | Monitor Agent | Health checks & Metrics | ⚠️ **Limited** - Metrics only |

## 🔒 Internal Ports (Network-only Access)

| Port | Service | Purpose | Why Protected |
|------|---------|---------|---------------|
| 9090 | Prometheus | Metrics & Configuration | 🔒 Contains sensitive queries |
| 9093 | Alertmanager | Alert Configuration | 🔒 **Contains E-mail addresses** |
| 8080 | Webhook Logger | Debug Logs | 🔒 May contain sensitive data |

## 🛠️ Debug Access (When Needed)

### Temporary Port Exposure for Debugging

If you need to access internal services for debugging, temporarily modify `docker-compose.yml`:

```yaml
# ONLY FOR DEBUGGING - Remove after use!
  prometheus:
    ports:
      - "9090:9090"  # Add this line temporarily
    # ... rest of config

  alertmanager:
    ports:
      - "9093:9093"  # Add this line temporarily
    # ... rest of config
```

**⚠️ Important Security Notes:**
1. **Never expose these ports in production**
2. **Remove port mappings after debugging**
3. **E-mail addresses visible at http://localhost:9093/#/status**
4. **Sensitive metrics visible at http://localhost:9090**

### Safe Debugging Alternatives

```bash
# Access logs instead of web interfaces
docker compose logs prometheus
docker compose logs alertmanager

# Execute commands inside containers
docker compose exec prometheus promtool query instant 'up'
docker compose exec alertmanager amtool config show
```

## 🔐 Security Benefits

✅ **E-mail addresses protected** - No external access to Alertmanager config  
✅ **Metrics secured** - No external access to sensitive performance data  
✅ **Reduced attack surface** - Only necessary ports exposed  
✅ **Internal communication** - Services communicate via Docker network  

## 🚀 Production Deployment

This configuration is production-ready with minimal exposed ports for maximum security.