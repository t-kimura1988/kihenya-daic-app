package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/t-kimura1988/kihenya-daic-app/internal/account"
	"github.com/t-kimura1988/kihenya-daic-app/internal/db/sqlc"
	"github.com/t-kimura1988/kihenya-daic-app/internal/util"
	taskv1 "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/task"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateTask(ctx context.Context, req *connect.Request[taskv1.CreateTaskRequest]) (*connect.Response[taskv1.CreateTaskResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	goalRow, err := s.queries.GetGoalByID(ctx, sqlc.GetGoalByIDParams{
		ID:        req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("goal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("verify goal: %w", err))
	}
	if goalRow.Achievement == "2" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot add tasks to an achieved goal"))
	}

	taskCd, err := util.GenerateTaskCode()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate task code: %w", err))
	}

	created, err := s.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		GoalID:      req.Msg.GoalId,
		AccountID:   accountID,
		Title:       req.Msg.Title,
		Description: util.ToPgText(req.Msg.Description),
		Priority:    priorityToDB(req.Msg.Priority),
		StoryPoint:  normalizeStoryPoint(req.Msg.StoryPoint),
		TaskCd:      taskCd,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create task: %w", err))
	}

	return connect.NewResponse(&taskv1.CreateTaskResponse{Task: toProtoTask(created)}), nil
}

func (s *Service) ListTasksByGoal(ctx context.Context, req *connect.Request[taskv1.ListTasksByGoalRequest]) (*connect.Response[taskv1.ListTasksByGoalResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	rows, err := s.queries.ListTasksByGoal(ctx, sqlc.ListTasksByGoalParams{
		GoalID:    req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list tasks: %w", err))
	}

	tasks := make([]*taskv1.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, toProtoTask(row))
	}

	return connect.NewResponse(&taskv1.ListTasksByGoalResponse{Tasks: tasks}), nil
}

func (s *Service) ListTasksByPriority(ctx context.Context, req *connect.Request[taskv1.ListTasksByPriorityRequest]) (*connect.Response[taskv1.ListTasksByPriorityResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	rows, err := s.queries.ListTasksByIdealPriority(ctx, sqlc.ListTasksByIdealPriorityParams{
		IdealID:   req.Msg.IdealId,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list tasks by priority: %w", err))
	}

	tasks := make([]*taskv1.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, toProtoTask(row))
	}

	return connect.NewResponse(&taskv1.ListTasksByPriorityResponse{Tasks: tasks}), nil
}

func (s *Service) UpdateTask(ctx context.Context, req *connect.Request[taskv1.UpdateTaskRequest]) (*connect.Response[taskv1.UpdateTaskResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	row, err := s.queries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:          req.Msg.TaskId,
		AccountID:   accountID,
		Title:       req.Msg.Title,
		Description: util.ToPgText(req.Msg.Description),
		Priority:    priorityToDB(req.Msg.Priority),
		Status:      taskStatusToDB(req.Msg.Status),
		StoryPoint:  normalizeStoryPoint(req.Msg.StoryPoint),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("task not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update task: %w", err))
	}

	return connect.NewResponse(&taskv1.UpdateTaskResponse{Task: toProtoTask(row)}), nil
}

func (s *Service) CompleteTask(ctx context.Context, req *connect.Request[taskv1.CompleteTaskRequest]) (*connect.Response[taskv1.CompleteTaskResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	row, err := s.queries.CompleteTask(ctx, sqlc.CompleteTaskParams{
		ID:        req.Msg.TaskId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("task not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("complete task: %w", err))
	}

	return connect.NewResponse(&taskv1.CompleteTaskResponse{Task: toProtoTask(row)}), nil
}

func (s *Service) CancelTask(ctx context.Context, req *connect.Request[taskv1.CancelTaskRequest]) (*connect.Response[taskv1.CancelTaskResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	row, err := s.queries.CancelTask(ctx, sqlc.CancelTaskParams{
		ID:        req.Msg.TaskId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("task not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cancel task: %w", err))
	}

	return connect.NewResponse(&taskv1.CancelTaskResponse{Task: toProtoTask(row)}), nil
}

func (s *Service) DeleteTask(ctx context.Context, req *connect.Request[taskv1.DeleteTaskRequest]) (*connect.Response[taskv1.DeleteTaskResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := s.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:        req.Msg.TaskId,
		AccountID: accountID,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete task: %w", err))
	}

	return connect.NewResponse(&taskv1.DeleteTaskResponse{}), nil
}

func toProtoTask(row sqlc.TTask) *taskv1.Task {
	return &taskv1.Task{
		TaskId:      row.ID,
		GoalId:      row.GoalID,
		Title:       row.Title,
		Description: util.PgTextString(row.Description),
		Priority:    priorityFromDB(row.Priority),
		Status:      taskStatusFromDB(row.Status),
		StoryPoint:  row.StoryPoint,
		TaskCd:      row.TaskCd,
		CreatedAt:   util.FormatPgTimestamp(row.CreatedAt),
		UpdatedAt:   util.FormatPgTimestamp(row.UpdatedAt),
	}
}

func normalizeStoryPoint(v int32) int32 {
	if v < 0 {
		return 0
	}
	return v
}

func priorityToDB(priority taskv1.TaskPriority) string {
	switch priority {
	case taskv1.TaskPriority_TASK_PRIORITY_LOW:
		return "1"
	case taskv1.TaskPriority_TASK_PRIORITY_HIGH:
		return "3"
	case taskv1.TaskPriority_TASK_PRIORITY_URGENT:
		return "4"
	default:
		return "2"
	}
}

func priorityFromDB(priority string) taskv1.TaskPriority {
	switch priority {
	case "1":
		return taskv1.TaskPriority_TASK_PRIORITY_LOW
	case "3":
		return taskv1.TaskPriority_TASK_PRIORITY_HIGH
	case "4":
		return taskv1.TaskPriority_TASK_PRIORITY_URGENT
	default:
		return taskv1.TaskPriority_TASK_PRIORITY_MEDIUM
	}
}

func taskStatusToDB(status taskv1.TaskStatus) string {
	switch status {
	case taskv1.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return "2"
	case taskv1.TaskStatus_TASK_STATUS_COMPLETED:
		return "3"
	case taskv1.TaskStatus_TASK_STATUS_CANCELLED:
		return "4"
	case taskv1.TaskStatus_TASK_STATUS_WAITING:
		return "5"
	case taskv1.TaskStatus_TASK_STATUS_BLOCKED:
		return "6"
	case taskv1.TaskStatus_TASK_STATUS_WORK_DONE:
		return "7"
	default:
		return "1"
	}
}

func taskStatusFromDB(status string) taskv1.TaskStatus {
	switch status {
	case "2":
		return taskv1.TaskStatus_TASK_STATUS_IN_PROGRESS
	case "3":
		return taskv1.TaskStatus_TASK_STATUS_COMPLETED
	case "4":
		return taskv1.TaskStatus_TASK_STATUS_CANCELLED
	case "5":
		return taskv1.TaskStatus_TASK_STATUS_WAITING
	case "6":
		return taskv1.TaskStatus_TASK_STATUS_BLOCKED
	case "7":
		return taskv1.TaskStatus_TASK_STATUS_WORK_DONE
	default:
		return taskv1.TaskStatus_TASK_STATUS_NOT_STARTED
	}
}

func validate(msg proto.Message) error {
	validator, err := protovalidate.New()
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("validation error: %w", err))
	}
	if err := validator.Validate(msg); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("validation failed: %w", err))
	}
	return nil
}
