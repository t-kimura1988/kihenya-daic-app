-- schema for sqlc (synced with migrations)

CREATE TABLE t_daic_account (
    id BIGINT PRIMARY KEY,
    uid VARCHAR(50) NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    del_flg VARCHAR(1) NOT NULL DEFAULT '0',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE t_ideal (
    id BIGINT PRIMARY KEY,
    account_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(1) NOT NULL DEFAULT '1',
    del_flg VARCHAR(1) NOT NULL DEFAULT '0',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE t_goal (
    id BIGINT PRIMARY KEY,
    ideal_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(1) NOT NULL DEFAULT '1',
    achievement VARCHAR(1) NOT NULL DEFAULT '1',
    sort_order INTEGER NOT NULL DEFAULT 0,
    depends_on_goal_id BIGINT,
    del_flg VARCHAR(1) NOT NULL DEFAULT '0',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE t_task (
    id BIGINT PRIMARY KEY,
    goal_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    priority VARCHAR(1) NOT NULL DEFAULT '2',
    status VARCHAR(1) NOT NULL DEFAULT '1',
    story_point INTEGER NOT NULL DEFAULT 1,
    del_flg VARCHAR(1) NOT NULL DEFAULT '0',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE t_kblog_booklet_link (
    id BIGINT PRIMARY KEY,
    ideal_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    kblog_booklet_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
