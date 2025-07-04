# Night Owls Go Security Hardening Tasks

## Overview
This document tracks the implementation of security fixes identified in the security audit report.
**Status**: 🔴 Critical | 🟡 High Risk | 🟢 Medium Priority

---

## 🔴 CRITICAL VULNERABILITIES (Immediate Action Required)

### Task 1: OTP Brute Force Protection
**Priority**: Critical | **Status**: ❌ Not Started

**Issue**: No rate limiting on OTP verification attempts
**Impact**: Attackers can brute force 6-digit OTPs (1M combinations)

**Implementation Plan**:
- [ ] Create OTP attempt tracking in database
- [ ] Add rate limiting middleware for OTP endpoints  
- [ ] Implement exponential backoff
- [ ] Add account lockout after multiple failed attempts
- [ ] Log suspicious OTP activity

**Files to Modify**:
- `internal/auth/otp.go`
- `internal/service/user_service.go`
- `internal/api/auth_handlers.go`
- Database migration for OTP attempts table

---

### Task 2: JWT Secret Hardening
**Priority**: Critical | **Status**: ✅ COMPLETED

**Issue**: Hardcoded default JWT secret in production
**Impact**: Token forgery if default secret used

**Implementation Plan**:
- [x] Add startup validation for default JWT secret
- [x] Force application exit if default secret detected in production
- [x] Add environment variable validation
- [x] Update deployment documentation
- [x] Generate secure random secret for production

**Files Modified**:
- `internal/config/config.go` - Added ValidateSecurityConfig() and isProductionEnvironment()
- `cmd/server/main.go` - Added security validation at startup

---

### Task 3: Dev Mode Security Controls
**Priority**: Critical | **Status**: ✅ COMPLETED

**Issue**: Authentication bypass via `/auth/dev-login` endpoint
**Impact**: Complete authentication bypass if misconfigured

**Implementation Plan**:
- [x] Add production environment detection
- [x] Disable dev endpoints in production builds
- [x] Add compile-time flags for dev features
- [x] Implement environment-based route registration
- [x] Add startup warnings for dev mode

**Files Modified**:
- `internal/config/config.go` - Added dev mode validation in ValidateSecurityConfig()
- `cmd/server/main.go` - Existing dev mode conditional already present
- Application now exits if dev mode enabled in production

---

### Task 4: Account Lockout Policy
**Priority**: Critical | **Status**: ❌ Not Started

**Issue**: Unlimited failed authentication attempts allowed
**Impact**: Persistent brute force attacks possible

**Implementation Plan**:
- [ ] Create failed attempts tracking table
- [ ] Implement progressive lockout (5 attempts = 30min lock)
- [ ] Add lockout status to user queries
- [ ] Implement lockout bypass for admins
- [ ] Add lockout notifications

**Files to Modify**:
- Database migration for failed attempts
- `internal/service/user_service.go`
- `internal/api/auth_handlers.go`

---

## 🟡 HIGH RISK VULNERABILITIES

### Task 5: Secure JWT Storage
**Priority**: High Risk | **Status**: ❌ Not Started

**Issue**: JWT stored in localStorage, vulnerable to XSS
**Impact**: Token theft via malicious scripts

**Implementation Plan**:
- [ ] Replace localStorage with HTTP-only cookies
- [ ] Implement SameSite=Strict cookie policy
- [ ] Add Secure flag for HTTPS
- [ ] Update frontend authentication flow
- [ ] Add CSRF protection for cookie-based auth

**Files to Modify**:
- `app/src/lib/stores/authStore.ts`
- `internal/api/auth_handlers.go`
- Frontend authentication components

---

### Task 6: User Enumeration Prevention
**Priority**: High Risk | **Status**: ❌ Not Started

**Issue**: Different error messages reveal valid phone numbers
**Impact**: Phone number enumeration for targeted attacks

**Implementation Plan**:
- [ ] Standardize all authentication error messages
- [ ] Return generic "invalid credentials" for all auth failures
- [ ] Implement constant-time responses
- [ ] Add timing attack protection
- [ ] Update API documentation

**Files to Modify**:
- `internal/api/auth_handlers.go`
- `internal/service/user_service.go`

---

### Task 7: Security Headers Implementation
**Priority**: High Risk | **Status**: ❌ Not Started

**Issue**: Missing Content Security Policy and security headers
**Impact**: XSS and clickjacking vulnerabilities

**Implementation Plan**:
- [ ] Add Content Security Policy header
- [ ] Implement X-Frame-Options: DENY
- [ ] Add X-Content-Type-Options: nosniff
- [ ] Set Strict-Transport-Security header
- [ ] Configure Referrer-Policy

**Files to Modify**:
- `app/src/app.html`
- Backend middleware for security headers
- `internal/api/middleware.go` (if exists)

---

## 🟢 MEDIUM PRIORITY IMPROVEMENTS

### Task 8: Error Message Standardization
**Priority**: Medium | **Status**: ❌ Not Started

**Implementation Plan**:
- [ ] Create standard error response structure
- [ ] Implement error message enum/constants
- [ ] Remove detailed error information from responses
- [ ] Keep detailed errors in server logs only

---

### Task 9: Constant-Time Comparison
**Priority**: Medium | **Status**: ❌ Not Started

**Implementation Plan**:
- [ ] Replace string comparison with crypto/subtle
- [ ] Apply to OTP verification
- [ ] Apply to password/token comparisons
- [ ] Add timing attack tests

---

### Task 10: Enhanced Account Lockout
**Priority**: Medium | **Status**: ❌ Not Started

**Implementation Plan**:
- [ ] Implement progressive lockout durations
- [ ] Add admin unlock capabilities
- [ ] Email notifications for lockouts
- [ ] Suspicious activity monitoring

---

## Implementation Order

### Phase 1: Critical Security Fixes
1. **JWT Secret Hardening** (Task 2) - Fastest to implement
2. **Dev Mode Controls** (Task 3) - High impact, low effort  
3. **OTP Rate Limiting** (Task 1) - Core authentication security
4. **Account Lockout** (Task 4) - Prevents brute force

### Phase 2: High Risk Mitigations
5. **Security Headers** (Task 7) - Quick frontend hardening
6. **Error Message Standardization** (Task 6 & 8) - Prevents enumeration
7. **Secure JWT Storage** (Task 5) - Frontend security improvement

### Phase 3: Additional Hardening
8. **Constant-Time Comparison** (Task 9) - Timing attack prevention
9. **Enhanced Lockout Features** (Task 10) - Advanced protection

---

## Testing Strategy

### Security Testing Requirements
- [ ] OTP brute force testing with rate limits
- [ ] JWT secret validation in different environments  
- [ ] Dev mode endpoint accessibility testing
- [ ] Account lockout functionality testing
- [ ] XSS testing with new security headers
- [ ] Timing attack testing for constant-time operations

### Automated Security Tests
- [ ] Unit tests for rate limiting logic
- [ ] Integration tests for lockout policies
- [ ] Security header validation tests
- [ ] Authentication bypass prevention tests

---

## Documentation Updates Required
- [ ] Security architecture documentation
- [ ] Deployment security checklist
- [ ] Environment configuration guide
- [ ] Incident response procedures
- [ ] Security monitoring setup

---

## Completion Criteria
- [ ] All critical vulnerabilities addressed
- [ ] Security tests passing
- [ ] Documentation updated
- [ ] Production deployment validated
- [ ] Security audit re-run shows improvements 