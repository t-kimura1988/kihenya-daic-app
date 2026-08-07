-- =============== 表示用コード列を追加 ===============
ALTER TABLE t_ideal
    ADD COLUMN IF NOT EXISTS ideal_cd VARCHAR(20) NOT NULL DEFAULT '';

COMMENT ON COLUMN t_ideal.ideal_cd IS '理想コード (il_[17桁英数字])';

ALTER TABLE t_goal
    ADD COLUMN IF NOT EXISTS goal_cd VARCHAR(20) NOT NULL DEFAULT '';

COMMENT ON COLUMN t_goal.goal_cd IS '目標コード (gl_[17桁英数字])';

ALTER TABLE t_task
    ADD COLUMN IF NOT EXISTS task_cd VARCHAR(20) NOT NULL DEFAULT '';

COMMENT ON COLUMN t_task.task_cd IS 'タスクコード (tk_[17桁英数字])';

-- 既存レコードに一意のコードを割り当て
UPDATE t_ideal
SET ideal_cd = 'il_' || substr(md5(random()::text || id::text), 1, 17)
WHERE ideal_cd = '';

UPDATE t_goal
SET goal_cd = 'gl_' || substr(md5(random()::text || id::text), 1, 17)
WHERE goal_cd = '';

UPDATE t_task
SET task_cd = 'tk_' || substr(md5(random()::text || id::text), 1, 17)
WHERE task_cd = '';

CREATE UNIQUE INDEX IF NOT EXISTS t_ideal_uindex_ideal_cd
    ON t_ideal (ideal_cd);

CREATE UNIQUE INDEX IF NOT EXISTS t_goal_uindex_goal_cd
    ON t_goal (goal_cd);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_uindex_task_cd
    ON t_task (task_cd);
