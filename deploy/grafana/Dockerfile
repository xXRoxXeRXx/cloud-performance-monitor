FROM grafana/grafana:11.0.0

# Provisioning für Datasource und Dashboard
COPY provisioning/datasources /etc/grafana/provisioning/datasources
COPY provisioning/dashboards /etc/grafana/provisioning/dashboards
COPY *.json /etc/grafana/provisioning/dashboards/
