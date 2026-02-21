# High Availability Deployment Guide

This guide documents how to deploy GO-NMOS in High Availability (HA) mode for production environments.

## Overview

GO-NMOS can be deployed in HA mode to ensure:
- **Zero-downtime deployments** - Rolling updates without service interruption
- **Fault tolerance** - Automatic failover when components fail
- **Scalability** - Handle increased load by scaling horizontally
- **Data redundancy** - Database replication ensures data availability

## Architecture Components

### 1. Database High Availability

#### PostgreSQL Streaming Replication

**Primary-Replica Setup:**

1. **Primary Database Configuration:**
```bash
# postgresql.conf (Primary)
wal_level = replica
max_wal_senders = 3
max_replication_slots = 3
hot_standby = on

# pg_hba.conf (Primary)
host    replication    replicator    192.168.1.0/24    md5
```

2. **Replica Database Configuration:**
```bash
# postgresql.conf (Replica)
hot_standby = on
```

3. **Create Replication User:**
```sql
CREATE USER replicator WITH REPLICATION PASSWORD 'secure_password';
```

4. **Initialize Replica:**
```bash
# On replica server
pg_basebackup -h primary-host -D /var/lib/postgresql/data -U replicator -P -W -R
```

5. **Configure GO-NMOS Connection:**
```env
# Use connection pooling with read/write splitting
DATABASE_URL=postgres://nmos_user:password@db-primary:5432/go_nmos?sslmode=require&pool_max_conns=20
# For read replicas (optional, requires application-level routing)
DATABASE_READ_REPLICA_URL=postgres://nmos_user:password@db-replica:5432/go_nmos?sslmode=require&pool_max_conns=10
```

**Connection Pooling:**
- GO-NMOS uses `pgxpool` which supports connection pooling
- Configure `pool_max_conns` in DATABASE_URL for optimal performance
- Recommended: 20 connections for primary, 10 for read replicas

**Automatic Failover Options:**

- **Patroni** - PostgreSQL HA with etcd/Consul/ZooKeeper
- **pg_auto_failover** - Simple PostgreSQL HA solution
- **Cloud Managed Services** - AWS RDS Multi-AZ, Google Cloud SQL HA, Azure Database

**Example with Patroni:**
```yaml
# patroni.yml
scope: go-nmos-cluster
namespace: /go-nmos/
name: postgresql-primary

restapi:
  listen: 0.0.0.0:8008
  connect_address: primary-host:8008

etcd:
  hosts: etcd-host:2379

bootstrap:
  dcs:
    postgresql:
      parameters:
        wal_level: replica
        max_wal_senders: 3
        max_replication_slots: 3
        hot_standby: on
```

#### Database Clustering (Advanced)

For very large deployments, consider:
- **PostgreSQL Cluster** (Citus, Postgres-XL)
- **Read Replicas** - Distribute read queries across multiple replicas
- **Sharding** - Partition data by site/region (future enhancement)

### 2. MQTT Broker High Availability

#### Mosquitto Cluster Setup

**Option 1: Mosquitto Bridge (Simple HA)**

1. **Primary Broker Configuration:**
```conf
# mosquitto-primary.conf
listener 1883
persistence true
persistence_location /mosquitto/data/
connection bridge-secondary
address secondary-broker:1883
topic # both 0
```

2. **Secondary Broker Configuration:**
```conf
# mosquitto-secondary.conf
listener 1883
persistence true
persistence_location /mosquitto/data/
connection bridge-primary
address primary-broker:1883
topic # both 0
```

3. **GO-NMOS Configuration:**
```env
# Use load balancer or DNS round-robin
MQTT_BROKER_URL=tcp://mqtt-cluster.local:1883
```

**Option 2: VerneMQ Cluster (Production-Grade)**

VerneMQ provides built-in clustering:

```yaml
# docker-compose.vernemq-cluster.yml
services:
  vernemq-1:
    image: vernemq/vernemq:latest
    environment:
      - DOCKER_VERNEMQ_ACCEPT_EULA=yes
      - DOCKER_VERNEMQ_ALLOW_ANONYMOUS=off
      - DOCKER_VERNEMQ_NODENAME=VerneMQ@vernemq-1
      - DOCKER_VERNEMQ_DISCOVERY_NODE=VerneMQ@vernemq-1
    ports:
      - "1883:1883"
      - "9001:9001"

  vernemq-2:
    image: vernemq/vernemq:latest
    environment:
      - DOCKER_VERNEMQ_ACCEPT_EULA=yes
      - DOCKER_VERNEMQ_ALLOW_ANONYMOUS=off
      - DOCKER_VERNEMQ_NODENAME=VerneMQ@vernemq-2
      - DOCKER_VERNEMQ_DISCOVERY_NODE=VerneMQ@vernemq-1
    depends_on:
      - vernemq-1
```

**Option 3: Cloud Managed MQTT**

- **AWS IoT Core** - Managed MQTT broker
- **Azure IoT Hub** - Managed messaging
- **Google Cloud IoT Core** - Managed MQTT

**GO-NMOS Configuration:**
```env
# For cloud managed services, use their endpoint
MQTT_BROKER_URL=ssl://your-iot-endpoint.amazonaws.com:8883
MQTT_TOPIC_PREFIX=go-nmos/flows/events
```

### 3. Application Layer High Availability

#### Load Balancing

**Option 1: Nginx Load Balancer**

```nginx
# nginx.conf
upstream go_nmos_backend {
    least_conn;
    server backend-1:8080 max_fails=3 fail_timeout=30s;
    server backend-2:8080 max_fails=3 fail_timeout=30s;
    server backend-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl http2;
    server_name nmos-control.your-campus.local;

    ssl_certificate /etc/ssl/certs/go-nmos.crt;
    ssl_certificate_key /etc/ssl/private/go-nmos.key;

    location / {
        proxy_pass http://go_nmos_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Health check
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    location /api/health {
        proxy_pass http://go_nmos_backend;
        access_log off;
    }
}
```

**Option 2: Traefik Load Balancer**

```yaml
# traefik.yml
api:
  dashboard: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false

certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@your-campus.local
      storage: /letsencrypt/acme.json
      httpChallenge:
        entryPoint: web
```

**Option 3: Kubernetes Service**

```yaml
# k8s-backend-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: go-nmos-backend
spec:
  selector:
    app: go-nmos-backend
  ports:
    - port: 80
      targetPort: 8080
  type: LoadBalancer
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-nmos-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-nmos-backend
  template:
    metadata:
      labels:
        app: go-nmos-backend
    spec:
      containers:
      - name: backend
        image: go-nmos-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: go-nmos-secrets
              key: database-url
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### Session Affinity

For WebSocket connections (MQTT over WebSocket), use session affinity:

```nginx
# Sticky sessions for WebSocket
upstream go_nmos_backend {
    ip_hash;  # Session affinity by IP
    server backend-1:8080;
    server backend-2:8080;
    server backend-3:8080;
}
```

### 4. Registry High Availability

#### External Registry Integration

GO-NMOS supports multiple external NMOS registries. For HA, configure multiple registry endpoints:

**Configuration via UI:**
- Navigate to Settings → Registry Configuration
- Add multiple registry URLs
- Enable/disable registries as needed

**Configuration via API:**
```bash
curl -X PUT http://localhost:9090/api/registry/config \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '[
    {
      "name": "Primary Registry",
      "query_url": "https://registry-primary.your-campus.local/x-nmos/query/v1.3",
      "enabled": true
    },
    {
      "name": "Secondary Registry",
      "query_url": "https://registry-secondary.your-campus.local/x-nmos/query/v1.3",
      "enabled": true
    }
  ]'
```

**External Registry HA Options:**

1. **NMOS Registry with Load Balancer**
   - Deploy multiple registry instances behind load balancer
   - Use DNS round-robin or load balancer
   - Configure GO-NMOS to query load-balanced endpoint

2. **Multiple Independent Registries**
   - Deploy registries per site/region
   - Configure GO-NMOS to query all registries
   - GO-NMOS aggregates results from all registries

3. **Cloud-Managed Registries**
   - Use vendor-specific HA registry solutions
   - Configure GO-NMOS to connect to managed endpoint

### 5. Shared State Considerations

#### Stateless Backend Design

GO-NMOS backend is designed to be stateless:
- All state stored in PostgreSQL database
- No in-memory state that requires session affinity
- Multiple instances can run concurrently

**Exceptions:**
- WebSocket connections for registry events (ephemeral, can reconnect)
- MQTT client connections (reconnect automatically)

#### Database Connection Pooling

Each backend instance maintains its own connection pool:
- No shared connection pool needed
- Each instance connects independently to database
- Connection pool size: 20 connections per instance (adjust based on load)

### 6. Deployment Patterns

#### Pattern 1: Docker Compose with Multiple Replicas

```yaml
# docker-compose.ha.yml
services:
  backend:
    image: go-nmos-backend:latest
    deploy:
      replicas: 3
    environment:
      - DATABASE_URL=postgres://user:pass@db-primary:5432/go_nmos
    depends_on:
      - db-primary

  db-primary:
    image: postgres:16-alpine
    # Configure replication (see Database HA section)

  db-replica:
    image: postgres:16-alpine
    # Configure as replica (see Database HA section)

  mqtt-primary:
    image: eclipse-mosquitto:2
    # Configure bridge (see MQTT HA section)

  mqtt-secondary:
    image: eclipse-mosquitto:2
    # Configure bridge (see MQTT HA section)

  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "443:443"
    depends_on:
      - backend
```

#### Pattern 2: Kubernetes Deployment

See Kubernetes example in "Application Layer High Availability" section above.

#### Pattern 3: Cloud-Native (AWS/GCP/Azure)

**AWS Example:**
- **RDS Multi-AZ** for database HA
- **ECS Fargate** or **EKS** for application HA
- **Application Load Balancer** for traffic distribution
- **MQTT**: AWS IoT Core or self-managed VerneMQ cluster

**GCP Example:**
- **Cloud SQL HA** for database
- **Cloud Run** or **GKE** for application
- **Cloud Load Balancing** for traffic
- **MQTT**: Cloud IoT Core or self-managed

### 7. Health Checks and Monitoring

#### Application Health Checks

GO-NMOS provides health check endpoints:

- **Liveness**: `GET /api/health` - Fast check, returns 200 if service is alive
- **Readiness**: `GET /api/health/detail` - Detailed check, includes database and MQTT status

**Kubernetes Health Checks:**
```yaml
livenessProbe:
  httpGet:
    path: /api/health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /api/health/detail
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

#### Database Health Monitoring

Monitor database replication lag:
```sql
-- On replica
SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) AS replication_lag_seconds;
```

Monitor connection pool usage:
```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'go_nmos';
```

### 8. Disaster Recovery

#### Backup Strategy

1. **Database Backups:**
```bash
# Automated daily backups
pg_dump -h db-primary -U nmos_user go_nmos | gzip > backup-$(date +%Y%m%d).sql.gz

# Point-in-time recovery (PITR)
# Configure WAL archiving in postgresql.conf
archive_mode = on
archive_command = 'cp %p /backup/wal/%f'
```

2. **Configuration Backups:**
- Backup `backend/.env` files
- Backup registry configurations (via API export)
- Backup SSL certificates

3. **Recovery Procedures:**
- Document recovery time objectives (RTO) and recovery point objectives (RPO)
- Test restore procedures regularly
- Maintain off-site backups

### 9. Performance Considerations

#### Connection Pooling

- **Database**: 20 connections per backend instance
- **MQTT**: Single connection per backend instance (reconnects automatically)
- **HTTP**: No connection pooling needed (stateless)

#### Caching Strategy

- GO-NMOS does not currently implement caching
- Consider Redis/Memcached for:
  - Registry query results (future enhancement)
  - User session data (if implementing session-based auth)
  - Frequently accessed settings

### 10. Security in HA Deployments

#### Shared Secrets

- **JWT_SECRET**: Must be identical across all backend instances
- Use Kubernetes Secrets or HashiCorp Vault
- Rotate secrets regularly

#### Network Security

- Use TLS/SSL for all inter-service communication
- Implement network policies (Kubernetes) or security groups (AWS)
- Use VPN or private networks for database access

### 11. Testing HA Setup

#### Failover Testing

1. **Database Failover:**
   - Stop primary database
   - Verify replica promotion
   - Verify backend reconnection

2. **Backend Failover:**
   - Stop one backend instance
   - Verify load balancer routes traffic to remaining instances
   - Verify no data loss

3. **MQTT Failover:**
   - Stop primary MQTT broker
   - Verify clients reconnect to secondary
   - Verify message delivery continues

#### Load Testing

Use tools like `k6` or `Apache Bench` to test:
- Concurrent user sessions
- Database connection pool limits
- MQTT message throughput
- API response times under load

## Quick Reference

### Minimum HA Setup

- **Database**: Primary + 1 Replica (PostgreSQL streaming replication)
- **MQTT**: 2 Brokers with bridge configuration
- **Backend**: 2+ instances behind load balancer
- **Load Balancer**: Nginx or cloud LB

### Recommended HA Setup

- **Database**: Primary + 2 Replicas with automatic failover (Patroni)
- **MQTT**: 3-node cluster (VerneMQ or Mosquitto bridge)
- **Backend**: 3+ instances behind load balancer
- **Load Balancer**: Nginx/Traefik with health checks
- **Monitoring**: Prometheus + Grafana
- **Alerting**: Configured alert hooks (Slack/webhook)

## Additional Resources

- [PostgreSQL High Availability](https://www.postgresql.org/docs/current/high-availability.html)
- [Mosquitto Bridge Configuration](https://mosquitto.org/man/mosquitto-conf-5.html)
- [VerneMQ Clustering](https://docs.vernemq.com/clustering/introduction)
- [Kubernetes Deployment Patterns](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
