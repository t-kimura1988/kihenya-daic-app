-- name: GetAccountByUID :one
SELECT id, uid, display_name, del_flg, created_at, updated_at
FROM t_daic_account
WHERE uid = $1 AND del_flg = '0';

-- name: CreateAccount :one
INSERT INTO t_daic_account (uid, display_name)
VALUES ($1, $2)
RETURNING id, uid, display_name, del_flg, created_at, updated_at;

-- name: GetAccountByID :one
SELECT id, uid, display_name, del_flg, created_at, updated_at
FROM t_daic_account
WHERE id = $1 AND del_flg = '0';

-- name: CreateIdeal :one
INSERT INTO t_ideal (account_id, title, description, ideal_cd)
VALUES ($1, $2, $3, $4)
RETURNING id, account_id, title, description, status, ideal_cd, del_flg, created_at, updated_at;

-- name: ListIdealsByAccount :many
SELECT id, account_id, title, description, status, ideal_cd, del_flg, created_at, updated_at
FROM t_ideal
WHERE account_id = $1 AND del_flg = '0'
ORDER BY created_at DESC;

-- name: GetIdealByID :one
SELECT id, account_id, title, description, status, ideal_cd, del_flg, created_at, updated_at
FROM t_ideal
WHERE id = $1 AND account_id = $2 AND del_flg = '0';

-- name: UpdateIdeal :one
UPDATE t_ideal
SET title = $3, description = $4, status = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, account_id, title, description, status, ideal_cd, del_flg, created_at, updated_at;

-- name: DeleteIdeal :exec
UPDATE t_ideal
SET del_flg = '1', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0';

-- name: CreateGoal :one
INSERT INTO t_goal (ideal_id, account_id, title, description, sort_order, depends_on_goal_id, goal_cd)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, ideal_id, account_id, title, description, status, achievement, sort_order, depends_on_goal_id, goal_cd, del_flg, created_at, updated_at;

-- name: ListGoalsByIdeal :many
SELECT id, ideal_id, account_id, title, description, status, achievement, sort_order, depends_on_goal_id, goal_cd, del_flg, created_at, updated_at
FROM t_goal
WHERE ideal_id = $1 AND account_id = $2 AND del_flg = '0'
ORDER BY sort_order ASC, created_at ASC;

-- name: GetGoalByID :one
SELECT id, ideal_id, account_id, title, description, status, achievement, sort_order, depends_on_goal_id, goal_cd, del_flg, created_at, updated_at
FROM t_goal
WHERE id = $1 AND account_id = $2 AND del_flg = '0';

-- name: UpdateGoal :one
UPDATE t_goal
SET title = $3, description = $4, status = $5, sort_order = $6, depends_on_goal_id = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, ideal_id, account_id, title, description, status, achievement, sort_order, depends_on_goal_id, goal_cd, del_flg, created_at, updated_at;

-- name: DeleteGoal :exec
UPDATE t_goal
SET del_flg = '1', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0';

-- name: CancelGoal :one
UPDATE t_goal
SET status = '4', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, ideal_id, account_id, title, description, status, achievement, sort_order, depends_on_goal_id, goal_cd, del_flg, created_at, updated_at;

-- name: AchieveGoal :one
UPDATE t_goal
SET achievement = '2', status = '3', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, ideal_id, account_id, title, description, status, achievement, sort_order, depends_on_goal_id, goal_cd, del_flg, created_at, updated_at;

-- name: GetGoalPoints :one
SELECT
    COALESCE(SUM(CASE WHEN t.del_flg = '0' AND t.status <> '4' THEN t.story_point ELSE 0 END), 0)::int AS total_story_point,
    COALESCE(SUM(CASE WHEN t.del_flg = '0' AND t.status = '3' THEN t.story_point ELSE 0 END), 0)::int AS completed_story_point
FROM t_task t
WHERE t.goal_id = $1 AND t.account_id = $2;

-- name: GetGoalPointsByIdeal :many
SELECT g.id AS goal_id,
       COALESCE(SUM(CASE WHEN t.del_flg = '0' AND t.status <> '4' THEN t.story_point ELSE 0 END), 0)::int AS total_story_point,
       COALESCE(SUM(CASE WHEN t.del_flg = '0' AND t.status = '3' THEN t.story_point ELSE 0 END), 0)::int AS completed_story_point
FROM t_goal g
LEFT JOIN t_task t ON t.goal_id = g.id
WHERE g.ideal_id = $1 AND g.account_id = $2 AND g.del_flg = '0'
GROUP BY g.id;

-- name: CountTasksByGoal :one
SELECT COUNT(*) FROM t_task
WHERE goal_id = $1 AND account_id = $2 AND del_flg = '0';

-- name: CreateTask :one
INSERT INTO t_task (goal_id, account_id, title, description, priority, story_point, task_cd)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, goal_id, account_id, title, description, priority, status, story_point, task_cd, del_flg, created_at, updated_at;

-- name: ListTasksByGoal :many
SELECT id, goal_id, account_id, title, description, priority, status, story_point, task_cd, del_flg, created_at, updated_at
FROM t_task
WHERE goal_id = $1 AND account_id = $2 AND del_flg = '0'
ORDER BY priority DESC, created_at ASC;

-- name: ListTasksByIdealPriority :many
SELECT t.id, t.goal_id, t.account_id, t.title, t.description, t.priority, t.status, t.story_point, t.task_cd, t.del_flg, t.created_at, t.updated_at
FROM t_task t
INNER JOIN t_goal g ON t.goal_id = g.id
WHERE g.ideal_id = $1 AND t.account_id = $2 AND t.del_flg = '0' AND g.del_flg = '0' AND t.status NOT IN ('3', '4')
ORDER BY t.priority DESC, t.created_at ASC;

-- name: GetTaskByID :one
SELECT id, goal_id, account_id, title, description, priority, status, story_point, task_cd, del_flg, created_at, updated_at
FROM t_task
WHERE id = $1 AND account_id = $2 AND del_flg = '0';

-- name: UpdateTask :one
UPDATE t_task
SET title = $3, description = $4, priority = $5, status = $6, story_point = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, goal_id, account_id, title, description, priority, status, story_point, task_cd, del_flg, created_at, updated_at;

-- name: CompleteTask :one
UPDATE t_task
SET status = '3', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, goal_id, account_id, title, description, priority, status, story_point, task_cd, del_flg, created_at, updated_at;

-- name: DeleteTask :exec
UPDATE t_task
SET del_flg = '1', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0';

-- name: CancelTask :one
UPDATE t_task
SET status = '4', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND account_id = $2 AND del_flg = '0'
RETURNING id, goal_id, account_id, title, description, priority, status, story_point, task_cd, del_flg, created_at, updated_at;

-- name: GetKblogBookletLink :one
SELECT id, ideal_id, account_id, kblog_booklet_id, created_at
FROM t_kblog_booklet_link
WHERE ideal_id = $1 AND account_id = $2;

-- name: CreateKblogBookletLink :one
INSERT INTO t_kblog_booklet_link (ideal_id, account_id, kblog_booklet_id)
VALUES ($1, $2, $3)
RETURNING id, ideal_id, account_id, kblog_booklet_id, created_at;
