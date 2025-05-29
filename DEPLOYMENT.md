# Night Owls Control - Production Deployment Guide

## Overview

This guide covers deploying the Night Owls Control application to a DigitalOcean droplet for public trial. The setup includes automated deployments, monitoring, backups, and security best practices.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   DigitalOcean Droplet                   │
├─────────────────────────────────────────────────────────┤
│  Nginx (Reverse Proxy)                                  │
│    ├── Frontend (SvelteKit) - Port 3000                │
│    └── Backend (Go API) - Port 5888                    │
│                                                         │
│  SQLite Database                                        │
│  Systemd Services                                       │
│  Let's Encrypt SSL                                      │
└─────────────────────────────────────────────────────────┘
```

## Prerequisites

### 1. DigitalOcean Droplet Setup
- Ubuntu 22.04 LTS
- Minimum 2GB RAM, 2 vCPUs
- 50GB SSD storage
- Reserved IP address

### 2. Domain Setup
- Point your domain to the droplet IP
- Configure DNS records:
  - A record: `@` → Droplet IP
  - A record: `www` → Droplet IP
  - A record: `api` → Droplet IP (optional)

## Initial Server Setup

### 1. Create Deploy User

```bash
# As root
adduser deploy
usermod -aG sudo deploy
su - deploy
```

### 2. SSH Key Setup

```bash
# On your local machine
ssh-copy-id deploy@your-server-ip

# On server, disable password auth
sudo nano /etc/ssh/sshd_config
# Set: PasswordAuthentication no
sudo systemctl restart ssh
```

### 3. Firewall Configuration

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 4. Install Dependencies

```bash
# System packages
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git build-essential nginx certbot python3-certbot-nginx

# Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Node.js 20+
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# SQLite tools
sudo apt install -y sqlite3

# Process manager
sudo npm install -g pm2
```

## Application Setup

### 1. Clone Repository

```bash
cd ~
git clone https://github.com/yourusername/night-owls-go.git
cd night-owls-go
```

### 2. Backend Setup

```bash
# Install Go dependencies
go mod download

# Build backend
CGO_ENABLED=1 go build -o night-owls-server ./cmd/server

# Create production env file
cat > .env.production << EOF
# Server Configuration
SERVER_PORT=5888
STATIC_DIR=./app/build
DEV_MODE=false

# Database Configuration
DATABASE_PATH=/home/deploy/night-owls-data/production.db

# JWT Configuration (Generate secure secret!)
JWT_SECRET=$(openssl rand -base64 32)

# VAPID Keys for Web Push (Generate with: npx web-push generate-vapid-keys)
VAPID_PUBLIC_KEY=your-public-key
VAPID_PRIVATE_KEY=your-private-key
VAPID_SUBJECT=mailto:admin@yourdomain.com

# SMS/OTP Configuration
OTP_LOG_PATH=/home/deploy/night-owls-data/logs/sms_outbox.log

# Twilio (for production SMS)
TWILIO_ACCOUNT_SID=your-sid
TWILIO_AUTH_TOKEN=your-token
TWILIO_FROM_NUMBER=+1234567890
EOF

# Create data directories
mkdir -p ~/night-owls-data/{logs,backups}
chmod 700 ~/night-owls-data
```

### 3. Frontend Setup

```bash
cd app
npm ci --production=false

# Create production env
cat > .env.production << EOF
PUBLIC_API_URL=https://api.yourdomain.com
PUBLIC_APP_URL=https://yourdomain.com
EOF

# Build frontend
npm run build
```

### 4. Database Migration

```bash
cd ~/night-owls-go
./night-owls-server -migrate-only
```

## Systemd Service Setup

### 1. Backend Service

```bash
sudo nano /etc/systemd/system/night-owls-backend.service
```

```ini
[Unit]
Description=Night Owls Control Backend
After=network.target

[Service]
Type=simple
User=deploy
WorkingDirectory=/home/deploy/night-owls-go
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"
EnvironmentFile=/home/deploy/night-owls-go/.env.production
ExecStart=/home/deploy/night-owls-go/night-owls-server
Restart=always
RestartSec=5
StandardOutput=append:/home/deploy/night-owls-data/logs/backend.log
StandardError=append:/home/deploy/night-owls-data/logs/backend-error.log

[Install]
WantedBy=multi-user.target
```

### 2. Frontend Service (if using Node adapter)

```bash
sudo nano /etc/systemd/system/night-owls-frontend.service
```

```ini
[Unit]
Description=Night Owls Control Frontend
After=network.target

[Service]
Type=simple
User=deploy
WorkingDirectory=/home/deploy/night-owls-go/app
Environment="NODE_ENV=production"
Environment="PORT=3000"
ExecStart=/usr/bin/node build
Restart=always
RestartSec=5
StandardOutput=append:/home/deploy/night-owls-data/logs/frontend.log
StandardError=append:/home/deploy/night-owls-data/logs/frontend-error.log

[Install]
WantedBy=multi-user.target
```

### 3. Enable Services

```bash
sudo systemctl daemon-reload
sudo systemctl enable night-owls-backend
sudo systemctl enable night-owls-frontend
sudo systemctl start night-owls-backend
sudo systemctl start night-owls-frontend
```

## Nginx Configuration

### 1. Create Site Configuration

```bash
sudo nano /etc/nginx/sites-available/night-owls
```

```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# Main HTTPS server
server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL configuration (will be managed by Certbot)
    # ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self' https:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';" always;

    # Logging
    access_log /var/log/nginx/night-owls-access.log;
    error_log /var/log/nginx/night-owls-error.log;

    # Frontend proxy
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # API proxy
    location /api {
        proxy_pass http://localhost:5888;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Increase timeout for long-running requests
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }

    # Additional backend routes
    location ~ ^/(schedules|shifts|bookings|reports|push|swagger) {
        proxy_pass http://localhost:5888;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static file caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 2. Enable Site

```bash
sudo ln -s /etc/nginx/sites-available/night-owls /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 3. SSL Certificate

```bash
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

## Deployment Script

Create `deploy.sh` in the repository:

```bash
#!/bin/bash
set -e

echo "🚀 Starting deployment..."

# Configuration
REMOTE_USER="deploy"
REMOTE_HOST="your-server-ip"
REMOTE_DIR="/home/deploy/night-owls-go"

# Build frontend locally
echo "📦 Building frontend..."
cd app
npm ci
npm run build
cd ..

# Build backend
echo "🔨 Building backend..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o night-owls-server ./cmd/server

# Deploy files
echo "📤 Uploading files..."
rsync -avz --delete \
  --exclude='.git' \
  --exclude='node_modules' \
  --exclude='.env' \
  --exclude='*.db' \
  --exclude='*.log' \
  . $REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/

# Run remote commands
echo "🔄 Restarting services..."
ssh $REMOTE_USER@$REMOTE_HOST << 'ENDSSH'
cd /home/deploy/night-owls-go

# Backup database
cp /home/deploy/night-owls-data/production.db \
   /home/deploy/night-owls-data/backups/production-$(date +%Y%m%d-%H%M%S).db

# Run migrations
./night-owls-server -migrate-only

# Restart services
sudo systemctl restart night-owls-backend
sudo systemctl restart night-owls-frontend

# Check status
sudo systemctl status night-owls-backend --no-pager
sudo systemctl status night-owls-frontend --no-pager
ENDSSH

echo "✅ Deployment complete!"
```

Make it executable:
```bash
chmod +x deploy.sh
```

## Monitoring Setup

### 1. Log Rotation

```bash
sudo nano /etc/logrotate.d/night-owls
```

```
/home/deploy/night-owls-data/logs/*.log {
    daily
    missingok
    rotate 14
    compress
    notifempty
    create 0644 deploy deploy
    sharedscripts
    postrotate
        systemctl reload night-owls-backend > /dev/null 2>&1 || true
    endscript
}
```

### 2. Health Check Endpoint

Add to your Go backend:
```go
// Health check endpoint
fuego.GetStd(s, "/health", func(w http.ResponseWriter, r *http.Request) {
    // Check database
    if err := dbConn.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "db": "down"})
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "db": "up"})
})
```

### 3. Uptime Monitoring

Use a service like:
- UptimeRobot
- Pingdom
- DigitalOcean Monitoring

Configure to check `https://yourdomain.com/health` every 5 minutes.

## Database Backup

### 1. Automated Backup Script

```bash
nano ~/backup-database.sh
```

```bash
#!/bin/bash
BACKUP_DIR="/home/deploy/night-owls-data/backups"
DB_PATH="/home/deploy/night-owls-data/production.db"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/production-$TIMESTAMP.db"

# Create backup
sqlite3 $DB_PATH ".backup '$BACKUP_FILE'"

# Compress
gzip $BACKUP_FILE

# Keep only last 7 days
find $BACKUP_DIR -name "*.db.gz" -mtime +7 -delete

# Optional: Upload to cloud storage
# aws s3 cp $BACKUP_FILE.gz s3://your-bucket/backups/
```

### 2. Schedule Backup

```bash
crontab -e
```

Add:
```
0 2 * * * /home/deploy/backup-database.sh
```

## Security Checklist

### Before Going Live

- [ ] Change all default passwords
- [ ] Generate strong JWT secret
- [ ] Configure CORS properly
- [ ] Enable rate limiting
- [ ] Set up fail2ban
- [ ] Configure firewall rules
- [ ] Enable automatic security updates
- [ ] Implement input validation
- [ ] Set up HTTPS only
- [ ] Configure secure headers

### Environment Variables

- [ ] Remove `.env` files from repository
- [ ] Use strong, unique secrets
- [ ] Rotate secrets regularly
- [ ] Implement secret management

## Performance Optimization

### 1. Enable Gzip

Add to Nginx config:
```nginx
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss;
```

### 2. Database Optimization

```bash
# Analyze and optimize SQLite
sqlite3 /home/deploy/night-owls-data/production.db "VACUUM;"
sqlite3 /home/deploy/night-owls-data/production.db "ANALYZE;"
```

### 3. CDN Setup (Optional)

Consider using Cloudflare for:
- DDoS protection
- Global CDN
- SSL termination
- Caching

## Rollback Strategy

### Quick Rollback

```bash
# Keep previous build
cp night-owls-server night-owls-server.backup

# If deployment fails
mv night-owls-server.backup night-owls-server
sudo systemctl restart night-owls-backend
```

### Database Rollback

```bash
# Stop services
sudo systemctl stop night-owls-backend

# Restore backup
cp /home/deploy/night-owls-data/backups/production-TIMESTAMP.db \
   /home/deploy/night-owls-data/production.db

# Restart
sudo systemctl start night-owls-backend
```

## Monitoring Commands

```bash
# Check service status
sudo systemctl status night-owls-backend
sudo systemctl status night-owls-frontend

# View logs
sudo journalctl -u night-owls-backend -f
sudo journalctl -u night-owls-frontend -f

# Check resource usage
htop
df -h
free -m

# Database size
du -h /home/deploy/night-owls-data/production.db
```

## Troubleshooting

### Common Issues

1. **502 Bad Gateway**
   - Check if services are running
   - Verify port configurations
   - Check Nginx error logs

2. **Database Locked**
   - Check for long-running queries
   - Ensure single process access
   - Consider WAL mode

3. **High Memory Usage**
   - Monitor Go routine leaks
   - Check database connections
   - Implement connection pooling

## Next Steps

1. Set up monitoring and alerting
2. Configure automated backups
3. Implement CI/CD pipeline
4. Add application metrics
5. Set up error tracking (Sentry)
6. Configure log aggregation
7. Implement A/B testing
8. Set up staging environment 