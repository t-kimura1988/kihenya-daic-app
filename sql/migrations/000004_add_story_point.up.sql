ALTER TABLE t_task
    ADD COLUMN story_point INTEGER NOT NULL DEFAULT 1;

COMMENT ON COLUMN t_task.story_point IS 'ストーリーポイント（必要な木材の本数）';
