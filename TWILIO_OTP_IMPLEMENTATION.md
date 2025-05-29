# Twilio OTP Implementation

This document describes the implementation of real SMS OTP verification using Twilio Verify service, replacing the previous mock OTP system.

## Overview

The system now supports both **real SMS OTP** (via Twilio) and **mock OTP** (for development/testing) with automatic fallback logic:

- **Production/Real SMS**: When Twilio credentials are configured, uses Twilio Verify v2 service for real SMS delivery
- **Development/Testing**: When Twilio credentials are missing, falls back to the original mock OTP system with outbox logging

## Architecture

### Backend Components

1. **`internal/otp/twilio.go`** - Twilio Verify client implementation
2. **`internal/service/user_service.go`** - Updated to support both Twilio and mock flows
3. **`internal/config/config.go`** - Extended with Twilio configuration
4. **`internal/api/auth_handlers.go`** - Updated response messaging

### Frontend Components

1. **Login page** (`app/src/routes/login/+page.svelte`) - Smart messaging based on OTP method
2. **Register page** (`app/src/routes/register/+page.svelte`) - Updated user feedback
3. **Auth service** (`app/src/lib/services/authService.ts`) - Unchanged, works with both flows

## Configuration

### Environment Variables

Add these to your `.env.production` file:

```bash
# Twilio Verify Configuration
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_VERIFY_SID=your_twilio_verify_sid
TWILIO_FROM_NUMBER=your_twilio_from_number  # Optional, managed by Verify service
```

### Twilio Setup

1. **Create Twilio Account**: Sign up at https://console.twilio.com
2. **Create Verify Service**: 
   - Go to Verify > Services in Twilio Console
   - Create a new service and note the Service SID (starts with `VA`)
3. **Get Credentials**:
   - Account SID and Auth Token from Console Dashboard
   - Verify Service SID from the service you created

## Implementation Details

### Twilio OTP Client (`internal/otp/twilio.go`)

```go
type Client struct {
    accountSID string
    verifySID  string
    twilio     *twilio.RestClient
}

func (c *Client) StartSMS(ctx context.Context, phoneE164 string) error
func (c *Client) Check(ctx context.Context, phoneE164, code string) (bool, error)
```

**Key Features:**
- Uses Twilio Verify v2 API for managed OTP delivery
- Automatic SMS routing and optimization
- Built-in rate limiting and fraud protection
- E.164 phone number format required

### UserService Integration

The `UserService` automatically detects which OTP method to use:

```go
// Check if Twilio is configured
if s.twilioOTP != nil {
    // Use Twilio Verify
    err = s.twilioOTP.StartSMS(ctx, phone)
} else {
    // Fall back to mock OTP
    otp, err := auth.GenerateOTP()
    s.otpStore.StoreOTP(phone, otp, duration)
}
```

### Frontend Response Handling

The frontend automatically detects the OTP method from response messages:

```typescript
if (response.message.includes('Twilio')) {
    toast.success('OTP sent via SMS! Check your phone for the verification code.');
} else if (response.message.includes('sms_outbox.log')) {
    toast.success('OTP sent! Check sms_outbox.log for the code.');
}
```

## Development vs Production

### Development Mode (`DEV_MODE=true`)

- **With Twilio configured**: Sends real SMS to your phone
- **Without Twilio**: Uses mock OTP system with `sms_outbox.log`
- **Response includes**: Helpful debugging information

### Production Mode (`DEV_MODE=false`)

- **Requires Twilio configuration** for production-ready SMS
- **No debug information** in API responses
- **Real SMS delivery** with carrier-grade reliability

## Testing

### Local Development Testing

1. **Without Twilio** (Mock flow):
   ```bash
   # Don't set Twilio environment variables
   go run ./cmd/server
   # Check sms_outbox.log for OTP codes
   ```

2. **With Twilio** (Real SMS):
   ```bash
   export TWILIO_ACCOUNT_SID=your_account_sid
   export TWILIO_AUTH_TOKEN=your_auth_token  
   export TWILIO_VERIFY_SID=your_verify_sid
   export DEV_MODE=true
   go run ./cmd/server
   # Check your phone for real SMS
   ```

### Unit Tests

```bash
# Test Twilio client
go test ./internal/otp -v

# Test service integration  
go test ./internal/service -v
```

## Security Considerations

### Twilio Security Features

- **Rate Limiting**: Automatic per-phone-number rate limiting
- **Fraud Detection**: Built-in fraud protection and blocking
- **Global Delivery**: Optimized SMS routes for international delivery
- **Compliance**: Carrier-compliant message formatting

### Implementation Security

- **Credential Protection**: Twilio credentials via environment variables only
- **E.164 Validation**: Phone numbers validated and normalized
- **JWT Integration**: OTP verification tied to secure JWT generation
- **Fallback Safety**: Mock flow available when Twilio unavailable

## Cost Optimization

- **Verify Service**: $0.05 per verification attempt
- **Global Rates**: Consistent pricing across regions
- **Failed Attempts**: No charge for invalid verification attempts
- **Development**: Use mock flow to avoid SMS costs during development

## Troubleshooting

### Common Issues

1. **"Twilio credentials not configured"**
   - Ensure all three variables are set: `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_VERIFY_SID`
   - Check variable names match exactly

2. **"Invalid phone number format"**
   - Phone numbers must be in E.164 format: `+27821234567`
   - Use the PhoneInput component which auto-formats

3. **"OTP verification failed"**
   - OTP codes expire after 10 minutes (Twilio default)
   - Codes are single-use only
   - Check phone number matches exactly

4. **SMS not received**
   - Check phone number is valid and reachable
   - Verify Twilio account has sufficient balance
   - Check Twilio Console logs for delivery status

### Logs and Monitoring

- **Backend logs**: Include Twilio operation status
- **Development mode**: Additional debug information
- **Twilio Console**: Real-time delivery and error logs

## Migration Notes

### From Mock to Production

1. Set up Twilio account and Verify service
2. Add environment variables to production config
3. Deploy updated code
4. Test with real phone number
5. Monitor Twilio Console for delivery metrics

### Backward Compatibility

- **Mock flow preserved**: Still works without Twilio credentials
- **API unchanged**: Frontend code works with both flows
- **Database schema**: No changes required
- **Test suite**: Existing tests continue to work

## Future Enhancements

### Planned Features

- **WhatsApp OTP**: Change channel to `whatsapp` 
- **Voice OTP**: Change channel to `call`
- **Email OTP**: Change channel to `email`
- **Custom templates**: Use Twilio message templates
- **Multi-language**: Localized OTP messages

### Code Examples

```go
// Switch to WhatsApp OTP
params.SetChannel("whatsapp")

// Switch to Voice OTP  
params.SetChannel("call")

// Custom message template
params.SetTemplateSid("HJ...")
```

The system is designed to easily support these additional channels with minimal code changes. 