package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"mowsy-api/pkg/storage"
)

type UploadService struct {
	s3Service *storage.S3Service
}

func NewUploadService() (*UploadService, error) {
	s3Service, err := storage.NewS3Service()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 service: %w", err)
	}

	return &UploadService{
		s3Service: s3Service,
	}, nil
}

type UploadImageRequest struct {
	File     *multipart.FileHeader `form:"file" binding:"required"`
	Category string                `form:"category"`
}

type UploadImageResponse struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

func (s *UploadService) UploadImage(userID uint, file *multipart.FileHeader, category string) (*UploadImageResponse, error) {
	if file.Size > 10*1024*1024 { // 10MB limit
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	result, err := s.s3Service.UploadFile(src, file.Filename, mimeType, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadImageResponse{
		URL:      result.URL,
		Key:      result.Key,
		Size:     file.Size,
		MimeType: mimeType,
	}, nil
}

func (s *UploadService) UploadFromReader(userID uint, reader io.Reader, fileName, mimeType string) (*UploadImageResponse, error) {
	result, err := s.s3Service.UploadFile(reader, fileName, mimeType, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadImageResponse{
		URL:      result.URL,
		Key:      result.Key,
		MimeType: mimeType,
	}, nil
}

func (s *UploadService) DeleteFile(key string) error {
	return s.s3Service.DeleteFile(key)
}

func (s *UploadService) GetPresignedURL(key string) (string, error) {
	return s.s3Service.GetPresignedURL(key, 1*time.Hour)
}

func (s *UploadService) GetPresignedUploadURL(userID uint, fileName, mimeType string) (string, error) {
	key := fmt.Sprintf("uploads/%d/%d_%s", userID, time.Now().Unix(), fileName)
	return s.s3Service.GetPresignedUploadURL(key, mimeType, 1*time.Hour)
}