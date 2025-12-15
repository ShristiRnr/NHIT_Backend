package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient handles file storage operations with MinIO
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

	log.Printf("✅ MinIO client initialized (endpoint: %s, bucket: %s)", endpoint, bucket)

	return &MinIOClient{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadSignature uploads a user signature image to MinIO
func (m *MinIOClient) UploadSignature(ctx context.Context, userID, filename string, file io.Reader, size int64) (string, error) {
	// Create object name with user ID prefix
	objectName := fmt.Sprintf("signatures/%s/%s", userID, filename)

	// Upload the file
	_, err := m.client.PutObject(ctx, m.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: "image/jpeg", // You can make this dynamic based on file extension
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload signature: %w", err)
	}

	// Return the object path (can be used to construct full URL later)
	signatureURL := fmt.Sprintf("/%s/%s", m.bucket, objectName)
	log.Printf("✅ Uploaded signature: %s", signatureURL)

	return signatureURL, nil
}

// GetSignatureURL returns the presigned URL for a signature file
func (m *MinIOClient) GetSignatureURL(ctx context.Context, objectPath string) (string, error) {
	// Generate presigned URL valid for 24 hours
	url, err := m.client.PresignedGetObject(ctx, m.bucket, objectPath, 24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteSignature deletes a signature file from MinIO
func (m *MinIOClient) DeleteSignature(ctx context.Context, objectPath string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete signature: %w", err)
	}

	log.Printf("✅ Deleted signature: %s", objectPath)
	return nil
}
