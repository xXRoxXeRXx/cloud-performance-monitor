# 🔒 Port Security Configuration

## 🌐 Exposed Ports (External Access)

| Port | Service | Purpose | Security Level |
|------|---------|---------|----------------|
| 3003 | Grafana | Main Dashboard UI | ✅ **Safe** - Login required |

**🔒 All metrics and internal services are network-isolated for maximum security.**

## 🔒 Internal Ports (Network-only Access)

| Port | Service | Purpose | Why Protected |
|------|---------|---------|---------------|
| 8080 | Monitor Agent | Metrics & Health checks | 🔒 No external access needed |
| 9090 | Prometheus | Metrics & Configuration | 🔒 Contains sensitive queries |
| 9093 | Alertmanager | Alert Configuration | 🔒 **Contains E-mail addresses** |
| 8080 | Webhook Logger | Debug Logs | 🔒 May contain sensitive data |

## 🛠️ Debug Access (When Needed)

### Temporary Port Exposure for Debugging

If you need to access internal services for debugging, temporarily modify `docker-compose.yml`:

```yaml
# ONLY FOR DEBUGGING - Remove after use!
  monitor-agent:
    ports:
      - "8080:8080"  # Add this line temporarily
    # ... rest of config

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
3. **Monitor Agent metrics only needed by Prometheus internally**
4. **E-mail addresses visible at http://localhost:9093/#/status**
5. **Sensitive metrics visible at http://localhost:9090**

### Safe Debugging Alternatives

```bash
# Access logs instead of web interfaces
docker compose logs monitor-agent
docker compose logs prometheus
docker compose logs alertmanager

# Execute commands inside containers
docker compose exec monitor-agent curl http://localhost:8080/metrics
docker compose exec prometheus promtool query instant 'up'
docker compose exec alertmanager amtool config show
```

## 🔐 Security Benefits

✅ **Complete isolation** - Only Grafana accessible externally  
✅ **Metrics secured** - Monitor Agent only accessible internally by Prometheus  
✅ **E-mail addresses protected** - No external access to Alertmanager config  
✅ **Zero external attack surface** - All monitoring infrastructure internal  
✅ **Internal communication** - Services communicate via Docker network  

## 🚀 Production Deployment

This configuration is production-ready with minimal exposed ports for maximum security.