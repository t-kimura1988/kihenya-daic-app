ALTER TABLE t_goal
    ADD COLUMN achievement VARCHAR(1) NOT NULL DEFAULT '1';

COMMENT ON COLUMN t_goal.achievement IS '目標の達成コード 1:未達成 2:達成';
