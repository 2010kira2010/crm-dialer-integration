#!/bin/bash

echo "=== Docker Containers Status ==="
docker-compose ps

echo -e "\n=== Port 8080 Check ==="
echo "Checking if port 8080 is listening..."
lsof -i :8080 || echo "Port 8080 is not in use"

echo -e "\n=== API Gateway Logs ==="
docker-compose logs --tail=50 api-gateway

echo -e "\n=== Testing API Gateway Health ==="
curl -v http://localhost:8080/health || echo "API Gateway health check failed"

echo -e "\n=== Checking if frontend files exist in container ==="
docker-compose exec api-gateway ls -la /root/web/build || echo "Frontend build not found in container"

echo -e "\n=== Network connectivity check ==="
docker network ls
docker network inspect crm-dialer-integration_crm-dialer-network

echo -e "\n=== Container inspection ==="
docker inspect crm-dialer-api-gateway | grep -A 5 "Ports"