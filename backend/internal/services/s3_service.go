package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Service wraps the AWS S3 client for presigned URL generation and object operations.
type S3Service struct {
	client         *s3.Client
	presignClient  *s3.PresignClient
	bucketName     string
	presignExpiry  time.Duration
}

// NewS3Service creates an S3Service using ambient AWS credentials.
func NewS3Service(ctx context.Context, bucketName string) (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3Service{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucketName:    bucketName,
		presignExpiry: 15 * time.Minute,
	}, nil
}

// GetPresignedUploadURL generates a presigned PUT URL for direct browser/client uploads.
func (s *S3Service) GetPresignedUploadURL(ctx context.Context, key, contentType string) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(s.presignExpiry))
	if err != nil {
		return "", fmt.Errorf("presign put object: %w", err)
	}
	return req.URL, nil
}

// GetObjectURL returns the permanent HTTPS URL for an S3 object.
// For public buckets the URL is direct; for private buckets a presigned GET URL is returned.
func (s *S3Service) GetObjectURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}
	return req.URL, nil
}

// DeleteObject removes a single object from the bucket.
func (s *S3Service) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %s: %w", key, err)
	}
	return nil
}

// ListObjects returns the keys of all objects under a given folder prefix.
func (s *S3Service) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list objects (prefix=%s): %w", prefix, err)
		}
		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}
	return keys, nil
}

// MediaKey builds the S3 key for a listing media file.
// Pattern: listings/{listingID}/{filename}
func MediaKey(listingID, filename string) string {
	return fmt.Sprintf("listings/%s/%s", listingID, filename)
}
