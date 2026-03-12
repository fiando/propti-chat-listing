package utils

import (
	"strings"
	"testing"
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
		isPremium bool
		images    int
		videos    int
		wantErr   bool
		errSubstr string
	}{
		// Free tier
		{name: "free: 0 media OK", isPremium: false, images: 0, videos: 0, wantErr: false},
		{name: "free: 3 media OK", isPremium: false, images: 3, videos: 0, wantErr: false},
		{name: "free: 2 img + 1 vid OK", isPremium: false, images: 2, videos: 1, wantErr: false},
		{name: "free: 4 media rejected", isPremium: false, images: 4, videos: 0, wantErr: true, errSubstr: "free tier"},
		{name: "free: 2 img + 2 vid rejected", isPremium: false, images: 2, videos: 2, wantErr: true, errSubstr: "free tier"},

		// Premium tier
		{name: "premium: 0 media OK", isPremium: true, images: 0, videos: 0, wantErr: false},
		{name: "premium: 3 media OK", isPremium: true, images: 3, videos: 0, wantErr: false},
		{name: "premium: 30 media OK", isPremium: true, images: 30, videos: 0, wantErr: false},
		{name: "premium: 20 img + 10 vid OK", isPremium: true, images: 20, videos: 10, wantErr: false},
		{name: "premium: 31 media rejected", isPremium: true, images: 31, videos: 0, wantErr: true, errSubstr: "premium tier"},
		{name: "premium: 25 img + 6 vid rejected", isPremium: true, images: 25, videos: 6, wantErr: true, errSubstr: "premium tier"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMediaLimits(tc.isPremium, makeSlice(tc.images), makeSlice(tc.videos))
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
