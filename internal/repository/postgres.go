package repository

import (
	"L3.4/internal/model"
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/logger"
)

type postgresRepo struct {
	db  *pgxdriver.Postgres
	log logger.Logger
}

func NewPostgresRepository(db *pgxdriver.Postgres, log logger.Logger) Repository {
	return &postgresRepo{db: db, log: log}
}

func (r *postgresRepo) CreateImage(ctx context.Context, img *model.Image) error {
	query := r.db.Insert("images").
		Columns("id", "original_filename", "original_path", "resized_path", "thumbnail_path",
			"watermarked_path", "status", "created_at", "updated_at").
		Values(img.ID, img.OriginalFilename, img.OriginalPath, img.ResizedPath,
			img.ThumbnailPath, img.WatermarkedPath, img.Status, img.CreatedAt, img.UpdatedAt)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *postgresRepo) GetImageByID(ctx context.Context, id string) (*model.Image, error) {
	query := r.db.Select("*").
		From("images").
		Where(squirrel.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	var img model.Image
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&img.ID,
		&img.OriginalPath,
		&img.ResizedPath,
		&img.ThumbnailPath,
		&img.WatermarkedPath,
		&img.Status,
		&img.OriginalFilename,
		&img.CreatedAt,
		&img.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *postgresRepo) DeleteImage(ctx context.Context, id string) error {
	query := r.db.Delete("images").
		Where(squirrel.Eq{"id": id})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *postgresRepo) ListImages(ctx context.Context) ([]*model.Image, error) {
	query := r.db.Select("*").
		From("images").
		OrderBy("created_at DESC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*model.Image
	for rows.Next() {
		var img model.Image
		err = rows.Scan(
			&img.ID,
			&img.OriginalPath,
			&img.ResizedPath,
			&img.ThumbnailPath,
			&img.WatermarkedPath,
			&img.Status,
			&img.OriginalFilename,
			&img.CreatedAt,
			&img.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, &img)
	}
	return images, rows.Err()
}

func (r *postgresRepo) UpdateStatus(ctx context.Context, id, status string) error {
	query := r.db.Update("images").
		Set("status", status).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build update: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *postgresRepo) UpdateProcessedPaths(ctx context.Context, id, resizedPath, thumbPath, watermarkedPath string) error {
	query := r.db.Update("images").
		Set("resized_path", resizedPath).
		Set("thumbnail_path", thumbPath).
		Set("watermarked_path", watermarkedPath).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build update: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}
