FROM grafana/grafana:11.0.0

# Provisioning für Datasource und Dashboard
COPY deploy/grafana/provisioning/datasources /etc/grafana/provisioning/datasources
COPY deploy/grafana/provisioning/dashboards /etc/grafana/provisioning/dashboards
COPY deploy/grafana/dashboard.json /etc/grafana/provisioning/dashboards/dashboard.json
COPY deploy/grafana/dashboard-enhanced.json /etc/grafana/provisioning/dashboards/dashboard-enhanced.json
