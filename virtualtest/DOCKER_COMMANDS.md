# Docker Compose Commands - Quick Reference

This file lists all commands for managing the test environment with Docker Compose.

## 🚀 Starting

### Start All Services

```bash
cd virtualtest
docker-compose up -d
```

This starts:
- ✅ Mock Node 1 (Port 8080)
- ✅ Mock Node 2 (Port 8081)
- ✅ Mock Registry (Port 8082)
- ✅ Mock IS-07 (Port 8083)
- ✅ Mock IS-08 (Port 8084)

### Start in Foreground (to see logs)

```bash
docker-compose up
```

### Start Specific Services

```bash
# Only mock-node-1 and mock-node-2
docker-compose up -d mock-node-1 mock-node-2

# Only registry
docker-compose up -d mock-registry
```

## 📊 Status

### Show Service Status

```bash
docker-compose ps
```

Example output:
```
NAME              STATUS          PORTS
mock-node-1       Up (healthy)    0.0.0.0:8080->8080/tcp
mock-node-2       Up (healthy)    0.0.0.0:8081->8081/tcp
mock-registry     Up (healthy)    0.0.0.0:8082->8082/tcp
mock-is07         Up (healthy)    0.0.0.0:8083->8083/tcp
mock-is08         Up (healthy)    0.0.0.0:8084->8084/tcp
```

### Check Service Health

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
```

### Check All Services in One Command

```bash
for port in 8080 8081 8082 8083 8084; do
  echo -n "Port $port: "
  curl -s http://localhost:$port/health > /dev/null && echo "✓ OK" || echo "✗ FAILED"
done
```

## 📝 Logs

### Follow All Service Logs

```bash
docker-compose logs -f
```

### Follow a Specific Service

```bash
docker-compose logs -f mock-node-1
docker-compose logs -f mock-node-2
docker-compose logs -f mock-registry
```

### Show Last N Lines

```bash
docker-compose logs --tail=50 mock-node-1
```

### Logs for a Time Range

```bash
docker-compose logs --since 10m mock-node-1
docker-compose logs --since 2024-01-01T00:00:00 mock-node-1
```

## 🛑 Stopping

### Stop All and Remove Containers

```bash
docker-compose down
```

### Stop Only (keep containers)

```bash
docker-compose stop
```

### Stop a Specific Service

```bash
docker-compose stop mock-node-1
```

## 🔄 Restart

### Restart All Services

```bash
docker-compose restart
```

### Restart a Specific Service

```bash
docker-compose restart mock-node-1
```

### Stop, Remove and Start Again

```bash
docker-compose down
docker-compose up -d
```

## 🔧 Management

### Rebuild Services

```bash
docker-compose build
docker-compose build mock-node-1
docker-compose build --no-cache
```

### Access Containers

```bash
docker-compose exec mock-node-1 /bin/bash
docker-compose exec mock-node-1 python -c "print('Hello')"
docker-compose exec mock-node-1 sh
```

### Container Info

```bash
docker-compose ps -a
docker inspect mock-node-1
docker-compose top mock-node-1
```

### Resource Usage

```bash
docker stats
docker stats mock-node-1 mock-node-2
```

## 🧹 Cleanup

### Stop and Remove All Containers

```bash
docker-compose down
```

### Remove Containers and Volumes

```bash
docker-compose down -v
```

### Remove Containers, Volumes and Images

```bash
docker-compose down -v --rmi all
```

### Prune Unused Images

```bash
docker image prune -a
```

## 🐛 Troubleshooting

### Services Not Starting

```bash
docker-compose logs
docker-compose build --no-cache
docker-compose down
docker-compose up -d
```

### Port Conflict

```bash
lsof -i :8080
lsof -i :8081
lsof -i :8082
# If a port is in use, change port numbers in docker-compose.yml
```

### Health Check Failing

```bash
docker-compose logs mock-node-1
docker-compose exec mock-node-1 python -c "import urllib.request; print(urllib.request.urlopen('http://localhost:8080/health').read())"
```

### Network Issues

```bash
docker network ls
docker network inspect virtualtest_nmos-test-network
docker-compose down
docker network prune
docker-compose up -d
```

## 📋 Useful Command Combinations

### Start and Follow Logs

```bash
docker-compose up -d && docker-compose logs -f
```

### Status and Health Check

```bash
docker-compose ps && echo "" && for port in 8080 8081 8082; do curl -s http://localhost:$port/health && echo ""; done
```

### Restart and Follow Logs

```bash
docker-compose restart && docker-compose logs -f
```

### Clean Start

```bash
docker-compose down && docker-compose build --no-cache && docker-compose up -d
```

## ✅ Quick Check Script

```bash
#!/bin/bash
# check-services.sh

cd virtualtest

echo "📊 Docker Compose status:"
docker-compose ps

echo ""
echo "🏥 Health checks:"
for port in 8080 8081 8082 8083 8084; do
  name=$(docker-compose ps | grep ":$port" | awk '{print $1}' | head -1)
  status=$(curl -s http://localhost:$port/health 2>/dev/null && echo "✓" || echo "✗")
  echo "  $name (Port $port): $status"
done
```

## 🎯 Example Scenarios

### Scenario 1: First-time start

```bash
cd virtualtest
docker-compose up -d
sleep 5
docker-compose ps
```

### Scenario 2: Start with logs

```bash
cd virtualtest
docker-compose up
```

### Scenario 3: Start only nodes

```bash
cd virtualtest
docker-compose up -d mock-node-1 mock-node-2
```

### Scenario 4: Restart services

```bash
cd virtualtest
docker-compose restart
```

### Scenario 5: Full clean and restart

```bash
cd virtualtest
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```
