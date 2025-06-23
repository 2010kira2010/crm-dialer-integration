#!/bin/bash

# Production deployment script

set -e

# Configuration
ENVIRONMENT=${1:-staging}
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env.${ENVIRONMENT}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

check_requirements() {
    print_header "Checking requirements"

    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}✗ Docker is not installed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker found${NC}"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}✗ Docker Compose is not installed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker Compose found${NC}"

    # Check environment file
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${RED}✗ Environment file $ENV_FILE not found${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Environment file found${NC}"
}

backup_database() {
    print_header "Backing up database"

    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    BACKUP_FILE="backups/db_backup_${ENVIRONMENT}_${TIMESTAMP}.sql"

    echo "Creating backup: $BACKUP_FILE"
    docker-compose exec -T postgres pg_dump -U postgres crm_dialer > "$BACKUP_FILE"

    if [ -f "$BACKUP_FILE" ]; then
        echo -e "${GREEN}✓ Database backed up successfully${NC}"
    else
        echo -e "${RED}✗ Database backup failed${NC}"
        exit 1
    fi
}

pull_images() {
    print_header "Pulling latest images"

    docker-compose pull

    echo -e "${GREEN}✓ Images updated${NC}"
}

deploy_services() {
    print_header "Deploying services"

    # Stop services
    echo "Stopping services..."
    docker-compose down

    # Start services with new images
    echo "Starting services..."
    docker-compose --env-file "$ENV_FILE" up -d

    # Wait for services to be healthy
    echo "Waiting for services to be healthy..."
    sleep 10

    # Check service status
    docker-compose ps
}

run_migrations() {
    print_header "Running database migrations"

    for migration in migrations/*.sql; do
        if [ -f "$migration" ]; then
            echo "Applying migration: $(basename $migration)"
            docker-compose exec -T postgres psql -U postgres -d crm_dialer < "$migration"
        fi
    done

    echo -e "${GREEN}✓ Migrations completed${NC}"
}

health_check() {
    print_header "Running health checks"

    # Check API Gateway
    echo -n "Checking API Gateway... "
    if curl -f -s http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi

    # Check Prometheus
    echo -n "Checking Prometheus... "
    if curl -f -s http://localhost:9090/-/healthy > /dev/null; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${YELLOW}⚠ Warning${NC}"
    fi

    # Check Grafana
    echo -n "Checking Grafana... "
    if curl -f -s http://localhost:3001/api/health > /dev/null; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${YELLOW}⚠ Warning${NC}"
    fi
}

cleanup_old_images() {
    print_header "Cleaning up old images"

    docker image prune -f

    echo -e "${GREEN}✓ Cleanup completed${NC}"
}

send_notification() {
    print_header "Sending deployment notification"

    # You can integrate with Slack, Discord, or email here
    echo "Deployment completed for environment: $ENVIRONMENT"
    echo "Timestamp: $(date)"
}

# Main deployment flow
main() {
    print_header "CRM-Dialer Integration Deployment"
    echo "Environment: $ENVIRONMENT"
    echo "Start time: $(date)"

    check_requirements

    if [ "$ENVIRONMENT" == "production" ]; then
        backup_database
    fi

    pull_images
    deploy_services
    run_migrations
    health_check
    cleanup_old_images
    send_notification

    print_header "Deployment completed successfully!"
    echo "End time: $(date)"
}

# Run deployment
main