// Package otp provides Twilio Verify v2 SMS OTP functionality
package otp

import (
	"context"

	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

// Client wraps Twilio Verify service for OTP operations
type Client struct {
	accountSID string
	verifySID  string
	twilio     *twilio.RestClient
}

// New returns a ready-to-use Twilio OTP client
func New(accountSID, authToken, verifySID string) *Client {
	// Initialize Twilio client with credentials
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &Client{
		accountSID: accountSID,
		verifySID:  verifySID,
		twilio:     client,
	}
}

// StartSMS sends an OTP verification code via SMS
func (c *Client) StartSMS(ctx context.Context, phoneE164 string) error {
	params := &verify.CreateVerificationParams{}
	params.SetChannel("sms")
	params.SetTo(phoneE164)

	_, err := c.twilio.VerifyV2.CreateVerification(c.verifySID, params)
	return err
}

// Check verifies the user-supplied OTP code
func (c *Client) Check(ctx context.Context, phoneE164, code string) (bool, error) {
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(phoneE164)
	params.SetCode(code)

	resp, err := c.twilio.VerifyV2.CreateVerificationCheck(c.verifySID, params)
	if err != nil {
		return false, err
	}

	// Status can be "pending", "approved", or "cancelled/expired"
	return resp.Status != nil && *resp.Status == "approved", nil
} 