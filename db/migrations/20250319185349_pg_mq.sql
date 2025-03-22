-- +goose Up
-- +goose StatementBegin
-- INITIALIZE
CREATE SCHEMA mq;

-- TABLES
CREATE TABLE mq.exchange (
    exchange_id serial PRIMARY KEY,
    exchange_name text NOT NULL UNIQUE
);

CREATE TABLE mq.message_intake (
    exchange_id int NOT NULL REFERENCES mq.exchange (exchange_id) ON DELETE CASCADE,
    routing_key text NOT NULL,
    body json NOT NULL,
    headers hstore NOT NULL DEFAULT '',
    publish_time timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE mq.queue (
    queue_id bigserial PRIMARY KEY,
    exchange_id int NOT NULL REFERENCES mq.exchange (exchange_id) ON DELETE CASCADE,
    queue_name text NOT NULL UNIQUE,
    routing_key_pattern text NOT NULL DEFAULT '^.*$'
);

CREATE TABLE mq.message (
    message_id bigserial PRIMARY KEY,
    LIKE mq.message_intake,
    queue_id bigint NOT NULL REFERENCES mq.queue (queue_id) ON DELETE CASCADE
);

CREATE INDEX ON mq.message (queue_id);

CREATE TABLE mq.message_waiting (
    message_id bigint PRIMARY KEY REFERENCES mq.message (message_id) ON DELETE CASCADE,
    queue_id bigint NOT NULL REFERENCES mq.queue (queue_id) ON DELETE CASCADE,
    since_time timestamptz NOT NULL DEFAULT now(),
    not_until_time timestamptz NULL
);

CREATE INDEX ON mq.message_waiting (queue_id);

CREATE TABLE mq.channel (
    channel_id bigserial PRIMARY KEY,
    channel_name text NOT NULL UNIQUE,
    queue_id bigint NOT NULL REFERENCES mq.queue (queue_id) ON DELETE CASCADE,
    maximum_messages int NOT NULL DEFAULT 1
);

CREATE TABLE mq.channel_waiting (
    channel_id bigint NOT NULL REFERENCES mq.channel (channel_id) ON DELETE CASCADE,
    slot int NOT NULL,
    queue_id bigint NOT NULL REFERENCES mq.queue (queue_id) ON DELETE CASCADE,
    since_time timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (channel_id, slot)
);

CREATE INDEX ON mq.channel_waiting (queue_id);

CREATE TABLE mq.delivery (
    delivery_id bigserial PRIMARY KEY,
    message_id bigint NOT NULL REFERENCES mq.message (message_id) ON DELETE CASCADE,
    channel_id bigint NOT NULL REFERENCES mq.channel (channel_id) ON DELETE CASCADE,
    slot int NOT NULL,
    queue_id bigint NOT NULL REFERENCES mq.queue (queue_id) ON DELETE CASCADE,
    delivery_time timestamptz NOT NULL DEFAULT now()
);

-- FUNCTIONS
CREATE FUNCTION mq.notify_channel (delivery_id bigint, message_id bigint, channel_name text)
    RETURNS VOID
    AS $$
DECLARE
    payload text;
BEGIN
    SELECT
        row_to_json(md) INTO payload
    FROM (
        SELECT
            delivery_id,
            m.routing_key,
            m.body,
            m.headers
        FROM
            mq.message m
        WHERE
            m.message_id = notify_channel.message_id) md;
    PERFORM
        pg_notify(channel_name, payload);
    RAISE NOTICE 'Sent message % to channel %', notify_channel.message_id, channel_name;
END;
$$
LANGUAGE plpgsql;

CREATE FUNCTION mq.take_waiting_message (queue_id bigint)
    RETURNS bigint
    AS $$
    DELETE FROM mq.message_waiting mw
    WHERE mw.message_id = (
            SELECT
                m.message_id
            FROM
                mq.message_waiting m
            WHERE
                m.queue_id = queue_id
                AND (not_until_time IS NULL
                    OR not_until_time <= now())
            ORDER BY
                m.message_id
            FOR UPDATE
                SKIP LOCKED
            LIMIT 1)
RETURNING
    mw.message_id;
$$
LANGUAGE SQL;

CREATE FUNCTION mq.take_waiting_channel (queue_id bigint)
    RETURNS SETOF mq.channel_waiting
    AS $$
    WITH row_to_delete AS (
        SELECT
            *
        FROM
            mq.channel_waiting c
        WHERE
            c.queue_id = queue_id
        ORDER BY
            c.since_time
        FOR UPDATE
            SKIP LOCKED
        LIMIT 1)
DELETE FROM mq.channel_waiting cw
WHERE cw.channel_id = (
        SELECT
            channel_id
        FROM
            row_to_delete)
    AND cw.slot = (
        SELECT
            slot
        FROM
            row_to_delete)
RETURNING
    *;
$$
LANGUAGE SQL;

-- TRIGGERS
-- INSERT MESSAGE
CREATE FUNCTION mq.insert_message ()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO mq.message (exchange_id, routing_key, body, headers, publish_time, queue_id)
    SELECT
        NEW.exchange_id,
        NEW.routing_key,
        NEW.body,
        NEW.headers,
        NEW.publish_time,
        q.queue_id
    FROM
        mq.queue q
    WHERE
        NEW.exchange_id = q.exchange_id
        AND NEW.routing_key ~ q.routing_key_pattern
    ON CONFLICT
        DO NOTHING;
    RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER insert_message_before_insert
    BEFORE INSERT ON mq.message_intake
    FOR EACH ROW
    EXECUTE PROCEDURE mq.insert_message ();

-- INSERT MESSAGE WAITING
CREATE FUNCTION mq.insert_message_waiting ()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO mq.message_waiting (message_id, queue_id)
        VALUES (NEW.message_id, NEW.queue_id)
    ON CONFLICT
        DO NOTHING;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER insert_message_waiting_after_insert
    AFTER INSERT ON mq.message
    FOR EACH ROW
    EXECUTE PROCEDURE mq.insert_message_waiting ();

-- DELIVER MESSAGE
CREATE FUNCTION mq.deliver_message ()
    RETURNS TRIGGER
    AS $$
BEGIN
    EXECUTE mq.notify_channel (NEW.delivery_id, NEW.message_id, text(NEW.channel_id));
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER deliver_message_after_insert
    AFTER INSERT ON mq.delivery
    FOR EACH ROW
    EXECUTE PROCEDURE mq.deliver_message ();

-- CHANNEL INITIALIZE
CREATE FUNCTION mq.channel_initialize ()
    RETURNS TRIGGER
    AS $$
BEGIN
    FOR i IN 1..NEW.maximum_messages LOOP
        RAISE NOTICE 'Creating channel waiting slot: %', i;
        INSERT INTO mq.channel_waiting (channel_id, slot, queue_id)
            VALUES (NEW.channel_id, i, NEW.queue_id);
    END LOOP;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER channel_initialize_after_insert
    AFTER INSERT ON mq.channel
    FOR EACH ROW
    EXECUTE PROCEDURE mq.channel_initialize ();

-- MATCH MESSAGE
CREATE FUNCTION mq.match_message ()
    RETURNS TRIGGER
    AS $$
DECLARE
    selected_message_id bigint;
BEGIN
    SELECT
        mq.take_waiting_message (NEW.queue_id) INTO selected_message_id;
    IF selected_message_id IS NULL THEN
        RETURN NEW;
    END IF;
    INSERT INTO mq.delivery (message_id, channel_id, slot, queue_id)
        VALUES (selected_message_id, NEW.channel_id, NEW.slot, NEW.queue_id);
    RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER match_message_before_insert
    BEFORE INSERT ON mq.channel_waiting
    FOR EACH ROW
    EXECUTE PROCEDURE mq.match_message ();

CREATE FUNCTION mq.match_channel ()
    RETURNS TRIGGER
    AS $$
DECLARE
    selected_channel record;
BEGIN
    IF NEW.not_until_time IS NOT NULL AND NEW.not_until_time > now() THEN
        RETURN NEW;
    END IF;
    SELECT
        *
    FROM
        mq.take_waiting_channel (NEW.queue_id) INTO selected_channel;
    IF selected_channel IS NULL THEN
        RETURN NEW;
    END IF;
    INSERT INTO mq.delivery (message_id, channel_id, slot, queue_id)
        VALUES (NEW.message_id, selected_channel.channel_id, selected_channel.slot, NEW.queue_id);
    RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER match_channel_before_insert
    BEFORE INSERT ON mq.message_waiting
    FOR EACH ROW
    EXECUTE PROCEDURE mq.match_channel ();

-- CONSUMER
CREATE PROCEDURE mq.create_exchange (exchange_name text)
LANGUAGE sql
AS $$
    INSERT INTO mq.exchange (exchange_name)
        VALUES (exchange_name);
$$;

CREATE PROCEDURE mq.delete_exchange (exchange_name text)
LANGUAGE sql
AS $$
    DELETE FROM mq.exchange
    WHERE exchange_name = delete_exchange.exchange_name;
$$;

CREATE PROCEDURE mq.create_queue (exchange_name text, queue_name text, routing_key_pattern text DEFAULT '^.*$')
LANGUAGE sql
AS $$
    INSERT INTO mq.queue (exchange_id, queue_name, routing_key_pattern)
        VALUES ((
                SELECT
                    exchange_id
                FROM
                    mq.exchange
                WHERE
                    exchange_name = create_queue.exchange_name), queue_name, routing_key_pattern);
$$;

CREATE PROCEDURE mq.delete_queue (queue_name text)
LANGUAGE sql
AS $$
    DELETE FROM mq.queue
    WHERE queue_name = delete_queue.queue_name;
$$;


/* REGISTER */
CREATE OR REPLACE FUNCTION mq.open_channel (queue_name text, maximum_messages int DEFAULT 1)
    RETURNS bigint
    LANGUAGE plpgsql
    AS $$
DECLARE
    queue_id bigint;
    new_channel_id bigint;
    new_channel_name text;
BEGIN
    SELECT
        q.queue_id INTO queue_id
    FROM
        mq.queue q
    WHERE
        q.queue_name = open_channel.queue_name;
    IF queue_id IS NULL THEN
        RAISE WARNING 'No such queue';
        RETURN NULL;
    END IF;
    new_channel_name := text(pg_backend_pid());
    SELECT
        channel_id
    FROM
        mq.channel c
    WHERE
        c.channel_name = new_channel_name INTO new_channel_id;
    IF new_channel_id IS NULL THEN
        INSERT INTO mq.channel (channel_name, queue_id, maximum_messages)
            VALUES (new_channel_name, queue_id, open_channel.maximum_messages)
        RETURNING
            channel_id INTO new_channel_id;
    END IF;
    EXECUTE format('LISTEN "%s"', new_channel_id);
    RETURN new_channel_id;
END;
$$;

CREATE PROCEDURE mq.close_channel ()
LANGUAGE plpgsql
AS $$
DECLARE
    current_channel_id bigint;
    current_channel_name text;
BEGIN
    current_channel_name := text(pg_backend_pid());
    SELECT
        channel_id
    FROM
        mq.channel c
    WHERE
        c.channel_name = current_channel_name INTO current_channel_id;
    IF current_channel_id IS NULL THEN
        RETURN;
    END IF;
    CALL mq.close_channel (current_channel_id);
END;
$$;

CREATE PROCEDURE mq.close_channel (close_channel_id bigint)
LANGUAGE plpgsql
AS $$
BEGIN
    EXECUTE format('UNLISTEN "%s"', close_channel_id);
    INSERT INTO mq.message_waiting (
        SELECT
            message_id,
            queue_id
        FROM
            mq.delivery
        WHERE
            channel_id = close_channel_id)
ON CONFLICT
    DO NOTHING;
    DELETE FROM mq.channel c
    WHERE c.channel_id = close_channel_id;
END;
$$;

CREATE PROCEDURE mq.close_dead_channels ()
LANGUAGE plpgsql
AS $$
DECLARE
    dead_channel record;
BEGIN
    FOR dead_channel IN (
        SELECT
            *
        FROM
            mq.channel c
        WHERE
            c.channel_name::int NOT IN (
                SELECT
                    pid
                FROM
                    pg_catalog.pg_stat_activity psa))
            LOOP
                CALL mq.close_channel (dead_channel.channel_id);
            END LOOP;
END;
$$;

CREATE PROCEDURE mq.sweep_waiting_message (queue_name text)
LANGUAGE plpgsql
AS $$
DECLARE
    queue_id bigint;
    current_channel record;
BEGIN
    SELECT
        q.queue_id INTO queue_id
    FROM
        mq.queue q
    WHERE
        q.queue_name = sweep_waiting_message.queue_name;
    IF queue_id IS NULL THEN
        RAISE WARNING 'No such queue';
        RETURN;
    END IF;
    CALL mq.sweep_waiting_message (queue_id);
END;
$$;

CREATE PROCEDURE mq.sweep_waiting_message (queue_id bigint)
LANGUAGE plpgsql
AS $$
DECLARE
    selected_message_id bigint;
    selected_channel record;
BEGIN
    SELECT
        mq.take_waiting_message (queue_id) INTO selected_message_id;
    IF selected_message_id IS NULL THEN
        RETURN;
    END IF;
    SELECT
        *
    FROM
        mq.take_waiting_channel (queue_id) INTO selected_channel;
    IF selected_channel IS NULL THEN
        RETURN;
    END IF;
    INSERT INTO mq.delivery (message_id, channel_id, slot, queue_id)
        VALUES (selected_message_id, selected_channel.channel_id, selected_channel.slot, queue_id);
END;
$$;


/* ACK */
CREATE PROCEDURE mq.ack (delivery_id bigint)
LANGUAGE plpgsql
AS $$
DECLARE
    delivery RECORD;
BEGIN
    SELECT
        * INTO delivery
    FROM
        mq.delivery d
    WHERE
        d.delivery_id = ack.delivery_id;
    IF delivery IS NULL THEN
        RAISE WARNING 'No such delivery';
        RETURN;
    END IF;
    DELETE FROM mq.message m
    WHERE m.message_id = delivery.message_id;
    INSERT INTO mq.channel_waiting (channel_id, slot, queue_id)
        VALUES (delivery.channel_id, delivery.slot, delivery.queue_id)
    ON CONFLICT
        DO NOTHING;
END;
$$;


/* NACK */
CREATE PROCEDURE mq.nack (delivery_id bigint, retry_after interval DEFAULT '0s' ::interval)
LANGUAGE plpgsql
AS $$
DECLARE
    delivery RECORD;
BEGIN
    SELECT
        * INTO delivery
    FROM
        mq.delivery d
    WHERE
        d.delivery_id = nack.delivery_id;
    IF delivery IS NULL THEN
        RAISE WARNING 'No such delivery';
        RETURN;
    END IF;
    DELETE FROM mq.delivery d
    WHERE d.delivery_id = nack.delivery_id;
    INSERT INTO mq.message_waiting (message_id, queue_id, not_until_time)
        VALUES (delivery.message_id, delivery.queue_id, now() + nack.retry_after)
    ON CONFLICT
        DO NOTHING;
    INSERT INTO mq.channel_waiting (channel_id, slot, queue_id)
        VALUES (delivery.channel_id, delivery.slot, delivery.queue_id)
    ON CONFLICT
        DO NOTHING;
END;
$$;


/* PUBLISH */
CREATE PROCEDURE mq.publish (exchange_name text, routing_key text, body json, headers hstore)
LANGUAGE plpgsql
AS $$
DECLARE
    exchange_id bigint;
BEGIN
    SELECT
        e.exchange_id INTO exchange_id
    FROM
        mq.exchange e
    WHERE
        e.exchange_name = publish.exchange_name;
    IF exchange_id IS NULL THEN
        RAISE WARNING 'No such exchange.';
        RETURN;
    END IF;
    INSERT INTO mq.message_intake (exchange_id, routing_key, body, headers)
        VALUES (exchange_id, routing_key, body, headers);
END;
$$;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP SCHEMA mq CASCADE;

-- +goose StatementEnd
