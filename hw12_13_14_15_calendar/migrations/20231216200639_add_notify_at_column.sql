-- +goose Up
-- +goose StatementBegin
ALTER TABLE event
ADD COLUMN notify_at TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event
DROP COLUMN notify_at;
-- +goose StatementEnd
