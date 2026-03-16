-- +goose Up
CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY,
    original_path TEXT NOT NULL,
    resized_path TEXT,
    thumbnail_path TEXT,
    watermarked_path TEXT,
    status VARCHAR(20) NOT NULL,
    original_filename TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
    DROP TABLE IF EXISTS images;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
