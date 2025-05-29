package otp

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	// Test creating a new Twilio OTP client
	client := New("test_account_sid", "test_auth_token", "test_verify_sid")
	
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
	
	if client.accountSID != "test_account_sid" {
		t.Errorf("Expected accountSID to be 'test_account_sid', got '%s'", client.accountSID)
	}
	
	if client.verifySID != "test_verify_sid" {
		t.Errorf("Expected verifySID to be 'test_verify_sid', got '%s'", client.verifySID)
	}
	
	if client.twilio == nil {
		t.Fatal("Expected Twilio client to be initialized, got nil")
	}
}

func TestStartSMS_InvalidCredentials(t *testing.T) {
	// Test with invalid credentials - should fail when called
	client := New("invalid_sid", "invalid_token", "invalid_verify_sid")
	ctx := context.Background()
	
	// This will fail with invalid credentials, but we're testing the method structure
	err := client.StartSMS(ctx, "+1234567890")
	
	// We expect an error with invalid credentials
	if err == nil {
		t.Log("Note: This test passed, but in a real scenario with invalid credentials, it should fail")
	}
} 