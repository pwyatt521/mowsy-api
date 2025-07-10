package storage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Service struct {
	bucket   string
	region   string
	s3Client *s3.S3
	uploader *s3manager.Uploader
}

func NewS3Service() (*S3Service, error) {
	bucket := os.Getenv("AWS_S3_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")

	if bucket == "" {
		return nil, fmt.Errorf("AWS_S3_BUCKET_NAME environment variable not set")
	}

	if region == "" {
		region = "us-east-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	return &S3Service{
		bucket:   bucket,
		region:   region,
		s3Client: s3Client,
		uploader: uploader,
	}, nil
}

type UploadResult struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

func (s *S3Service) UploadFile(file io.Reader, fileName, mimeType string, userID uint) (*UploadResult, error) {
	if err := s.validateFile(fileName, mimeType); err != nil {
		return nil, err
	}

	key := s.generateKey(fileName, userID)

	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(mimeType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		URL:      result.Location,
		Key:      key,
		MimeType: mimeType,
	}, nil
}

func (s *S3Service) validateFile(fileName, mimeType string) error {
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"application/pdf": true,
	}

	if !allowedMimeTypes[mimeType] {
		return fmt.Errorf("unsupported file type: %s", mimeType)
	}

	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".pdf":  true,
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	return nil
}

func (s *S3Service) generateKey(fileName string, userID uint) string {
	timestamp := time.Now().Unix()
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(filepath.Base(fileName), ext)
	baseName = url.QueryEscape(baseName)
	
	return fmt.Sprintf("uploads/%d/%d_%s%s", userID, timestamp, baseName, ext)
}

func (s *S3Service) DeleteFile(key string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *S3Service) GetPresignedURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

func (s *S3Service) GetPresignedUploadURL(key string, mimeType string, expiration time.Duration) (string, error) {
	req, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(mimeType),
		ACL:         aws.String("public-read"),
	})

	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return url, nil
}