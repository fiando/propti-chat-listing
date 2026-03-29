package repository

import (
	"os"
	"testing"
)

func TestResolveOTPChallengesTableEnv_PrefersPrimaryEnv(t *testing.T) {
	t.Setenv("DYNAMODB_OTP_CHALLENGES_TABLE", "otp-primary")
	t.Setenv("DYNAMODB_WHATSAPP_OTP_TABLE", "otp-legacy")

	if got := resolveOTPChallengesTableEnv(); got != "otp-primary" {
		t.Fatalf("expected primary env value, got %q", got)
	}
}

func TestResolveOTPChallengesTableEnv_FallsBackToLegacyEnv(t *testing.T) {
	_ = os.Unsetenv("DYNAMODB_OTP_CHALLENGES_TABLE")
	t.Setenv("DYNAMODB_WHATSAPP_OTP_TABLE", "otp-legacy")

	if got := resolveOTPChallengesTableEnv(); got != "otp-legacy" {
		t.Fatalf("expected legacy env fallback, got %q", got)
	}
}

func TestResolveOTPChallengesTableEnv_UsesDefaultWhenUnset(t *testing.T) {
	_ = os.Unsetenv("DYNAMODB_OTP_CHALLENGES_TABLE")
	_ = os.Unsetenv("DYNAMODB_WHATSAPP_OTP_TABLE")

	if got := resolveOTPChallengesTableEnv(); got != "propti-otp-challenges" {
		t.Fatalf("expected default table name, got %q", got)
	}
}
