-- =============== kblog冊子連携テーブル ===============
-- 理想(t_ideal)ごとにkihenya-blog(kblog)の冊子を連携する
CREATE SEQUENCE IF NOT EXISTS t_kblog_booklet_link_id_seq;
COMMENT ON SEQUENCE t_kblog_booklet_link_id_seq IS 'kblog冊子連携IDシーケンス';

CREATE TABLE IF NOT EXISTS t_kblog_booklet_link
(
    id               BIGINT    NOT NULL CONSTRAINT t_kblog_booklet_link_pk PRIMARY KEY DEFAULT nextval('t_kblog_booklet_link_id_seq'),
    ideal_id         BIGINT    NOT NULL,
    account_id       BIGINT    NOT NULL,
    kblog_booklet_id BIGINT    NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_t_kblog_booklet_link UNIQUE (ideal_id, account_id)
);

COMMENT ON COLUMN t_kblog_booklet_link.ideal_id IS '理想ID';
COMMENT ON COLUMN t_kblog_booklet_link.account_id IS 'アカウントID';
COMMENT ON COLUMN t_kblog_booklet_link.kblog_booklet_id IS 'kblog側の冊子ID';

CREATE INDEX IF NOT EXISTS t_kblog_booklet_link_ideal_id_index ON t_kblog_booklet_link (ideal_id);
CREATE INDEX IF NOT EXISTS t_kblog_booklet_link_account_id_index ON t_kblog_booklet_link (account_id);
