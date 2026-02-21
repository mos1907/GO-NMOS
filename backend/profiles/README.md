# Environment Profiles

This directory contains reference configuration profiles for different deployment scenarios.

## Profiles

### 1. Lab / Single Node (`lab.env.example`)

**Use Case:** Development, testing, small labs, single encoder/decoder setups

**Characteristics:**
- Minimal configuration
- Relaxed rate limiting (1200 RPM)
- Single registry expected
- No multi-site considerations
- PTP domain validation optional
- Suitable for < 10 nodes
- Authentication can be disabled for convenience
- HTTPS typically disabled

**Quick Start:**
```bash
cp profiles/lab.env.example backend/.env
# Edit .env and adjust as needed
```

### 2. Small Facility (`small-facility.env.example`)

**Use Case:** Small studios, OB vans, regional broadcasters

**Characteristics:**
- Production-ready security settings
- Moderate rate limiting (600 RPM)
- Single registry typically sufficient
- 1-2 sites possible (e.g., Studio A, Studio B)
- PTP domain should be configured and validated
- Alerting recommended (Slack/webhook)
- Database backups recommended
- HTTPS recommended
- Suitable for 10-50 nodes

**Quick Start:**
```bash
cp profiles/small-facility.env.example backend/.env
# IMPORTANT: Change JWT_SECRET and INIT_ADMIN_PASSWORD!
# Configure alerting webhooks
# Set up SSL certificates for HTTPS
```

### 3. Large Campus (`large-campus.env.example`)

**Use Case:** Large broadcast campuses, multiple sites, HA-ready deployments

**Characteristics:**
- Production-grade security (strong secrets, HTTPS required)
- Higher rate limiting (1200 RPM)
- Multiple registries expected (per site/region)
- Multi-site routing and cross-site visibility
- PTP domain validation critical (multiple domains possible)
- Alerting essential (multiple channels)
- Database HA/clustering recommended
- MQTT broker HA recommended
- Load balancing for API recommended
- Regular backups and disaster recovery planning
- Suitable for 50+ nodes, multiple sites
- May need multiple backend instances

**Quick Start:**
```bash
cp profiles/large-campus.env.example backend/.env
# CRITICAL: Change JWT_SECRET (64+ characters) and INIT_ADMIN_PASSWORD!
# Configure alerting webhooks (Slack + webhook)
# Set up SSL certificates
# Configure database HA/clustering
# Set up MQTT broker cluster
# Configure load balancer
```

## Configuration Guidelines

### Security

- **JWT_SECRET**: Use a strong random secret. Minimum 32 characters for production, 64+ for large campuses.
- **INIT_ADMIN_PASSWORD**: Use a strong password. Never use defaults in production.
- **HTTPS_ENABLED**: Enable HTTPS for all production deployments.
- **DISABLE_AUTH**: Never disable authentication in production.

### Database

- **Lab**: Local PostgreSQL, no special requirements
- **Small Facility**: Single PostgreSQL instance, regular backups
- **Large Campus**: HA/clustered PostgreSQL, connection pooling, automated backups

### MQTT

- **Lab**: Single broker, optional
- **Small Facility**: Single broker, recommended for real-time events
- **Large Campus**: HA/clustered broker, required for campus-wide events

### Alerting

- **Lab**: Logging only (LoggerHook)
- **Small Facility**: Slack or webhook recommended
- **Large Campus**: Multiple channels (Slack + webhook) essential

### Rate Limiting

- **Lab**: 1200 RPM (relaxed for development)
- **Small Facility**: 600 RPM (moderate)
- **Large Campus**: 1200 RPM (higher for scale)

### Multi-Site Considerations

- **Lab**: Not applicable
- **Small Facility**: 1-2 sites possible, basic site tagging
- **Large Campus**: Multiple sites, cross-site routing, site-specific policies

## Migration Between Profiles

When moving from one profile to another:

1. **Backup current configuration**: Copy your `.env` file
2. **Review differences**: Compare current config with target profile
3. **Update gradually**: Change settings incrementally, test after each change
4. **Security first**: Update secrets and passwords immediately
5. **Test thoroughly**: Verify all functionality after migration

## Docker Compose Overrides

For different profiles, you may want to use Docker Compose override files:

```bash
# Lab profile
docker compose -f docker-compose.yml -f docker-compose.lab.yml up

# Small facility profile
docker compose -f docker-compose.yml -f docker-compose.small-facility.yml up

# Large campus profile
docker compose -f docker-compose.yml -f docker-compose.large-campus.yml up
```

## Environment Variables Reference

See `backend/env.example` for a complete list of all available environment variables.
