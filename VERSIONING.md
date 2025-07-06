# Night Owls Go - Versioning Strategy

## Overview

Night Owls Go uses **"Community CalVer"** - a Calendar-based versioning system optimized for community-focused applications with small, known user bases.

## Version Format: `YYYY.MM.PATCH`

### Examples:
- `2025.01.1` - January 2025, first release
- `2025.01.2` - January 2025, second release (hotfix)
- `2025.02.1` - February 2025, new features
- `2025.03.1` - March 2025, major update

### Why This Format?

**✅ Community-Friendly:**
- Easy to understand: "January 2025 version"
- Clear timeline for when features were added
- Predictable release schedule

**✅ Safety-First:**
- Monthly boundaries prevent rushed deployments
- Clear release cadence for proper testing
- Database migration coordination

**✅ Maintenance-Optimized:**
- Simple for small teams to manage
- No complex branching strategies needed
- Backward compatibility tracking

## Implementation

### Frontend (SvelteKit)
```json
// app/package.json
{
  "version": "2025.01.1"
}
```

### Backend (Go)
```go
// internal/config/version.go
const Version = "2025.01.1"
```

### Docker Images
Tagged with:
- `2025.01.1` (specific version)
- `latest` (most recent)
- `main-a1b2c3d` (git SHA for debugging)

## Release Process

### 1. Version Bump
**Automated (Recommended):**
```bash
./scripts/bump-version.sh 2025.07.2
```

**Manual:**
Update version in:
- `app/package.json` (source of truth for frontend)
- `internal/config/version.go` (backend version)
- `cmd/server/main.go` (swagger @version annotation)

Note: `app/src/lib/config/version.ts` automatically syncs from package.json at build time

### 2. Commit & Tag
```bash
git add .
git commit -m "chore: bump version to 2025.01.1"
git tag v2025.01.1
git push origin main --tags
```

### 3. Automatic Deployment
- GitHub Actions automatically builds and deploys
- Docker images tagged with new version
- Health endpoints show new version info

## Version Information

### Runtime Version Check
```bash
# Check backend version
curl https://mm.nightowls.app/api/health

# Response includes:
{
  "status": "ok",
  "version": "2025.01.1",
  "build": {
    "version": "2025.01.1",
    "git_sha": "a1b2c3d4...",
    "build_time": "2025-01-06T10:30:00Z",
    "go_version": "go1.24.2"
  }
}
```

### Application Startup
```
INFO Starting Night Owls Server version=2025.01.1 git_sha=a1b2c3d build_time=2025-01-06T10:30:00Z
```

## Versioning Guidelines

### PATCH Increments (Same Month)
- `2025.01.1` → `2025.01.2`
- Bug fixes
- Security patches
- Minor improvements
- Database schema fixes

### MINOR Increments (New Month)
- `2025.01.2` → `2025.02.1`
- New features
- UI improvements
- Database schema additions
- API enhancements

### MAJOR Increments (New Year)
- `2025.12.3` → `2026.01.1`
- Major rewrites
- Breaking changes
- Architecture updates
- Database migrations requiring downtime

## Rollback Strategy

### Emergency Rollback
```bash
# Rollback to previous version
docker pull ghcr.io/username/night-owls-go:2025.01.1
docker-compose down
docker-compose up -d
```

### Database Rollback
- Database migrations include `.down.sql` files
- Test rollbacks in demo environment first
- Coordinate with community for planned maintenance

## PWA Considerations

### Cache Invalidation
- Version changes trigger PWA cache refresh
- Users get notification to refresh app
- Offline functionality maintained during updates

### Update Notifications
```javascript
// Frontend shows version mismatch
if (serverVersion !== clientVersion) {
  showUpdateAvailable();
}
```

## Best Practices

### 🎯 Release Timing
- **Monthly releases**: First Monday of each month
- **Emergency patches**: Within 24 hours of critical issues
- **Community communication**: 48 hours advance notice

### 📋 Pre-Release Checklist
- [ ] Run `./scripts/bump-version.sh [new-version]` (automatically updates all version files)
- [ ] Verify `./scripts/check-version.sh` shows consistent versions
- [ ] Run full test suite (included in bump script)
- [ ] Deploy to demo environment
- [ ] Test critical user journeys
- [ ] Update CHANGELOG.md
- [ ] Notify community of upcoming release

### 🔒 Security Releases
- Use `.patch` increment for security fixes
- Deploy immediately after testing
- Include CVE references in commit messages

## Migration Path

### From Previous Versions
- `0.0.1` → `2025.01.1` (Initial Community CalVer)
- Legacy versions supported during transition
- Gradual migration of documentation

### Future Considerations
- May adopt semantic versioning if project scales
- Could switch to date-based tags for major releases
- Community feedback will guide evolution

## Monitoring

### Version Drift Detection
- Health checks include version information
- Deployment pipelines verify version consistency
- Monitoring alerts on version mismatches
- Automated version consistency tests

### Rollback Triggers
- Health check failures
- Database connectivity issues
- User-reported critical bugs
- Security vulnerabilities

## Recent Improvements (2025.07.1)

### ✅ Version Synchronization Fixes
- **Issue**: Multiple version declarations could drift apart
- **Solution**: Created automated version consistency tests
- **Files**: `internal/config/version_test.go`

### ✅ Date Logic Correction
- **Issue**: Using `2025.01.1` in July 2025 was confusing
- **Solution**: Updated to `2025.07.1` reflecting current month
- **Rationale**: CalVer should match release timing

### ✅ Swagger Integration
- **Issue**: Swagger docs showed hardcoded version `1.0`
- **Solution**: Updated `@version` annotation in `cmd/server/main.go`
- **Result**: API docs now show correct version

### ✅ Enhanced Automation
- **Issue**: Manual version management was error-prone
- **Solution**: Improved `scripts/bump-version.sh` with:
  - Multi-location version updates
  - Cross-platform compatibility (macOS/Linux)
  - Date logic validation
  - Comprehensive testing before commit
  - Automatic synchronization verification

### ✅ Automated Version Injection (2025.07.1+)
- **Issue**: Frontend config version required manual sync with package.json
- **Solution**: Vite build-time injection eliminates duplication:
  - `app/src/lib/config/version.ts` now reads from `__APP_VERSION__` constant
  - Version automatically injected from `package.json` during build
  - Only 3 source files require manual updates (down from 4)
  - Eliminates version drift between package.json and frontend config

### ✅ Testing Infrastructure
- `TestVersionConsistency`: Ensures frontend/backend versions match
- `TestVersionFormat`: Validates CalVer format compliance
- `TestBuildInfo`: Verifies build metadata structure
- Integrated into CI/CD pipeline

---

**Next Version**: `2025.08.1` (Planned for August 2025)
**Current Stable**: `2025.07.1`
**Emergency Contact**: Admin team via community channels 