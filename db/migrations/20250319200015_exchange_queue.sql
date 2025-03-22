-- +goose Up
-- +goose StatementBegin
CALL mq.create_exchange (exchange_name => 'users');

CALL mq.create_queue (exchange_name => 'users', queue_name => 'create_user', routing_key_pattern => '^user\.create$');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM mq.queue
WHERE queue_name = 'create_user';

DELETE FROM mq.exchange
WHERE exchange_name = 'users';

-- +goose StatementEnd
