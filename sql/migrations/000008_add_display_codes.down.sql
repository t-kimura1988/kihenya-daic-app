DROP INDEX IF EXISTS t_task_uindex_task_cd;
DROP INDEX IF EXISTS t_goal_uindex_goal_cd;
DROP INDEX IF EXISTS t_ideal_uindex_ideal_cd;

ALTER TABLE t_task DROP COLUMN IF EXISTS task_cd;
ALTER TABLE t_goal DROP COLUMN IF EXISTS goal_cd;
ALTER TABLE t_ideal DROP COLUMN IF EXISTS ideal_cd;
