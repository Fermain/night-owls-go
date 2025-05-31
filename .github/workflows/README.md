# GitHub Actions Workflows

This repository uses GitHub Actions for CI/CD with a focus on reliability, security, and efficiency.

## 🔄 Workflow Overview

### 1. **CI Workflow** (`ci.yml`)
**Triggers**: Push to main, Pull Requests  
**Purpose**: Comprehensive testing, building, and security scanning

**Jobs**:
- **test**: Runs Go tests with coverage, builds backend and frontend
- **security**: Dependency scanning and vulnerability detection  
- **docker**: Docker image build verification

**Key Features**:
- ✅ Test coverage reporting with Codecov
- 🔒 Security scanning with Gosec and Trivy
- 📦 Dependency caching for faster builds
- 🚀 Parallel job execution
- 🧪 Frontend testing and linting

### 2. **Deploy Workflow** (`deploy.yml`)
**Triggers**: Push to main branch, Manual dispatch  
**Purpose**: Build and deploy to production

**Jobs**:
1. **ci**: Reuses the CI workflow (ensures all tests pass)
2. **build**: Builds and pushes Docker image to GHCR
3. **frontend**: Builds frontend and uploads as artifact
4. **deploy**: Deploys to production server

**Key Features**:
- 🔗 Depends on CI passing first
- 🐳 Multi-stage Docker build with caching
- 📦 Artifact-based frontend deployment
- 🏥 Health checks and rollback capability
- 🔒 Secure deployment with SSH keys

### 3. **Security Monitoring** (`security.yml`)
**Triggers**: Daily schedule (6 AM UTC), Manual dispatch, Dependency changes  
**Purpose**: Continuous security monitoring

**Jobs**:
- **security-scan**: Vulnerability scanning of codebase and Docker images
- **dependency-updates**: Checks for outdated dependencies

## 🚀 Improvements Made

### **Previous Issues Fixed**:

1. ❌ **Version Mismatch**: CI used Go 1.24, but go.mod specified 1.24.2
   - ✅ **Fixed**: Aligned versions exactly

2. ❌ **No Security Scanning**: Missing vulnerability detection
   - ✅ **Added**: Gosec, Trivy, and dependency review

3. ❌ **Inefficient Caching**: No dependency caching
   - ✅ **Added**: Go modules and pnpm cache

4. ❌ **No Test Coverage**: Tests ran without coverage reporting
   - ✅ **Added**: Coverage reports with Codecov integration

5. ❌ **Monolithic Deploy**: 185-line single job
   - ✅ **Split**: Into logical, parallel jobs

6. ❌ **No CI Dependency**: Deploy could run without CI passing
   - ✅ **Fixed**: Deploy depends on CI completion

7. ❌ **Missing Frontend Testing**: Only type checking
   - ✅ **Added**: Linting and unit tests

### **New Features**:

- 🔒 **Security Tab Integration**: Vulnerability results in GitHub Security tab
- 📊 **Coverage Tracking**: Test coverage trends over time  
- 🚨 **Daily Security Scans**: Proactive vulnerability detection
- 📦 **Dependency Monitoring**: Automated update notifications
- ⚡ **Performance**: Faster builds with better caching
- 🎯 **Artifact Management**: Cleaner frontend deployment

## 🔧 Required Secrets

Add these to your repository secrets:

```bash
SSH_PRIVATE_KEY          # SSH key for production server
CONTAINER_REGISTRY_TOKEN # GitHub token for GHCR access
CODECOV_TOKEN           # Codecov integration (optional)
```

## 📈 Monitoring

### **GitHub Security Tab**
- View vulnerability scan results
- Track security improvements over time
- Get automated alerts for critical issues

### **Actions Tab**  
- Monitor workflow success rates
- View detailed build logs
- Track deployment history

### **Codecov Dashboard** (if configured)
- Track test coverage trends
- Identify untested code areas
- Coverage diff in pull requests

## 🛠 Local Development

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

## 📚 Best Practices

1. **Always run CI first**: Deploy workflow depends on CI passing
2. **Use caching**: Dependencies are cached for faster builds  
3. **Monitor security**: Check the Security tab regularly
4. **Review dependencies**: Update outdated packages regularly
5. **Test locally**: Run the same checks locally before pushing 