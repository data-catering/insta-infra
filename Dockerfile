FROM docker:dind

LABEL maintainer="Peter Flook <peter.flook@data.catering>"
LABEL description="Insta-Infra: Spin up any service straight away on your local laptop"

# Install dependencies
RUN apk add --no-cache bash git curl jq

# Set up working directory
WORKDIR /app

# Copy project files
COPY . /app/

# Make script executable
RUN chmod +x /app/run.sh

# Create a symbolic link to /usr/local/bin
RUN ln -sf /app/run.sh /usr/local/bin/insta

# Set environment variables
ENV DOCKER_HOST=unix:///var/run/docker.sock

# Entrypoint
ENTRYPOINT ["/app/run.sh"]

# Default command
CMD ["help"] 