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

func TestResolveDynamoDBClientConfigFromEnv_UsesAmbientAWSConfigWhenEndpointUnset(t *testing.T) {
	t.Setenv("DYNAMODB_ENDPOINT_URL", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	cfg := resolveDynamoDBClientConfigFromEnv()

	if cfg.EndpointURL != "" {
		t.Fatalf("expected empty endpoint URL, got %q", cfg.EndpointURL)
	}
	if cfg.Region != "" {
		t.Fatalf("expected empty region, got %q", cfg.Region)
	}
	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		t.Fatalf("expected ambient credentials, got %q / %q", cfg.AccessKeyID, cfg.SecretAccessKey)
	}
}

func TestResolveDynamoDBClientConfigFromEnv_ProvidesLocalDefaults(t *testing.T) {
	t.Setenv("DYNAMODB_ENDPOINT_URL", "http://127.0.0.1:8000")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	cfg := resolveDynamoDBClientConfigFromEnv()

	if cfg.EndpointURL != "http://127.0.0.1:8000" {
		t.Fatalf("expected local endpoint URL, got %q", cfg.EndpointURL)
	}
	if cfg.Region != "ap-southeast-1" {
		t.Fatalf("expected default local region, got %q", cfg.Region)
	}
	if cfg.AccessKeyID != "local" || cfg.SecretAccessKey != "local" {
		t.Fatalf("expected local credentials defaults, got %q / %q", cfg.AccessKeyID, cfg.SecretAccessKey)
	}
}
