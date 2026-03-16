package repository

import (
	"L3.4/internal/model"
	"context"
)

type Repository interface {
	CreateImage(ctx context.Context, img *model.Image) error
	GetImageByID(ctx context.Context, id string) (*model.Image, error)
	DeleteImage(ctx context.Context, id string) error
	ListImages(ctx context.Context) ([]*model.Image, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateProcessedPaths(ctx context.Context, id, resizedPath, thumbPath, watermarkedPath string) error
}
