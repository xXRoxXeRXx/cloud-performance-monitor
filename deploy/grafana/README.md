# Grafana Dashboards - Cloud Performance Monitor

This directory contains all Grafana dashboards for the Cloud Performance Monitor system.

## 📊 Available Dashboards

### 1. Main Performance Dashboard (`dashboard.json`)
**Purpose**: Overall system performance monitoring
- Service availability and uptime
- Upload/download performance metrics
- Network latency monitoring
- Test duration trends
- Service comparison views

**Key Panels**:
- Success rate by service
- Performance trends over time
- Test duration histograms
- Speed comparisons
- Network latency graphs

### 2. Enhanced Analytics Dashboard (`enhanced-dashboard.json`)
**Purpose**: Deep-dive analytics and historical trends
- Advanced performance analytics
- Historical data analysis
- Trend forecasting
- Capacity planning metrics
- Detailed service breakdowns

**Key Panels**:
- Advanced trend analysis
- Performance predictions
- Resource utilization
- Historical comparisons
- Statistical analysis

### 3. Error Analysis Dashboard (`error-dashboard.json`) 🆕
**Purpose**: Comprehensive error monitoring and troubleshooting
- Real-time error tracking
- Error categorization and trends
- Troubleshooting guidance
- Success rate monitoring
- Error code analysis

**Key Features**:
- 🚨 **Current Status Overview**: Active errors and success rates
- 📊 **Error Trends**: Historical error patterns and success rate trends
- 🔧 **Detailed Analysis**: Specific error codes with troubleshooting links
- 🏷️ **Error Categories**: HTTP errors, network issues, auth problems, etc.
- 📈 **Success Rate Monitoring**: Real-time and historical success rates

**Error Categories Tracked**:
- **HTTP 5xx Errors**: Server-side issues (red)
- **HTTP 4xx Errors**: Client-side issues (orange)
- **Network Errors**: Connectivity problems (purple)
- **Authentication Errors**: Auth/token issues (yellow)
- **File Operation Errors**: Upload/download problems (blue)
- **WebDAV Errors**: Protocol-specific issues (green)

### 4. Performance Dashboard (`performance-dashboard.json`)
**Purpose**: Focused performance metrics
- Speed and latency monitoring
- Performance benchmarking
- SLA monitoring
- Response time analysis

## 🔧 Error Dashboard Features

### Interactive Elements
- **Service Filter**: Select specific cloud services
- **Instance Filter**: Filter by instance URL
- **Error Code Filter**: Focus on specific error types
- **Time Range**: Adjust monitoring period

### Visual Indicators
- **Color Coding**: Different colors for error categories
- **Gauges**: Success rate visualization
- **Heatmaps**: Error distribution patterns
- **Trends**: Historical error rate analysis

### Troubleshooting Integration
- **Direct Links**: Click error codes to access troubleshooting guides
- **Runbook Integration**: Links to GitHub Wiki runbooks
- **Error Code Reference**: Complete error code documentation

## 📚 Documentation Links

The Error Dashboard includes direct links to:
- [Error Code Reference](https://github.com/MarcelWMeyer/cloud-performance-monitor/wiki/Error-Code-Reference)
- [Service Test Failure Runbook](https://github.com/MarcelWMeyer/cloud-performance-monitor/wiki/Runbook-ServiceTestFailure)
- [Complete Wiki Documentation](https://github.com/MarcelWMeyer/cloud-performance-monitor/wiki)

## 🔍 Key Metrics Monitored

### Error Tracking
```promql
# Total active errors
sum(cloud_test_errors_total{error_code!="none"})

# Success rate
(sum(cloud_test_success{error_code="none"}) / sum(cloud_test_success)) * 100

# Error rate by category
rate(cloud_test_errors_total{error_code=~"http_5.*"}[5m])
```

### Performance Metrics
```promql
# Upload/download duration
cloud_test_duration_seconds

# Transfer speeds
cloud_test_speed_mbytes_per_sec

# Network latency
cloud_network_latency_seconds
```

## 🚀 Setup and Usage

### Automatic Import
All dashboards are automatically imported when starting Grafana:
```bash
docker compose up -d
```

### Manual Import
1. Access Grafana at http://localhost:3000
2. Go to Dashboards → Import
3. Upload JSON file or paste JSON content
4. Configure data source (Prometheus)

### Dashboard Navigation
- **Main Dashboard**: Overview of all services
- **Error Dashboard**: Focus on error analysis and troubleshooting
- **Enhanced Dashboard**: Deep analytics and trends
- **Performance Dashboard**: Speed and latency focus

## 🎯 Best Practices

### Error Monitoring
1. **Start with Error Dashboard** for quick issue identification
2. **Use filters** to focus on specific services or error types
3. **Follow links** to troubleshooting guides for resolution steps
4. **Monitor trends** to identify recurring issues

### Performance Analysis
1. **Compare services** using main dashboard
2. **Analyze trends** with enhanced dashboard
3. **Set up alerts** based on error thresholds
4. **Regular review** of error patterns and success rates

### Troubleshooting Workflow
1. **Identify errors** in Error Dashboard
2. **Click error codes** for specific troubleshooting steps
3. **Check service logs** using provided commands
4. **Follow runbook procedures** for resolution
5. **Monitor recovery** in real-time

## 🔄 Dashboard Updates

Dashboards are version-controlled and automatically updated:
- Updates deployed via Docker Compose restart
- Configuration preserved across updates
- Custom panels can be added without conflicts

## 📊 Sample Queries

### Error Analysis
```promql
# Top error codes
topk(10, sum by (error_code) (cloud_test_errors_total{error_code!="none"}))

# Service health overview
(sum by (service) (cloud_test_success{error_code="none"}) / sum by (service) (cloud_test_success)) * 100

# Error trend analysis
rate(cloud_test_errors_total{error_code!="none"}[5m]) * 300
```

### Success Monitoring
```promql
# Overall system success rate
(sum(cloud_test_success{error_code="none"}) / sum(cloud_test_success)) * 100

# Service comparison
sum by (service, instance) (cloud_test_success{error_code="none"}) / sum by (service, instance) (cloud_test_success) * 100
```

For more detailed information, see the [complete documentation](https://github.com/MarcelWMeyer/cloud-performance-monitor/wiki).