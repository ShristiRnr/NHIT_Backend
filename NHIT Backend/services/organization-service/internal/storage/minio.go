package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient handles file storage operations with MinIO for organizations
type MinIOClient struct {
	client *minio.Client
	bucket string
}

// NewMinIOClient creates a new MinIO client instance
func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Check if bucket exists, create if it doesn't
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("✅ Created MinIO bucket: %s", bucket)
	}

	log.Printf("✅ MinIO client initialized for Organization Service (endpoint: %s, bucket: %s)", endpoint, bucket)

	return &MinIOClient{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadLogo uploads an organization logo image to MinIO
func (m *MinIOClient) UploadLogo(ctx context.Context, orgID, filename string, file io.Reader, size int64) (string, error) {
	// Create object name with organization ID prefix
	objectName := fmt.Sprintf("logos/%s/%s", orgID, filename)

	// Upload the file
	_, err := m.client.PutObject(ctx, m.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: "image/png", // Default to PNG, can be dynamic
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload logo: %w", err)
	}

	// Return the object path
	logoURL := fmt.Sprintf("/%s/%s", m.bucket, objectName)
	log.Printf("✅ Uploaded logo: %s", logoURL)

	return logoURL, nil
}

// GetLogoURL returns the presigned URL for a logo file
func (m *MinIOClient) GetLogoURL(ctx context.Context, objectPath string) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucket, objectPath, 24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteLogo deletes a logo file from MinIO
func (m *MinIOClient) DeleteLogo(ctx context.Context, objectPath string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete logo: %w", err)
	}

	log.Printf("✅ Deleted logo: %s", objectPath)
	return nil
}
