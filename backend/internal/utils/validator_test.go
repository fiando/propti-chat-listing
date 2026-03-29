package utils

import (
	"strings"
	"testing"

	"github.com/fiando/propti/backend/internal/models"
)

func TestValidateMediaLimits(t *testing.T) {
	makeSlice := func(n int) []string {
		s := make([]string, n)
		for i := range s {
			s[i] = "url"
		}
		return s
	}

	tests := []struct {
		name      string
		max       int
		tier      string
		images    int
		videos    int
		wantErr   bool
		errSubstr string
	}{
		// Free tier
		{name: "free: 0 media OK", max: 3, tier: "free", images: 0, videos: 0, wantErr: false},
		{name: "free: 3 media OK", max: 3, tier: "free", images: 3, videos: 0, wantErr: false},
		{name: "free: 2 img + 1 vid OK", max: 3, tier: "free", images: 2, videos: 1, wantErr: false},
		{name: "free: 4 media rejected", max: 3, tier: "free", images: 4, videos: 0, wantErr: true, errSubstr: "free tier"},
		{name: "free: 2 img + 2 vid rejected", max: 3, tier: "free", images: 2, videos: 2, wantErr: true, errSubstr: "free tier"},

		{name: "basic: 8 media OK", max: 8, tier: "basic", images: 8, videos: 0, wantErr: false},
		{name: "basic: 9 media rejected", max: 8, tier: "basic", images: 9, videos: 0, wantErr: true, errSubstr: "basic tier"},

		{name: "premium: 15 media OK", max: 15, tier: "premium", images: 10, videos: 5, wantErr: false},
		{name: "premium: 16 media rejected", max: 15, tier: "premium", images: 16, videos: 0, wantErr: true, errSubstr: "premium tier"},
		{name: "premium: rejection message cites correct limit", max: 15, tier: "premium", images: 16, videos: 0, wantErr: true, errSubstr: "15"},

		{name: "pro: 20 media OK", max: 20, tier: "pro", images: 20, videos: 0, wantErr: false},
		{name: "pro: 21 media rejected", max: 20, tier: "pro", images: 21, videos: 0, wantErr: true, errSubstr: "pro tier"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMediaLimits(tc.max, tc.tier, makeSlice(tc.images), makeSlice(tc.videos))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Fatalf("expected error to contain %q, got %q", tc.errSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateGoogleAuthRequestAcceptsAccessTokenFallback(t *testing.T) {
	if err := ValidateGoogleAuthRequest(&models.GoogleAuthRequest{AccessToken: "google-access-token"}); err != nil {
		t.Fatalf("expected accessToken-only request to be valid, got %v", err)
	}
}

func TestNormalizeWhatsAppPhone(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "already international", input: "+628123456789", want: "+628123456789"},
		{name: "local indonesia", input: "08123456789", want: "+628123456789"},
		{name: "strip spaces and dashes", input: "+62 812-3456-789", want: "+628123456789"},
		{name: "reject missing plus", input: "628123456789", wantErr: true},
		{name: "reject non digits", input: "+62812ABCD", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeWhatsAppPhone(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestValidateWhatsAppOTPAndLinkFields(t *testing.T) {
	if err := ValidateWhatsAppLinkIdentity("user-1", "+628123456789"); err != nil {
		t.Fatalf("expected valid link identity, got %v", err)
	}
	if err := ValidateWhatsAppLinkIdentity("", "+628123456789"); err == nil {
		t.Fatal("expected missing user id error")
	}
	if err := ValidateOTPCode("123456"); err != nil {
		t.Fatalf("expected valid otp code, got %v", err)
	}
	if err := ValidateOTPCode("12345"); err == nil {
		t.Fatal("expected invalid otp length error")
	}
	if err := ValidateOTPChallengeID("challenge-1"); err != nil {
		t.Fatalf("expected valid challenge id, got %v", err)
	}
	if err := ValidateOTPChallengeID(" "); err == nil {
		t.Fatal("expected missing challenge id error")
	}
}
