# Branch Protection Setup Guide

This document outlines the recommended branch protection rules for the Night Owls Control repository to ensure code quality and safe deployments.

## GitHub Repository Settings

### 1. Enable Branch Protection for `main`

Go to **Settings** → **Branches** → **Add rule** and configure:

#### Branch name pattern
```
main
```

#### Protection Rules

**✅ Require a pull request before merging**
- ✅ Require approvals: `1`
- ✅ Dismiss stale PR approvals when new commits are pushed
- ✅ Require review from code owners (if you have a CODEOWNERS file)

**✅ Require status checks to pass before merging**
- ✅ Require branches to be up to date before merging

**Required Status Checks:**
```
Backend Tests & Linting
Frontend Tests & Linting
Security Scanning
Docker Build Test
Dependency Vulnerability Check
Build Status
```

**✅ Require conversation resolution before merging**

**✅ Require signed commits** (recommended)

**✅ Include administrators** (recommended for consistency)

**✅ Restrict pushes that create files** (prevents large files)

**✅ Allow force pushes** (unchecked - prevents force pushes)

**✅ Allow deletions** (unchecked - prevents branch deletion)

### 2. Optional: Protection for `develop` branch

If you use a develop branch for staging:

#### Branch name pattern
```
develop
```

#### Protection Rules
- ✅ Require a pull request before merging
- ✅ Require status checks to pass before merging
- Required status checks: (same as main, but exclude E2E tests for faster feedback)

### 3. Environment Protection (Production)

Go to **Settings** → **Environments** → **New environment**

#### Environment name
```
production
```

#### Protection Rules
- ✅ Required reviewers: Add trusted team members
- ✅ Wait timer: `5 minutes` (gives time to cancel if needed)
- ✅ Deployment branches: `Selected branches` → `main`

## Repository Secrets

Add these secrets in **Settings** → **Secrets and variables** → **Actions**:

### Deployment Secrets
```
DEPLOY_SSH_KEY       # Private SSH key for deploy user
DEPLOY_HOST          # Server hostname (mm.nightowls.app)
DEPLOY_USER          # Username (deploy)
```

### Optional Notification Secrets
```
SLACK_WEBHOOK        # Slack webhook URL for deployment notifications
```

## Workflow Configuration

### CI Workflow (`ci.yml`)
- ✅ Runs on all pushes and PRs to `main` and `develop`
- ✅ Provides status checks for branch protection
- ✅ Includes security scanning and dependency checks
- ✅ Runs E2E tests only for main branch (performance optimization)

### Deploy Workflow (`deploy.yml`)
- ✅ Only runs on pushes to `main` branch
- ✅ Waits for CI workflow to pass
- ✅ Uses environment protection for production deployments
- ✅ Creates Docker images and deploys them
- ✅ Includes rollback on health check failure

## CODEOWNERS File (Optional)

Create `.github/CODEOWNERS` to require specific reviewers:

```
# Global owners
* @yourusername

# Backend changes
*.go @yourusername @backend-team
go.mod @yourusername @backend-team
go.sum @yourusername @backend-team

# Frontend changes
/app/ @yourusername @frontend-team

# Infrastructure changes
Dockerfile @yourusername @devops-team
docker-compose.yml @yourusername @devops-team
.github/workflows/ @yourusername @devops-team

# Deployment configuration
Caddyfile @yourusername @devops-team
deploy-docker.sh @yourusername @devops-team
```

## Pull Request Template (Optional)

Create `.github/pull_request_template.md`:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] No breaking changes

## Checklist
- [ ] Code follows the project's style guidelines
- [ ] Self-review completed
- [ ] Documentation updated (if needed)
- [ ] No sensitive information exposed

## Screenshots (if applicable)
Add screenshots here

## Related Issues
Closes #(issue number)
```

## Auto-merge Configuration (Advanced)

For trusted contributors, you can enable auto-merge on PRs that pass all checks:

1. Enable auto-merge in repository settings
2. Configure GitHub Apps like Mergify for advanced auto-merge rules

Example `.mergify.yml`:
```yaml
pull_request_rules:
  - name: Automatic merge on approval
    conditions:
      - "#approved-reviews-by>=1"
      - check-success=Build Status
      - base=main
      - author~=^(dependabot\[bot\]|yourusername)$
    actions:
      merge:
        method: squash
```

## Monitoring and Alerts

### GitHub Security Alerts
- ✅ Enable Dependabot alerts
- ✅ Enable Dependabot security updates
- ✅ Enable Code scanning alerts

### Status Monitoring
- Monitor workflow success rates
- Set up alerts for deployment failures
- Review security scan results regularly

## Best Practices Summary

1. **Always use PRs** - Even for small changes
2. **Require reviews** - At least one reviewer for main branch
3. **Status checks** - All CI checks must pass
4. **Environment protection** - Production deployments need approval
5. **Signed commits** - Verify commit authenticity
6. **Regular updates** - Keep dependencies up to date
7. **Security scanning** - Monitor for vulnerabilities
8. **Backup strategy** - Automated backups before deployment

## Quick Setup Commands

After configuring branch protection, test the setup:

```bash
# Create a test branch
git checkout -b test-branch-protection

# Make a small change
echo "# Test" >> TEST.md
git add TEST.md
git commit -m "test: branch protection setup"

# Push and create PR
git push origin test-branch-protection

# The PR should show required status checks
```

This setup ensures that:
- ✅ No direct pushes to main
- ✅ All code is reviewed
- ✅ All tests pass before merge
- ✅ Security scans are clean
- ✅ Deployments are controlled and monitored 