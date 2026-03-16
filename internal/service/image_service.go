package service

import (
	"L3.4/internal/processor"
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"L3.4/internal/model"
	"L3.4/internal/repository"
	"L3.4/internal/storage"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/logger"
)

type ImageService struct {
	repo        repository.Repository
	storage     storage.Storage
	producer    *kafka.Writer
	log         logger.Logger
	storagePath string
}

func NewImageService(
	repo repository.Repository,
	st storage.Storage,
	producer *kafka.Writer,
	log logger.Logger,
	storagePath string,
) *ImageService {
	return &ImageService{
		repo:        repo,
		storage:     st,
		producer:    producer,
		log:         log,
		storagePath: storagePath,
	}
}

func (s *ImageService) UploadImage(ctx context.Context, filename string, data []byte) (*model.Image, error) {
	id := uuid.New().String()

	relativePath := filepath.Join(id, filename)
	originalPath, err := s.storage.SaveFile(relativePath, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("save original: %w", err)
	}
	if originalPath == "" {
		return nil, errors.New("saved file path is empty")
	}
	s.log.Info("file saved", "path", originalPath)

	img := &model.Image{
		ID:               id,
		OriginalFilename: filename,
		OriginalPath:     originalPath,
		Status:           model.StatusUploaded,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.CreateImage(ctx, img); err != nil {
		_ = s.storage.DeleteFile(originalPath)
		return nil, fmt.Errorf("create image in db: %w", err)
	}

	err = s.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(id),
		Value: []byte(id),
	})
	if err != nil {
		s.log.Error("failed to send kafka message", "image_id", id, "error", err)
	}

	return img, nil
}

func (s *ImageService) GetImage(ctx context.Context, id string) (*model.Image, error) {
	return s.repo.GetImageByID(ctx, id)
}

func (s *ImageService) DeleteImage(ctx context.Context, id string) error {
	img, err := s.repo.GetImageByID(ctx, id)
	if err != nil {
		return err
	}
	if img == nil {
		return nil
	}

	paths := []string{img.OriginalPath, img.ResizedPath, img.ThumbnailPath, img.WatermarkedPath}
	for _, p := range paths {
		if p != "" {
			_ = s.storage.DeleteFile(p)
		}
	}

	return s.repo.DeleteImage(ctx, id)
}

func (s *ImageService) ListImages(ctx context.Context) ([]*model.Image, error) {
	images, err := s.repo.ListImages(ctx)
	if err != nil {
		return nil, err
	}
	if images == nil {
		return []*model.Image{}, nil
	}
	return images, nil
}

func (s *ImageService) GetImageFile(img *model.Image, variant string) (string, error) {
	switch variant {
	case "original":
		return img.OriginalPath, nil
	case "resized":
		return img.ResizedPath, nil
	case "thumbnail":
		return img.ThumbnailPath, nil
	case "watermarked":
		return img.WatermarkedPath, nil
	default:
		return "", errors.New("unknown variant")
	}
}

func (s *ImageService) ProcessImage(ctx context.Context, id string) error {
	img, err := s.repo.GetImageByID(ctx, id)
	if err != nil {
		return err
	}
	if img == nil {
		return errors.New("image not found")
	}

	if err := s.repo.UpdateStatus(ctx, id, model.StatusProcessing); err != nil {
		return err
	}

	baseName := filepath.Base(img.OriginalPath)
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]

	proc := processor.NewProcessor()
	resizedPath, thumbPath, watermarkedPath, err := proc.Process(
		img.OriginalPath,
		s.storagePath,
		baseName,
	)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, id, model.StatusFailed)
		return err
	}

	if err := s.repo.UpdateProcessedPaths(ctx, id, resizedPath, thumbPath, watermarkedPath); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, model.StatusDone)
}
