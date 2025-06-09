# GitHub Actions Workflows

This repository uses an optimized GitHub Actions CI/CD pipeline that eliminates redundancy and provides fast feedback.

## 🎯 **Optimized Strategy Overview**

### **For Pull Requests** → `pr-check.yml`
- ✅ **Smart Testing**: Only tests changed components (backend/frontend/docker)
- ✅ **Fast Feedback**: ~3-8 minutes depending on changes
- ✅ **Security Scans**: Dependency review and vulnerability scanning
- ✅ **Build Verification**: Docker build check (no push)
- ❌ **No Deployment**: PRs don't deploy anywhere

### **For Main Branch** → `ci.yml` + `deploy.yml`  
- ✅ **Full CI Suite**: Comprehensive testing on merged code
- ✅ **Artifact Reuse**: No redundant rebuilds between CI and Deploy
- ✅ **Production Deploy**: Automatic deployment to production
- ✅ **Optional Demo**: Manual or triggered demo deployment

## 🔄 **Workflow Details**

### 1. **PR Checks** (`pr-check.yml`)
**Triggers**: Pull Requests to main  
**Purpose**: Fast feedback and quality gates

**Smart Change Detection**:
```yaml
Backend Changes: internal/**, cmd/**, *.go, go.mod, go.sum
Frontend Changes: app/**  
Docker Changes: Dockerfile, docker-compose*.yml
```

**Conditional Jobs**:
- `backend-tests`: Go tests, security scan, build check
- `frontend-tests`: Lint, type check, unit tests, build check  
- `docker-check`: Docker build verification
- `security`: Dependency review for all PRs

**Performance**: 
- ⚡ **Backend only**: ~3-4 minutes
- ⚡ **Frontend only**: ~4-5 minutes  
- ⚡ **Full stack**: ~6-8 minutes

### 2. **Main Branch CI** (`ci.yml`)
**Triggers**: Push to main, Workflow calls  
**Purpose**: Comprehensive testing of merged code

**Jobs**:
- **test**: Full test suite with coverage
- **security**: Advanced security scanning
- **docker**: Docker build and caching

**Key Features**:
- 🚫 **No PR triggers**: Eliminates double testing
- 📊 **Coverage reporting**: Full test coverage analysis
- 🔒 **Security scanning**: Gosec and dependency checks
- 🚀 **Parallel execution**: Faster overall pipeline

### 3. **Deployment** (`deploy.yml`)
**Triggers**: Push to main, Manual dispatch  
**Purpose**: Production and demo deployments

**Optimized Jobs**:
1. **ci**: Reuses CI workflow (no duplication)
2. **build-and-push**: Single Docker build → push to GHCR
3. **build-frontend**: Single frontend build → artifact upload
4. **deploy-production**: Uses pre-built artifacts (automatic on main)
5. **deploy-demo**: Uses pre-built artifacts (manual or triggered)

**Key Optimizations**:
- ✅ **Zero redundant builds**: Reuses CI artifacts
- ✅ **Parallel builds**: Docker and frontend build simultaneously  
- ✅ **Artifact reuse**: Frontend built once, deployed multiple times
- ⚡ **~50% faster**: Eliminates duplicate Docker/frontend builds

## 🎛️ **Demo Deployment Options**

### **Manual Demo Deploy**
```bash
# GitHub UI: Actions → Deploy → "Run workflow"
# Check "Deploy to demo environment"
# Check "Reset demo data" (optional)
```

### **Automatic Demo Trigger**
Demo deploys automatically when:
- Manually triggered via workflow_dispatch on main branch

### **Demo Features**
- 🎭 **Separate containers**: Runs on port 5889
- 🌱 **Fresh data**: 50 demo users with realistic bookings
- 🔧 **Dev mode**: Any 6-digit OTP works
- 📅 **Extended sessions**: 1-week JWT expiration
- 📱 **No real SMS**: Mock OTP system

## 📊 **Performance Comparison**

### **Before Optimization**:
```
PR: CI (15 min) 
Merge: CI (15 min) + Deploy (12 min) = 27 min total
Demo: Manual process
```

### **After Optimization**:
```
PR: Smart Checks (3-8 min depending on changes)
Merge: Deploy only (8-10 min, reuses CI)  
Demo: Optional (5 min, reuses builds)
```

**Savings**: ~60% reduction in CI time, ~40% reduction in deployment time

## 🔧 **Required Secrets**

```bash
SSH_PRIVATE_KEY          # SSH key for production server
CONTAINER_REGISTRY_TOKEN # GitHub token for GHCR access  
CODECOV_TOKEN           # Codecov integration (optional)
```

## 📈 **Monitoring**

### **GitHub Security Tab**
- View vulnerability scan results
- Track security improvements over time
- Get automated alerts for critical issues

### **Actions Tab**  
- Monitor workflow success rates
- View detailed build logs
- Track deployment history

## 🛠 **Local Development**

```bash
# Run the same checks locally
go test -v -race -coverprofile=coverage.out ./...
go run github.com/securecodewarrior/gosec/v2/cmd/gosec@latest ./...

# Frontend
cd app
pnpm install
pnpm run lint
pnpm run check  
pnpm run test:unit
pnpm run build
```

## 📚 **Best Practices**

1. **PR-first development**: All changes via PRs for automated testing
2. **Smart commits**: Group related changes to minimize CI runs
3. **Monitor security**: Check the Security tab regularly  
4. **Test locally**: Run checks locally before pushing
5. **Demo testing**: Use manual demo deploys for stakeholder reviews 