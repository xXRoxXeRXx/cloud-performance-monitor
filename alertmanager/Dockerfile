FROM prom/alertmanager:v0.27.0

# Install gettext for envsubst command
USER root

# Check which Linux distribution and install gettext accordingly
RUN if command -v apk > /dev/null; then \
        apk add --no-cache gettext; \
    elif command -v apt-get > /dev/null; then \
        apt-get update && apt-get install -y gettext-base && rm -rf /var/lib/apt/lists/*; \
    elif command -v yum > /dev/null; then \
        yum install -y gettext && yum clean all; \
    else \
        echo "Package manager not found, trying to install gettext manually"; \
    fi

# Copy configuration template and startup script
COPY alertmanager.yml.template /etc/alertmanager/alertmanager.yml.template
COPY start-alertmanager.sh /start-alertmanager.sh

# Make script executable
RUN chmod +x /start-alertmanager.sh

# Use our startup script as entrypoint (runs as root to handle file permissions)
ENTRYPOINT ["/start-alertmanager.sh"]
