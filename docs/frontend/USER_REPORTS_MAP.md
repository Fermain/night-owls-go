# User Reports Map Components

This document describes the new user-facing report map components that allow users to view their submitted reports on a map.

## Components Overview

### `MyReportsWidget.svelte`
The main dashboard component that combines map view, list view, and report details in a unified widget.

**Features:**
- Toggle between map and list views
- Shows total reports and location data count
- Clickable report items to view details
- Responsive design for mobile and desktop

**Usage:**
```svelte
<script>
  import MyReportsWidget from '$lib/components/user/dashboard/MyReportsWidget.svelte';
</script>

<MyReportsWidget className="mb-4" />
```

### `UserReportsMap.svelte`
A map component specifically for displaying user's own reports with location data.

**Features:**
- Interactive map using OpenStreetMap
- Color-coded markers by severity (blue=normal, orange=suspicion, red=incident)
- Automatic bounds calculation to fit all reports
- Click handlers for report selection
- Fallback display when no location data is available

**Usage:**
```svelte
<script>
  import UserReportsMap from '$lib/components/user/dashboard/UserReportsMap.svelte';
  
  function handleReportClick(reportId) {
    console.log('Report clicked:', reportId);
  }
</script>

<UserReportsMap onReportClick={handleReportClick} />
```

### `UserReportDetail.svelte`
A modal dialog that displays detailed information about a specific report.

**Features:**
- Full report details with severity badges
- Date/time information
- Schedule information (if applicable)
- GPS coordinates and accuracy
- "View on Map" button for external map services

**Usage:**
```svelte
<script>
  import UserReportDetail from '$lib/components/user/dashboard/UserReportDetail.svelte';
  
  let showDetail = false;
  let selectedReportId = null;
  
  function openDetail(reportId) {
    selectedReportId = reportId;
    showDetail = true;
  }
  
  function closeDetail() {
    showDetail = false;
    selectedReportId = null;
  }
</script>

<UserReportDetail
  bind:open={showDetail}
  reportId={selectedReportId}
  onClose={closeDetail}
/>
```

## API Integration

### Backend Endpoint
The components use a new API endpoint for fetching user reports:

- **GET** `/api/user/reports` - Returns the current user's reports
- **Authentication**: Required (Bearer token)
- **Response**: Array of UserReport objects

### Data Structure
```typescript
interface UserReport {
  report_id: number;
  booking_id?: number;
  severity: number; // 0=Normal, 1=Suspicion, 2=Incident
  message: string;
  created_at: string;
  latitude?: number;
  longitude?: number;
  gps_accuracy?: number;
  gps_timestamp?: string;
  schedule_name?: string;
  shift_start?: string;
  shift_end?: string;
}
```

## Severity Levels

The components use a consistent severity system:

- **0 (Normal)**: Blue markers/badges - Routine observations
- **1 (Suspicion)**: Orange markers/badges - General incidents, suspicious activity
- **2 (Incident)**: Red markers/badges - Security threats, immediate attention needed

## Map Technology

- Uses `svelte-maplibre` with OpenStreetMap tiles
- Consistent with admin-side mapping for unified experience
- Automatic zoom level calculation based on GPS accuracy
- Responsive design with mobile-friendly controls

## Integration Example

The main dashboard integration shows how to add the widget:

```svelte
<!-- In app/src/routes/+page.svelte -->
{#if $userSession.isAuthenticated}
  <MyReportsWidget className="mb-4" />
{/if}
```

This provides users with immediate visibility of their reporting history and location data directly on their dashboard.

## Error Handling

- Loading states for API calls
- Error messages for failed requests
- Graceful fallbacks when no location data is available
- Timeout handling for service worker issues

## Mobile Responsiveness

All components are designed mobile-first:
- Touch-friendly map controls
- Responsive grid layouts
- Appropriate font sizes for small screens
- Collapsible/expandable sections where needed 