# Wiki Setup for Cloud Performance Monitor

This directory contains the markdown files for the GitHub Wiki. To set up the wiki:

## 🚀 Quick Setup

### 1. Enable GitHub Wiki
1. Go to your GitHub repository settings
2. Scroll down to "Features" section
3. Check "Wikis" to enable the wiki feature

### 2. Clone Wiki Repository
```bash
# Clone the wiki repository (separate from main repo)
git clone https://github.com/MarcelWMeyer/cloud-performance-monitor.wiki.git
cd cloud-performance-monitor.wiki
```

### 3. Copy Wiki Files
```bash
# Copy all wiki files from docs/wiki/
cp ../cloud-performance-monitor/docs/wiki/*.md .
```

### 4. Push to Wiki
```bash
git add .
git commit -m "Add comprehensive runbook documentation"
git push origin master
```

## 📖 Wiki Structure

### Main Pages
- **Home.md** - Wiki homepage with navigation
- **Runbook-ServiceDown.md** - Critical service availability issues
- **Runbook-CriticalUploadDuration.md** - Performance degradation
- **Runbook-ServiceTestFailure.md** - Test failures and error codes

### Runbook Categories

#### 🚨 Critical Alerts
- Service Down
- Critical Upload Duration  
- Service Test Failure
- Critical Error Rate
- Critical Network Latency
- Circuit Breaker Open
- Critical SLA Violation

#### ⚠️ Warning Alerts
- High Upload Duration
- Slow Upload Speed
- High Error Rate
- High Network Latency
- Connection Timeouts
- Slow Chunk Uploads
- High Chunk Retry Rate
- SLA Violation 99%
- Too Many Alerts

## 🛠️ Runbook Template

Each runbook follows this structure:
1. **Alert Description** - What the alert means
2. **Alert Details** - Technical specifications
3. **Investigation Steps** - How to diagnose
4. **Common Causes & Solutions** - Known fixes
5. **Resolution Actions** - Step-by-step fixes
6. **Escalation Path** - When to escalate
7. **Related Metrics** - Monitoring data
8. **Related Runbooks** - Cross-references

## 📝 Creating New Runbooks

To add a new runbook:

1. Create new file: `Runbook-AlertName.md`
2. Follow the template structure
3. Add links to Home.md navigation
4. Update alert_rules.yml with runbook_url
5. Test the alert and runbook

## 🔗 Integration with Alerts

Alert rules in `prometheus/alert_rules.yml` now include:
```yaml
annotations:
  summary: "Alert description"
  description: "Detailed description with context"
  runbook_url: "https://github.com/MarcelWMeyer/cloud-performance-monitor/wiki/Runbook-AlertName"
```

This creates clickable links in:
- Email notifications
- Alertmanager UI
- Grafana annotations
- Third-party integrations

## 🎯 Benefits

- **Faster Resolution**: Step-by-step troubleshooting guides
- **Knowledge Sharing**: Documented solutions for common issues
- **Reduced Escalation**: Self-service problem resolution
- **Consistent Response**: Standardized troubleshooting approach
- **Training Tool**: New team members can learn from runbooks

## 📊 Maintenance

### Regular Updates
- Review runbooks monthly
- Update based on new incidents
- Add lessons learned from outages
- Keep contact information current

### Version Control
- All runbooks are version controlled
- Changes tracked through git history
- Collaborative editing through GitHub
- Easy rollback if needed