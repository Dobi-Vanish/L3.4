package model

import "time"

const (
	StatusUploaded   = "uploaded"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

type Image struct {
	ID               string    `json:"id" db:"id"`
	OriginalFilename string    `json:"filename" db:"original_filename"`
	OriginalPath     string    `json:"-" db:"original_path"`
	ResizedPath      string    `json:"-" db:"resized_path"`
	ThumbnailPath    string    `json:"-" db:"thumbnail_path"`
	WatermarkedPath  string    `json:"-" db:"watermarked_path"`
	Status           string    `json:"status" db:"status"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"-" db:"updated_at"`
}
