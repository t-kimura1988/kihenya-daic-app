ALTER TABLE t_goal DROP CONSTRAINT IF EXISTS t_goal_depends_on_goal_id_fk;
DROP INDEX IF EXISTS t_goal_depends_on_goal_id_index;
ALTER TABLE t_goal DROP COLUMN IF EXISTS depends_on_goal_id;
