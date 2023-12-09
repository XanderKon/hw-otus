-- +goose Up
-- +goose StatementBegin
CREATE TABLE event (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title             VARCHAR(255) NOT NULL,
    date_time         TIMESTAMP WITH TIME ZONE NOT NULL,
    duration          INTEGER NOT NULL,
    description       TEXT,
    user_id           INTEGER NOT NULL,
    notification_time TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event;
-- +goose StatementEnd
