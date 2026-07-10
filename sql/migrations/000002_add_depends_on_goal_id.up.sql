ALTER TABLE t_goal
    ADD COLUMN depends_on_goal_id BIGINT;

COMMENT ON COLUMN t_goal.depends_on_goal_id IS '前提となる目標ID（この目標の前に達成する目標）';

CREATE INDEX IF NOT EXISTS t_goal_depends_on_goal_id_index ON t_goal (depends_on_goal_id);

ALTER TABLE t_goal
    ADD CONSTRAINT t_goal_depends_on_goal_id_fk
        FOREIGN KEY (depends_on_goal_id) REFERENCES t_goal (id);
