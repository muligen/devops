// Package storage provides MinIO object storage functionality.
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds MinIO configuration.
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// Client wraps MinIO client with common operations.
type Client struct {
	client *minio.Client
	bucket string
}

// New creates a new MinIO client.
func New(cfg *Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Check if bucket exists, create if not
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &Client{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload uploads an object to the bucket.
func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// UploadFile uploads a file to the bucket.
func (c *Client) UploadFile(ctx context.Context, objectName string, filePath string, contentType string) (int64, error) {
	info, err := c.client.FPutObject(ctx, c.bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to upload file: %w", err)
	}
	return info.Size, nil
}

// Download downloads an object from the bucket.
func (c *Client) Download(ctx context.Context, objectName string) (*minio.Object, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return obj, nil
}

// DownloadFile downloads an object to a file.
func (c *Client) DownloadFile(ctx context.Context, objectName string, filePath string) error {
	if err := c.client.FGetObject(ctx, c.bucket, objectName, filePath, minio.GetObjectOptions{}); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

// Delete deletes an object from the bucket.
func (c *Client) Delete(ctx context.Context, objectName string) error {
	if err := c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// Stat returns object information.
func (c *Client) Stat(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	info, err := c.client.StatObject(ctx, c.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to stat object: %w", err)
	}
	return info, nil
}

// PresignedURL generates a presigned URL for downloading an object.
func (c *Client) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedGetObject(ctx, c.bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// PresignedUploadURL generates a presigned URL for uploading an object.
func (c *Client) PresignedUploadURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedPutObject(ctx, c.bucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}
	return url.String(), nil
}

// ListObjects lists objects in the bucket.
func (c *Client) ListObjects(ctx context.Context, prefix string) <-chan minio.ObjectInfo {
	return c.client.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix: prefix,
	})
}

// BucketExists checks if the bucket exists.
func (c *Client) BucketExists(ctx context.Context) (bool, error) {
	return c.client.BucketExists(ctx, c.bucket)
}

// Close is a no-op for MinIO client (satisfies closer interface).
func (c *Client) Close() error {
	// MinIO client doesn't need explicit closing
	return nil
}
