# Photo Support Feature Specification

**Project**: Night Owls Go - Community Watch Scheduler  
**Feature**: Photo Evidence Attachment for Incident Reports  
**Status**: Proposed  
**Priority**: Medium  
**Estimated Effort**: 2-3 sprints  

## 🎯 **Executive Summary**

Enable community watch volunteers to attach photo evidence directly to incident reports, enhancing documentation quality and providing visual context for security incidents, property damage, and suspicious activities.

## 📝 **Background**

Currently, the help dialog promises photo evidence functionality (*"Photo evidence - attach images directly to reports"*), but this feature is not implemented. Users expect this capability based on the help documentation, creating a gap between promised and delivered functionality.

## 🎯 **Goals**

### Primary Goals
- **Evidence Documentation**: Enable visual evidence capture for incident reports
- **User Experience**: Fulfill help dialog promises and meet user expectations  
- **Security Enhancement**: Improve incident documentation quality for community safety

### Secondary Goals
- **Mobile Optimization**: Seamless photo capture on mobile devices
- **Storage Efficiency**: Optimize image storage and delivery
- **Privacy Protection**: Ensure appropriate handling of potentially sensitive images

## 🔧 **Technical Requirements**

### Backend Changes

#### Database Schema
```sql
-- Add photo support to reports table
ALTER TABLE reports ADD COLUMN photo_count INTEGER DEFAULT 0;

-- Create report_photos table
CREATE TABLE report_photos (
    photo_id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL REFERENCES reports(report_id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    original_filename TEXT,
    file_size_bytes INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    width_pixels INTEGER,
    height_pixels INTEGER,
    upload_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    storage_path TEXT NOT NULL,
    thumbnail_path TEXT,
    
    -- Security and validation
    checksum_sha256 TEXT NOT NULL,
    is_processed BOOLEAN DEFAULT FALSE,
    
    CONSTRAINT valid_mime_type CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/webp')),
    CONSTRAINT valid_file_size CHECK (file_size_bytes > 0 AND file_size_bytes <= 10485760) -- 10MB max
);

CREATE INDEX idx_report_photos_report_id ON report_photos(report_id);
CREATE INDEX idx_report_photos_upload_timestamp ON report_photos(upload_timestamp);
```

#### API Endpoints
```go
// New endpoints to add
POST   /api/reports/{id}/photos          // Upload photo to existing report
DELETE /api/reports/{id}/photos/{photoId} // Delete photo (admin only)
GET    /api/reports/{id}/photos/{photoId} // Get photo (authenticated users)
GET    /api/reports/{id}/photos/{photoId}/thumbnail // Get thumbnail
```

#### File Storage Strategy
```
Option 1: Local File Storage (Recommended for MVP)
- Store in `./data/report-photos/YYYY/MM/DD/`
- Filename format: `{report_id}_{photo_id}_{timestamp}_{checksum}.{ext}`
- Generate thumbnails at upload time (150x150px)

Option 2: Cloud Storage (Future Enhancement)
- AWS S3 or similar cloud storage
- CDN integration for faster delivery
- Automatic backup and versioning
```

#### Go Service Layer
```go
type PhotoService struct {
    querier     db.Querier
    logger      *slog.Logger
    storagePath string
    maxFileSize int64
}

type PhotoUploadRequest struct {
    File       multipart.File
    Header     *multipart.FileHeader
    ReportID   int64
    UserID     int64
}

type PhotoMetadata struct {
    PhotoID        int64     `json:"photo_id"`
    ReportID       int64     `json:"report_id"`
    Filename       string    `json:"filename"`
    OriginalName   string    `json:"original_filename"`
    FileSizeBytes  int64     `json:"file_size_bytes"`
    MimeType       string    `json:"mime_type"`
    Width          *int      `json:"width_pixels,omitempty"`
    Height         *int      `json:"height_pixels,omitempty"`
    UploadTime     time.Time `json:"upload_timestamp"`
    ThumbnailURL   string    `json:"thumbnail_url"`
    PhotoURL       string    `json:"photo_url"`
}

// Key methods
func (ps *PhotoService) UploadPhoto(ctx context.Context, req PhotoUploadRequest) (*PhotoMetadata, error)
func (ps *PhotoService) DeletePhoto(ctx context.Context, photoID, userID int64) error
func (ps *PhotoService) GetPhotoMetadata(ctx context.Context, reportID int64) ([]PhotoMetadata, error)
func (ps *PhotoService) GenerateThumbnail(imagePath string) error
```

### Frontend Changes

#### Photo Upload Component
```typescript
// PhotoUploadArea.svelte
interface PhotoUploadProps {
    reportId?: number;
    maxPhotos?: number;
    maxFileSizeBytes?: number;
    onPhotosUploaded?: (photos: PhotoMetadata[]) => void;
    onError?: (error: string) => void;
}

// Features:
// - Drag & drop interface
// - Camera capture on mobile
// - Image preview with thumbnails
// - Progress indicators during upload
// - Automatic compression/resizing
// - EXIF data stripping for privacy
```

#### Report Dialog Updates
```typescript
// Add to ReportDialog.svelte
interface ReportData {
    severity: number;
    message: string;
    latitude: number | null;
    longitude: number | null;
    accuracy: number | null;
    photos?: File[]; // New field
}
```

#### API Service Updates
```typescript
// UserApiService.ts
export interface CreateReportRequest {
    severity: number;
    message: string;
    latitude?: number | null;
    longitude?: number | null;
    accuracy?: number | null;
    photos?: File[]; // New field
}

async createShiftReportWithPhotos(
    bookingId: number, 
    data: CreateReportRequest
): Promise<ReportResponse>

async uploadReportPhoto(
    reportId: number, 
    file: File
): Promise<PhotoMetadata>
```

## 🔐 **Security Considerations**

### File Validation
- **File Type**: Only allow JPEG, PNG, WebP
- **File Size**: Maximum 10MB per photo
- **Quantity**: Maximum 5 photos per report
- **Content Scanning**: Basic malware/virus scanning
- **EXIF Stripping**: Remove GPS and camera metadata for privacy

### Access Control
- **View Photos**: Only authenticated users
- **Upload Photos**: Only report creator within 24 hours of report creation
- **Delete Photos**: Only admins or report creator within 1 hour of upload
- **Admin Access**: Full access to all photos for investigation purposes

### Storage Security
- **Path Traversal Protection**: Validate and sanitize file paths
- **Direct Access Prevention**: Photos not accessible via direct URL
- **Encryption at Rest**: Consider encrypting stored images
- **Backup Strategy**: Regular backups with retention policy

## 📱 **User Experience Design**

### Mobile-First Approach
```
Photo Capture Flow:
1. User taps "Add Photo" in report dialog
2. Options: "Take Photo" | "Choose from Gallery"
3. Camera opens with overlay guide for evidence capture
4. Auto-compress and preview image
5. Option to retake or add another photo
6. Submit report with photos attached
```

### Desktop Experience
```
Desktop Upload Flow:
1. Drag & drop area in report dialog
2. Click to select files from computer
3. Preview thumbnails with remove option
4. Progress indicator during upload
5. Success confirmation with photo count
```

### Photo Viewing
```
Admin Report View:
- Thumbnail grid in report details
- Click to view full-size image
- Download original image option
- Photo metadata display (size, upload time)

User Report History:
- Small thumbnails in report list
- Click to expand photo gallery
- Share photo functionality
```

## 🚀 **Implementation Phases**

### Phase 1: MVP (Sprint 1-2)
- [ ] Database schema migration
- [ ] Basic file upload API endpoint
- [ ] Simple photo upload component
- [ ] Local file storage implementation
- [ ] Admin photo viewing capability

### Phase 2: Enhanced UX (Sprint 3)
- [ ] Thumbnail generation
- [ ] Mobile camera integration
- [ ] Drag & drop interface
- [ ] Image compression/resizing
- [ ] Photo gallery component

### Phase 3: Security & Polish (Sprint 4)
- [ ] EXIF data stripping
- [ ] Advanced file validation
- [ ] Access control refinement
- [ ] Error handling improvements
- [ ] Performance optimization

### Phase 4: Advanced Features (Future)
- [ ] Cloud storage integration
- [ ] Image annotation tools
- [ ] Automatic image analysis
- [ ] Bulk photo operations
- [ ] Photo sharing with authorities

## 📊 **Success Metrics**

### Usage Metrics
- **Photo Upload Rate**: % of reports with photos attached
- **Average Photos per Report**: Track engagement level
- **Upload Success Rate**: Technical reliability measure
- **Mobile vs Desktop Usage**: Platform preference insights

### Quality Metrics
- **Storage Efficiency**: Average file size after compression
- **Load Performance**: Photo viewing speed
- **Error Rate**: Failed uploads or corrupted images
- **User Satisfaction**: Help dialog accuracy improvement

## 🔧 **Configuration**

### Environment Variables
```bash
# Photo storage configuration
PHOTO_STORAGE_PATH=./data/report-photos
PHOTO_MAX_FILE_SIZE_MB=10
PHOTO_MAX_PER_REPORT=5
PHOTO_ALLOWED_TYPES=image/jpeg,image/png,image/webp
PHOTO_THUMBNAIL_SIZE=150
PHOTO_COMPRESSION_QUALITY=85

# Security settings
PHOTO_VIRUS_SCANNING_ENABLED=true
PHOTO_EXIF_STRIPPING_ENABLED=true
PHOTO_ACCESS_LOG_ENABLED=true
```

### Feature Flags
```go
type PhotoConfig struct {
    Enabled              bool   `env:"PHOTO_FEATURE_ENABLED" default:"false"`
    StoragePath          string `env:"PHOTO_STORAGE_PATH" default:"./data/report-photos"`
    MaxFileSizeMB        int    `env:"PHOTO_MAX_FILE_SIZE_MB" default:"10"`
    MaxPhotosPerReport   int    `env:"PHOTO_MAX_PER_REPORT" default:"5"`
    ThumbnailSize        int    `env:"PHOTO_THUMBNAIL_SIZE" default:"150"`
    CompressionQuality   int    `env:"PHOTO_COMPRESSION_QUALITY" default:"85"`
    VirusScanningEnabled bool   `env:"PHOTO_VIRUS_SCANNING_ENABLED" default:"true"`
    EXIFStrippingEnabled bool   `env:"PHOTO_EXIF_STRIPPING_ENABLED" default:"true"`
}
```

## 🧪 **Testing Strategy**

### Unit Tests
- File validation logic
- Image processing functions
- Storage path handling
- Security controls

### Integration Tests
- End-to-end photo upload flow
- API endpoint functionality
- Database integrity
- File system operations

### Security Tests
- Malicious file upload attempts
- Path traversal attack vectors
- Access control validation
- File size limit enforcement

### Performance Tests
- Large file upload handling
- Concurrent upload scenarios
- Storage space management
- Thumbnail generation speed

## 📋 **Deployment Checklist**

### Pre-Deployment
- [ ] Database migration tested
- [ ] Storage directory permissions configured
- [ ] Environment variables set
- [ ] Feature flag configuration
- [ ] Backup strategy verified

### Post-Deployment
- [ ] Upload functionality verified
- [ ] Photo viewing confirmed
- [ ] Error logging monitored
- [ ] Performance metrics tracked
- [ ] Help dialog updated

## 🔄 **Migration Strategy**

### Existing Users
- No impact on existing reports (photos are optional)
- Help dialog updated to reflect actual capabilities
- Gradual rollout with feature flag control

### Data Migration
- No existing data migration required
- New tables created with migration script
- Storage directories created automatically

## 📚 **Documentation Updates**

### User Documentation
- [ ] Update help dialog to reflect photo capabilities
- [ ] Add photo upload guide to user manual
- [ ] Create mobile app usage instructions

### Developer Documentation
- [ ] API endpoint documentation
- [ ] Database schema updates
- [ ] Configuration guide
- [ ] Security best practices

## 🎯 **Success Criteria**

### Technical Success
- ✅ Photo upload works reliably on mobile and desktop
- ✅ Images are properly compressed and stored
- ✅ Security validation prevents malicious uploads
- ✅ Performance impact is minimal

### User Success
- ✅ Help dialog promises match reality
- ✅ Photo capture is intuitive and fast
- ✅ Visual evidence enhances report quality
- ✅ Admin investigation capabilities improved

### Business Success
- ✅ Increased user engagement with reporting
- ✅ Higher quality incident documentation
- ✅ Enhanced community safety capabilities
- ✅ Positive user feedback on new functionality

---

**Next Steps**:
1. Review and approve feature specification
2. Estimate development effort for each phase
3. Prioritize implementation phases
4. Begin Phase 1 development
5. Update help dialog to remove false claims (immediate action) 