# Demo Instance Setup Guide

This guide explains how to set up and deploy the Night Owls demo instance with realistic seed data.

## 🎯 Overview

The demo instance provides a fully functional version of Night Owls with:
- **50 demo users** with realistic South African phone numbers
- **Multiple user roles**: 2 admins, ~40 owls, ~8 guests  
- **Historical data**: Past bookings with realistic attendance patterns
- **Future bookings**: Next 30 days of scheduled shifts
- **Incident reports**: Various severity levels for demonstration
- **Dev mode enabled**: Any 6-digit OTP works for authentication

## 🌐 DNS Configuration

Your DNS should be configured as follows:

```
realinstance.nightowls.app  → Production data (test group)
demo.nightowls.app         → Demo instance (seeded data)
nightowls.app              → Marketing/info site
```

## 🚀 Deployment Options

### Option 1: GitHub Actions (Recommended)

1. **Manual Deployment**: 
   - Go to Actions → "Deploy Demo Instance"
   - Click "Run workflow"
   - Choose options:
     - Reset demo data: `true` (recommended)
     - User count: `50` (default)

2. **Automatic Deployment**:
   - Push to `demo` branch automatically deploys

### Option 2: Local Testing

```bash
# Start demo instance locally
make demo-start

# Check status
make demo-status

# View logs
make demo-logs

# Reset with fresh data
make demo-reset

# Stop demo
make demo-stop
```

## 🏗️ Infrastructure Setup

### Production Server Setup

Your server needs the following files:

1. **docker-compose.demo.yml** - Demo container configuration
2. **Caddyfile.multi** - Multi-domain reverse proxy config
3. **Environment variables** in `.env`

### GitHub Secrets Required

Add these to your repository secrets:

```bash
SSH_PRIVATE_KEY          # SSH key for server access
CONTAINER_REGISTRY_TOKEN # GitHub Container Registry token
```

### GitHub Environment

Create a "Demo" environment in GitHub Settings → Environments with:
- Same secrets as Production
- Optional: Protection rules for demo deployments

## 📦 What Gets Deployed

### Demo Container Configuration

```yaml
# Runs on port 5889 (internal: 5888)
services:
  night-owls-demo:
    image: ghcr.io/fermain/night-owls-go:latest
    environment:
      - DEV_MODE=true              # Dev features enabled
      - JWT_EXPIRATION_HOURS=168   # 1 week sessions
      - DATABASE_PATH=./data/demo.db
      # Twilio disabled (uses mock OTP)
```

### Demo Data Includes

**Users (50 total)**:
- **Alice Admin** (+27821234567) - Primary admin
- **Bob Manager** (+27821234568) - Secondary admin  
- **Charlie, Diana, Eve, Frank, Grace, Henry** - Owl volunteers
- **Iris, Jack, Maya, Ryan, Aria, Dean, Nora, Kyle** - Guest users
- **Leo, Zoe, Max, Ivy, Sam, Ruby, Alex, Nova, Finn, Luna** - Additional owls

**Historical Bookings**:
- Past 2 weeks of realistic shift patterns
- Mix of attended/missed shifts for realistic reporting
- Buddy assignments where appropriate

**Future Bookings**:
- Next 30 days of scheduled shifts
- Follows realistic patrol patterns:
  - Daily Evening Patrol: Weekdays 6 PM
  - Weekend Morning Watch: Weekends 6 AM & 10 AM
  - Weekday Lunch Security: Occasional noon shifts

**Incident Reports**:
- ~30% of shifts have reports
- Realistic severity distribution (60% normal, 30% suspicion, 10% incident)
- Sample messages for each severity level

## 🔧 Configuration Details

### Multi-Domain Caddy Setup

The `Caddyfile.multi` handles three domains:

```caddyfile
# Production + Real Instance  
mm.nightowls.app, realinstance.nightowls.app {
    reverse_proxy night-owls:5888
}

# Demo Instance
demo.nightowls.app {
    reverse_proxy night-owls-demo:5888
}

# Marketing Site
nightowls.app {
    root * /srv/marketing
    handle /demo { redir https://demo.nightowls.app }
    handle /app  { redir https://realinstance.nightowls.app }
}
```

### Demo-Specific Features

**Development Mode Benefits**:
- Any 6-digit OTP works (no real SMS needed)
- Extended debugging/logging
- Faster iteration for demos

**Security Considerations**:
- Demo uses mock authentication
- Separate database from production
- Shorter log retention
- No real SMS/email integration

## 🧪 Testing the Demo

### Health Checks

```bash
# Backend health
curl https://demo.nightowls.app/health

# API documentation
curl https://demo.nightowls.app/swagger/

# Admin dashboard (requires auth)
curl https://demo.nightowls.app/admin
```

### Demo Login

1. **Go to**: https://demo.nightowls.app
2. **Phone**: +27821234567 (Alice Admin)
3. **OTP**: Any 6 digits (e.g., 123456)
4. **Access**: Full admin features

### What to Demonstrate

**Admin Features**:
- User management (50 diverse users)
- Scheduling system (3 different schedule types)
- Reporting dashboard with real incident data
- Analytics with historical trends

**User Features**:
- Shift booking system
- Check-in functionality  
- Incident reporting
- Buddy system

**System Features**:
- Real-time updates
- Mobile-responsive design
- Push notifications (if configured)
- Audit logging

## 🔄 Maintenance

### Resetting Demo Data

**Automatic** (via GitHub Actions):
- Run "Deploy Demo Instance" workflow
- Enable "Reset demo data" option

**Manual** (via SSH):
```bash
cd ~/night-owls-go
docker compose -f docker-compose.demo.yml down
docker volume rm night_owls_demo_data
docker compose -f docker-compose.demo.yml up -d
```

### Monitoring

**Logs**:
```bash
# Via script
./scripts/deploy-demo.sh logs

# Direct Docker
docker compose -f docker-compose.demo.yml logs -f
```

**Status**:
```bash
# Via script  
./scripts/deploy-demo.sh status

# Direct Docker
docker compose -f docker-compose.demo.yml ps
```

## 🚨 Troubleshooting

### Common Issues

**Demo not accessible**:
1. Check DNS pointing to correct server IP
2. Verify Caddy is using multi-domain config
3. Check demo container health: `docker compose -f docker-compose.demo.yml ps`

**Seeding failed**:
1. Check container logs: `docker logs night-owls-demo`
2. Verify database permissions
3. Try manual seed: `docker compose -f docker-compose.demo.yml run --rm night-owls-demo-seed ./seed --reset --users 50`

**OTP not working**:
- Demo uses dev mode - any 6-digit code should work
- Check DEV_MODE=true in container environment

### Recovery Procedures

**Full demo reset**:
```bash
# Stop everything
docker compose -f docker-compose.demo.yml down

# Clean volumes  
docker volume rm night_owls_demo_data

# Restart
docker compose -f docker-compose.demo.yml up -d

# Re-seed
docker compose -f docker-compose.demo.yml run --rm night-owls-demo-seed \
  ./seed --reset --users 50 --future-bookings --verbose
```

**Rollback to single-domain**:
```bash
# Restore original Caddyfile
cp Caddyfile.single.backup Caddyfile
docker compose restart caddy
```

## 📋 Checklist for Demo Deployment

- [ ] DNS records point to server IP
- [ ] GitHub secrets configured
- [ ] Demo environment created
- [ ] Multi-domain Caddyfile tested
- [ ] Demo containers healthy
- [ ] Seed data populated successfully
- [ ] Admin login works (+27821234567)
- [ ] API health check passes
- [ ] SSL certificates generated
- [ ] Monitoring/logs accessible

## 🔗 Access URLs

After successful deployment:

- **Demo App**: https://demo.nightowls.app
- **API Health**: https://demo.nightowls.app/health
- **API Docs**: https://demo.nightowls.app/swagger/
- **Real Instance**: https://realinstance.nightowls.app  
- **Marketing**: https://nightowls.app (placeholder)

## 📞 Demo Credentials

**Admin Users**:
- Alice Admin: +27821234567
- Bob Manager: +27821234568

**Sample Owl Users**:
- Charlie Volunteer: +27821234569
- Diana Scout: +27821234570
- Eve Patrol: +27821234571

**Sample Guest Users**:
- Iris Guest: +27821234575
- Jack Visitor: +27821234576

**OTP**: Any 6 digits work in dev mode! 