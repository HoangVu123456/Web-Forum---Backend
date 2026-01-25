package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client handles all S3 operations
type S3Client struct {
	client *s3.Client
	bucket string
	region string
}

// NewS3Client creates a new S3 client instance
func NewS3Client(ctx context.Context, bucket, region string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error loading aws config: %w", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		region: region,
	}, nil
}

// UploadFile uploads a local file to S3 with .env specified the key
func (sc *S3Client) UploadFile(ctx context.Context, filePath, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	_, err = sc.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(sc.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	fmt.Println("File uploaded successfully:", filePath, "to key:", key)
	return nil
}

// CreatePresignedUploadURL generates a presigned PUT URL for uploading files to S3
func (sc *S3Client) CreatePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(sc.client)

	result, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(sc.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("error creating presigned URL: %w", err)
	}

	return result.URL, nil
}

// GetObjectURL returns the public URL for an object in S
func (sc *S3Client) GetObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", sc.bucket, sc.region, key)
}
