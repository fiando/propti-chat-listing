package services

import "testing"

func TestS3KeyHelpersBuildExpectedPaths(t *testing.T) {
	if got := BuildStagingKey("user-1", "session-1", "upload.jpeg"); got != "staging/user-1/session-1/upload.jpeg" {
		t.Fatalf("unexpected staging key: %q", got)
	}
	if got := BuildPermanentKey("listing-1", "image-1"); got != "listings/listing-1/image-1" {
		t.Fatalf("unexpected permanent key: %q", got)
	}
	if got := BuildThumbnailKey("listing-1", "image-1"); got != "thumbnails/listing-1/image-1" {
		t.Fatalf("unexpected thumbnail key: %q", got)
	}
	if got := BuildRejectedKey("listing-1", "image-1"); got != "rejected/listing-1/image-1" {
		t.Fatalf("unexpected rejected key: %q", got)
	}
}
