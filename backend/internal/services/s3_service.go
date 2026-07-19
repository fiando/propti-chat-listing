package services

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MediaObjectHead struct {
	ContentType string
	SizeBytes   int64
}

type MediaStorage interface {
	GetPresignedUploadURL(ctx context.Context, key, contentType string) (string, error)
	GetSignedDownloadURL(ctx context.Context, key string) (string, error)
	BuildPublicURL(key string) string
	HeadObject(ctx context.Context, key string) (*MediaObjectHead, error)
	CopyObject(ctx context.Context, sourceKey, destinationKey string) error
	DeleteObject(ctx context.Context, key string) error
	GetObjectBytes(ctx context.Context, key string) ([]byte, error)
}

type S3Service struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
	presignExpiry time.Duration
	region        string
	publicBaseURL string
}

func NewS3Service(ctx context.Context, bucketName string) (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Options := []func(*s3.Options){}
	s3Endpoint := os.Getenv("AWS_ENDPOINT_URL_S3")
	if s3Endpoint != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Endpoint)
			o.UsePathStyle = true
		})
	}
	client := s3.NewFromConfig(cfg, s3Options...)
	publicBaseURL := strings.TrimRight(os.Getenv("PUBLIC_MEDIA_BASE_URL"), "/")
	if publicBaseURL == "" {
		if s3Endpoint != "" {
			publicBaseURL = fmt.Sprintf("%s/%s", s3Endpoint, bucketName)
		} else {
			publicBaseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucketName, cfg.Region)
		}
	}

	return &S3Service{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucketName:    bucketName,
		presignExpiry: 15 * time.Minute,
		region:        cfg.Region,
		publicBaseURL: publicBaseURL,
	}, nil
}

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

func (s *S3Service) GetSignedDownloadURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}
	return req.URL, nil
}

func (s *S3Service) GetObjectURL(ctx context.Context, key string) (string, error) {
	return s.GetSignedDownloadURL(ctx, key)
}

func (s *S3Service) BuildPublicURL(key string) string {
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s.publicBaseURL, strings.TrimLeft(key, "/"))
}

func (s *S3Service) HeadObject(ctx context.Context, key string) (*MediaObjectHead, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("head object %s: %w", key, err)
	}

	return &MediaObjectHead{
		ContentType: aws.ToString(out.ContentType),
		SizeBytes:   aws.ToInt64(out.ContentLength),
	}, nil
}

func (s *S3Service) CopyObject(ctx context.Context, sourceKey, destinationKey string) error {
	copySource := url.PathEscape(fmt.Sprintf("%s/%s", s.bucketName, sourceKey))
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		Key:        aws.String(destinationKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		return fmt.Errorf("copy object %s -> %s: %w", sourceKey, destinationKey, err)
	}
	return nil
}

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

func (s *S3Service) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	defer out.Body.Close()

	body, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", key, err)
	}
	return body, nil
}

func (s *S3Service) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	keys := make([]string, 0)
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list objects %s: %w", prefix, err)
		}
		for _, object := range page.Contents {
			if object.Key != nil {
				keys = append(keys, *object.Key)
			}
		}
	}

	return keys, nil
}

func BuildStagingKey(userID, sessionID, filename string) string {
	return fmt.Sprintf("staging/%s/%s/%s", userID, sessionID, filename)
}

func BuildPermanentKey(listingID, imageID string) string {
	return fmt.Sprintf("listings/%s/%s", listingID, imageID)
}

func BuildThumbnailKey(listingID, imageID string) string {
	return fmt.Sprintf("thumbnails/%s/%s", listingID, imageID)
}

func BuildRejectedKey(listingID, imageID string) string {
	return fmt.Sprintf("rejected/%s/%s", listingID, imageID)
}
