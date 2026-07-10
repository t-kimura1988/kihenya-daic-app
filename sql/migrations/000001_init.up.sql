-- =============== アカウントテーブル ===============
CREATE SEQUENCE IF NOT EXISTS t_daic_account_id_seq;
COMMENT ON SEQUENCE t_daic_account_id_seq IS 'DaicのアカウントIDシーケンス';

CREATE TABLE IF NOT EXISTS t_daic_account
(
    id            BIGINT       NOT NULL CONSTRAINT t_daic_account_pk PRIMARY KEY DEFAULT nextval('t_daic_account_id_seq'),
    uid           VARCHAR(50)  NOT NULL,
    display_name  VARCHAR(50)  NOT NULL,
    del_flg       VARCHAR(1)   NOT NULL DEFAULT '0',
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN t_daic_account.uid IS 'Supabase auth uid';
COMMENT ON COLUMN t_daic_account.display_name IS '表示名';

CREATE UNIQUE INDEX IF NOT EXISTS t_daic_account_uid_uindex ON t_daic_account (uid);

-- =============== 理想テーブル ===============
CREATE SEQUENCE IF NOT EXISTS t_ideal_id_seq;
COMMENT ON SEQUENCE t_ideal_id_seq IS '理想IDシーケンス';

CREATE TABLE IF NOT EXISTS t_ideal
(
    id            BIGINT       NOT NULL CONSTRAINT t_ideal_pk PRIMARY KEY DEFAULT nextval('t_ideal_id_seq'),
    account_id    BIGINT       NOT NULL,
    title         VARCHAR(100) NOT NULL,
    description   TEXT,
    status        VARCHAR(1)   NOT NULL DEFAULT '1',
    del_flg       VARCHAR(1)   NOT NULL DEFAULT '0',
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN t_ideal.account_id IS 'アカウントID';
COMMENT ON COLUMN t_ideal.title IS '理想のタイトル';
COMMENT ON COLUMN t_ideal.description IS '理想の説明';
COMMENT ON COLUMN t_ideal.status IS '1:進行中 2:完了 3:アーカイブ';

CREATE INDEX IF NOT EXISTS t_ideal_account_id_index ON t_ideal (account_id);

-- =============== 目標テーブル ===============
CREATE SEQUENCE IF NOT EXISTS t_goal_id_seq;
COMMENT ON SEQUENCE t_goal_id_seq IS '目標IDシーケンス';

CREATE TABLE IF NOT EXISTS t_goal
(
    id            BIGINT       NOT NULL CONSTRAINT t_goal_pk PRIMARY KEY DEFAULT nextval('t_goal_id_seq'),
    ideal_id      BIGINT       NOT NULL,
    account_id    BIGINT       NOT NULL,
    title         VARCHAR(100) NOT NULL,
    description   TEXT,
    status        VARCHAR(1)   NOT NULL DEFAULT '1',
    sort_order    INTEGER      NOT NULL DEFAULT 0,
    del_flg       VARCHAR(1)   NOT NULL DEFAULT '0',
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN t_goal.ideal_id IS '理想ID';
COMMENT ON COLUMN t_goal.account_id IS 'アカウントID';
COMMENT ON COLUMN t_goal.title IS '目標のタイトル';
COMMENT ON COLUMN t_goal.description IS '目標の説明';
COMMENT ON COLUMN t_goal.status IS '1:未着手 2:進行中 3:完了';
COMMENT ON COLUMN t_goal.sort_order IS '表示順';

CREATE INDEX IF NOT EXISTS t_goal_ideal_id_index ON t_goal (ideal_id);
CREATE INDEX IF NOT EXISTS t_goal_account_id_index ON t_goal (account_id);

-- =============== タスクテーブル ===============
CREATE SEQUENCE IF NOT EXISTS t_task_id_seq;
COMMENT ON SEQUENCE t_task_id_seq IS 'タスクIDシーケンス';

CREATE TABLE IF NOT EXISTS t_task
(
    id            BIGINT       NOT NULL CONSTRAINT t_task_pk PRIMARY KEY DEFAULT nextval('t_task_id_seq'),
    goal_id       BIGINT       NOT NULL,
    account_id    BIGINT       NOT NULL,
    title         VARCHAR(200) NOT NULL,
    description   TEXT,
    priority      VARCHAR(1)   NOT NULL DEFAULT '2',
    status        VARCHAR(1)   NOT NULL DEFAULT '1',
    del_flg       VARCHAR(1)   NOT NULL DEFAULT '0',
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN t_task.goal_id IS '目標ID';
COMMENT ON COLUMN t_task.account_id IS 'アカウントID';
COMMENT ON COLUMN t_task.title IS 'タスクのタイトル';
COMMENT ON COLUMN t_task.description IS 'タスクの説明';
COMMENT ON COLUMN t_task.priority IS '1:低 2:中 3:高 4:緊急';
COMMENT ON COLUMN t_task.status IS '1:未着手 2:進行中 3:完了';

CREATE INDEX IF NOT EXISTS t_task_goal_id_index ON t_task (goal_id);
CREATE INDEX IF NOT EXISTS t_task_account_id_index ON t_task (account_id);
CREATE INDEX IF NOT EXISTS t_task_priority_index ON t_task (priority);
