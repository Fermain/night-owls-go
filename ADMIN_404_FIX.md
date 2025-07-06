# Admin 404 Fix - Production Deployment Guide

## Issue
Admin pages returning 404 errors in production while regular owl-facing pages work fine.

## Root Cause
- SSR is disabled globally (`export const ssr = false` in root layout)
- Admin routes use server-side authentication in `+layout.server.ts`
- With SSR disabled, server layouts are ignored in production static builds
- Result: Admin authentication never runs, routes fail to load

## Fix Applied
Moved admin authentication from server-side to client-side in `app/src/routes/admin/+layout.ts`

## Deployment Steps

### Option 1: Quick Deploy (Current Fix)
1. Pull latest changes: `git pull`
2. Build frontend: `cd app && npm run build`
3. Deploy the `app/build` directory to production `/srv`
4. Test admin routes immediately

### Option 2: Full Rebuild
1. Pull latest changes
2. Run full build process:
   ```bash
   cd app
   npm install
   npm run build
   ```
3. Deploy build output
4. Restart Caddy if needed: `sudo systemctl reload caddy`

## Testing
1. Visit `/admin/users` - should redirect to login if not authenticated
2. Login as admin user
3. Visit admin pages - should load without 404s
4. Check browser console for any client-side errors

## Long-term Solutions

### Option A: Enable SSR
- Remove `export const ssr = false` from root layout
- Switch from `adapter-static` to `adapter-node`
- Deploy as Node.js app instead of static files
- Update Caddy to proxy to Node server

### Option B: Hybrid Approach
- Keep SSR disabled for public pages
- Enable SSR only for admin routes
- Requires more complex configuration

### Option C: Full Client-Side Auth
- Keep current approach but enhance:
  - Add loading states during auth checks
  - Implement auth state management (stores)
  - Add route guards for all protected routes

## Monitoring
- Check Caddy logs: `/var/log/caddy/production.log`
- Monitor for 404s on admin routes
- Watch for authentication failures

## Rollback Plan
If issues persist:
1. Revert commit: `git revert dd677ff`
2. Rebuild and redeploy
3. Consider enabling SSR temporarily