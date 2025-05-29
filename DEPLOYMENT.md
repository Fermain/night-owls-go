# Night Owls Control - Production Deployment Guide

## Overview

This guide covers deploying the Night Owls Control application using Docker containers with Caddy as a reverse proxy for automatic HTTPS. The application will be deployed to `mm.nightowls.app` with modern containerized infrastructure.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Production Server                      │
├─────────────────────────────────────────────────────────┤
│  Caddy (Reverse Proxy + Automatic HTTPS)               │
│    ├── Frontend (SvelteKit Static) ──┐                 │
│    └── Backend (Go API)              │                 │
│                                       │                 │
│  Docker Container: night-owls-app     │                 │
│    ├── Go Backend Server (Port 5888) │                 │
│    ├── SvelteKit Frontend (Static)   │                 │
│    └── SQLite Database               │                 │
│                                       │                 │
│  Docker Volumes:                      │                 │
│    ├── night_owls_data (Database)    │                 │
│    ├── caddy_data (SSL Certificates) │                 │
│    └── caddy_config (Caddy Config)   │                 │
└─────────────────────────────────────────────────────────┘
```

## Quick Start (Recommended: Docker Deployment)

### Prerequisites

- Docker and Docker Compose installed
- Domain `mm.nightowls.app` pointing to your server
- Server with minimum 2GB RAM, 1 vCPU, 20GB storage

### 1. Clone and Setup

```bash
git clone https://github.com/Fermain/night-owls-go
cd night-owls-go

# Quick setup with pnpm (installs dependencies and generates VAPID keys)
./setup-pnpm.sh

# Or manual setup:
# Create production environment file
cp .env.example .env.production
```

### 2. Configure Environment Variables

Edit `.env.production` with your actual values:

```bash
# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Generate VAPID keys for push notifications
pnpm dlx web-push generate-vapid-keys

# Add your Twilio credentials
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_NUMBER=+1234567890
```

### 3. Deploy

```bash
# Make deployment script executable
chmod +x deploy-docker.sh

# Deploy the application
./deploy-docker.sh
```

The application will be available at `https://mm.nightowls.app` with automatic HTTPS via Let's Encrypt.

## Detailed Setup Instructions

### Domain Configuration

Ensure your DNS is configured correctly:

```
Type    Name              Value
A       mm.nightowls.app  YOUR_SERVER_IP
A       *.mm.nightowls.app YOUR_SERVER_IP (optional, for subdomains)
```

### Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | Secure random string for JWT tokens | Generated with `openssl rand -base64 32` |
| `VAPID_PUBLIC_KEY` | Public key for web push notifications | Generated with `pnpm dlx web-push generate-vapid-keys` |
| `VAPID_PRIVATE_KEY` | Private key for web push notifications | Generated with `pnpm dlx web-push generate-vapid-keys` |
| `VAPID_SUBJECT` | Contact email for push service | `mailto:admin@mm.nightowls.app` |
| `TWILIO_ACCOUNT_SID` | Twilio account identifier | `ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` |
| `TWILIO_AUTH_TOKEN` | Twilio authentication token | `your_auth_token` |
| `TWILIO_FROM_NUMBER` | Twilio phone number for SMS | `+1234567890` |

### File Structure

```
night-owls-go/
├── Dockerfile                 # Multi-stage build for app
├── docker-compose.yml         # Production compose file
├── Caddyfile                 # Caddy reverse proxy config
├── env.production.example    # Environment template
├── .env.production          # Your production config
├── deploy-docker.sh         # Deployment script
├── backup.sh               # Backup script
└── app/                    # Frontend source
```

## Container Management

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f night-owls
docker-compose logs -f caddy
```

### Service Management

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# Restart services
docker-compose restart

# Check status
docker-compose ps
```

### Health Checks

```bash
# Check application health
curl https://mm.nightowls.app/health

# Check container health
docker-compose ps
```

## Backup and Recovery

### Automated Backups

```bash
# Make backup script executable
chmod +x backup.sh

# Run manual backup
./backup.sh

# Schedule with cron (daily at 2 AM)
crontab -e
# Add: 0 2 * * * /path/to/night-owls-go/backup.sh
```

### Restore from Backup

```bash
# Stop the application
docker-compose down

# Extract backup
gunzip backups/backup-TIMESTAMP.db.gz

# Restore database
docker run --rm -v night_owls_data:/data -v $(pwd)/backups:/backups alpine \
  cp /backups/backup-TIMESTAMP.db /data/production.db

# Start the application
docker-compose up -d
```

## Monitoring

### Application Metrics

The application provides the following endpoints:

- **Health Check**: `https://mm.nightowls.app/health`
- **API Documentation**: `https://mm.nightowls.app/swagger`

### Log Analysis

```bash
# View recent application logs
docker-compose logs --tail=100 night-owls

# Follow logs in real-time
docker-compose logs -f

# Export logs for analysis
docker-compose logs --since=24h > logs-$(date +%Y%m%d).log
```

### Resource Monitoring

```bash
# Container resource usage
docker stats

# Disk usage
docker system df

# Container sizes
docker images
```

## Security Configuration

### Caddy Security Features

The Caddyfile includes:

- **Automatic HTTPS** with Let's Encrypt
- **Security Headers**:
  - HSTS (HTTP Strict Transport Security)
  - X-Content-Type-Options
  - X-Frame-Options
  - X-XSS-Protection
  - Referrer-Policy
  - Permissions-Policy

### Container Security

- **Non-root user**: Application runs as `appuser`
- **Read-only root filesystem** (except data volumes)
- **Resource limits** configured in docker-compose.yml
- **Health checks** for service monitoring

### Network Security

- **Internal network**: Containers communicate via private network
- **Exposed ports**: Only 80 and 443 exposed publicly
- **Proxy headers**: Real IP forwarding configured

## Performance Optimization

### Caching Strategy

- **Static assets**: Long-term caching (1 year)
- **API responses**: No caching for dynamic content
- **Gzip compression**: Enabled for all text content

### Container Optimization

- **Multi-stage builds**: Minimal production image
- **Alpine Linux**: Small base image
- **Build caching**: Docker layer caching optimized

## Scaling Options

### Vertical Scaling

Increase server resources:

```yaml
# docker-compose.yml
services:
  night-owls:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
```

### Horizontal Scaling

For high availability:

1. **Load Balancer**: Add multiple app containers
2. **Database**: Consider PostgreSQL for better concurrency
3. **File Storage**: Use shared storage for uploads
4. **Session Storage**: Use Redis for session management

## Alternative Deployment: Traditional Server

If you prefer not to use Docker:

### Prerequisites

- Ubuntu 22.04 LTS server
- Go 1.21+ installed
- Node.js 20+ installed
- Caddy installed

### Setup

```bash
# Install Caddy
sudo apt update
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Build application
go build -o night-owls-server ./cmd/server

# Build frontend
cd app && pnpm install && pnpm run build && cd ..

# Install systemd service
sudo cp scripts/night-owls.service /etc/systemd/system/
sudo systemctl enable night-owls
sudo systemctl start night-owls

# Configure Caddy
sudo cp Caddyfile /etc/caddy/
sudo systemctl reload caddy
```

## Troubleshooting

### Common Issues

1. **Container won't start**
   ```bash
   # Check logs
   docker-compose logs night-owls
   
   # Check environment variables
   docker-compose config
   ```

2. **SSL Certificate issues**
   ```bash
   # Check Caddy logs
   docker-compose logs caddy
   
   # Verify domain points to server
   dig mm.nightowls.app
   ```

3. **Database connection errors**
   ```bash
   # Check volume permissions
   docker-compose exec night-owls ls -la /app/data/
   
   # Check database file
   docker-compose exec night-owls sqlite3 /app/data/production.db ".tables"
   ```

4. **High memory usage**
   ```bash
   # Check container stats
   docker stats
   
   # Adjust memory limits in docker-compose.yml
   ```

### Recovery Procedures

1. **Complete system failure**
   ```bash
   # Restore from backup
   docker-compose down
   ./restore-backup.sh TIMESTAMP
   docker-compose up -d
   ```

2. **Corrupted database**
   ```bash
   # Stop application
   docker-compose stop night-owls
   
   # Restore database only
   docker run --rm -v night_owls_data:/data -v $(pwd)/backups:/backups alpine \
     cp /backups/backup-TIMESTAMP.db /data/production.db
   
   # Start application
   docker-compose start night-owls
   ```

## Production Checklist

Before going live:

- [ ] Domain configured and pointing to server
- [ ] SSL certificate obtained (automatic with Caddy)
- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] Backup strategy implemented
- [ ] Monitoring setup
- [ ] Log rotation configured
- [ ] Security headers validated
- [ ] Performance testing completed
- [ ] Disaster recovery plan documented

## Maintenance

### Regular Tasks

- **Daily**: Check application health and logs
- **Weekly**: Review backup integrity
- **Monthly**: Update dependencies and security patches
- **Quarterly**: Performance review and optimization

### Updates

```bash
# Pull latest code
git pull origin main

# Rebuild and deploy
./deploy-docker.sh

# Verify deployment
curl https://mm.nightowls.app/health
```

### Security Updates

```bash
# Update base images
docker-compose pull

# Rebuild application
docker-compose build --no-cache

# Deploy updates
./deploy-docker.sh
```

## Support

For deployment issues:

1. Check logs: `docker-compose logs -f`
2. Verify health: `curl https://mm.nightowls.app/health`
3. Review backup integrity: `./backup.sh`
4. Check resource usage: `docker stats`

## Next Steps

After successful deployment:

1. Set up monitoring and alerting
2. Configure automated backups
3. Implement CI/CD pipeline
4. Add application performance monitoring
5. Set up error tracking (e.g., Sentry)
6. Configure log aggregation
7. Implement blue-green deployments
8. Set up staging environment 