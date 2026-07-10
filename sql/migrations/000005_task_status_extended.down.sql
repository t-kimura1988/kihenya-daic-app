UPDATE t_task SET status = '1' WHERE status IN ('5', '6', '7');

COMMENT ON COLUMN t_task.status IS '1:未着手 2:進行中 3:完了 4:取り消し';
