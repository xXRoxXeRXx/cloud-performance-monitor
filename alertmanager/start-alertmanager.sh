#!/bin/sh

# Script to substitute environment variables in alertmanager.yml template
# and start alertmanager

echo "Starting Alertmanager with environment variable substitution..."

# Check if required environment variables are set
required_vars="SMTP_SMARTHOST SMTP_FROM SMTP_AUTH_USERNAME SMTP_AUTH_PASSWORD EMAIL_ADMIN EMAIL_DEVOPS EMAIL_NETWORK EMAIL_MANAGEMENT"

for var in $required_vars; do
    eval value=\$$var
    if [ -z "$value" ]; then
        echo "Warning: Environment variable $var is not set"
    else
        echo "✓ $var is configured"
    fi
done

echo "Configuration file generated:"
echo "SMTP Server: $SMTP_SMARTHOST"
echo "From Address: $SMTP_FROM"
echo "Admin Email: $EMAIL_ADMIN"
echo "DevOps Email: $EMAIL_DEVOPS"
echo "Network Email: $EMAIL_NETWORK"
echo "Management Email: $EMAIL_MANAGEMENT"

# Manual substitution since envsubst might not be available
sed -e "s/\${SMTP_SMARTHOST}/$SMTP_SMARTHOST/g" \
    -e "s/\${SMTP_FROM}/$SMTP_FROM/g" \
    -e "s/\${SMTP_AUTH_USERNAME}/$SMTP_AUTH_USERNAME/g" \
    -e "s/\${SMTP_AUTH_PASSWORD}/$SMTP_AUTH_PASSWORD/g" \
    -e "s/\${SMTP_REQUIRE_TLS}/$SMTP_REQUIRE_TLS/g" \
    -e "s/\${EMAIL_ADMIN}/$EMAIL_ADMIN/g" \
    -e "s/\${EMAIL_DEVOPS}/$EMAIL_DEVOPS/g" \
    -e "s/\${EMAIL_NETWORK}/$EMAIL_NETWORK/g" \
    -e "s/\${EMAIL_MANAGEMENT}/$EMAIL_MANAGEMENT/g" \
    /etc/alertmanager/alertmanager.yml.template > /etc/alertmanager/alertmanager.yml

echo "✓ Configuration file created"

# Start alertmanager
echo "Starting Alertmanager..."
exec /bin/alertmanager --config.file=/etc/alertmanager/alertmanager.yml --storage.path=/alertmanager --web.external-url=http://localhost:9093
