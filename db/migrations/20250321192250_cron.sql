-- +goose Up
-- +goose StatementBegin
SELECT
    cron.schedule ('close_dead_channels_job', '*/1 * * * *', 'CALL mq.close_dead_channels()');

SELECT
    cron.schedule ('sweep_users_job', '*/1 * * * *', 'CALL mq.sweep_waiting_message(queue_name=>''create_user'')');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    cron.unschedule (jobname)
FROM
    cron.job
WHERE
    command LIKE '%close_dead_channels%';

SELECT
    cron.unschedule (jobname)
FROM
    cron.job
WHERE
    command LIKE '%sweep_waiting_message%';

-- +goose StatementEnd
