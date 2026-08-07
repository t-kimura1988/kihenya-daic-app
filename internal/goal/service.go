package goal

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
	goalv1 "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/goal"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateGoal(ctx context.Context, req *connect.Request[goalv1.CreateGoalRequest]) (*connect.Response[goalv1.CreateGoalResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if _, err := s.queries.GetIdealByID(ctx, sqlc.GetIdealByIDParams{
		ID:        req.Msg.IdealId,
		AccountID: accountID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("ideal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("verify ideal: %w", err))
	}

	dependsOnID := req.Msg.GetDependsOnGoalId()
	if err := s.validateDependsOn(ctx, accountID, req.Msg.IdealId, 0, dependsOnID); err != nil {
		return nil, err
	}

	goalCd, err := util.GenerateGoalCode()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate goal code: %w", err))
	}

	created, err := s.queries.CreateGoal(ctx, sqlc.CreateGoalParams{
		IdealID:         req.Msg.IdealId,
		AccountID:       accountID,
		Title:           req.Msg.Title,
		Description:     util.ToPgText(req.Msg.Description),
		SortOrder:       req.Msg.SortOrder,
		DependsOnGoalID: util.ToPgInt8(dependsOnID),
		GoalCd:          goalCd,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create goal: %w", err))
	}

	return connect.NewResponse(&goalv1.CreateGoalResponse{Goal: toProtoGoal(created)}), nil
}

func (s *Service) ListGoalsByIdeal(ctx context.Context, req *connect.Request[goalv1.ListGoalsByIdealRequest]) (*connect.Response[goalv1.ListGoalsByIdealResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	rows, err := s.queries.ListGoalsByIdeal(ctx, sqlc.ListGoalsByIdealParams{
		IdealID:   req.Msg.IdealId,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list goals: %w", err))
	}

	pointRows, err := s.queries.GetGoalPointsByIdeal(ctx, sqlc.GetGoalPointsByIdealParams{
		IdealID:   req.Msg.IdealId,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("aggregate goal points: %w", err))
	}
	points := make(map[int64]sqlc.GetGoalPointsByIdealRow, len(pointRows))
	for _, p := range pointRows {
		points[p.GoalID] = p
	}

	goals := make([]*goalv1.Goal, 0, len(rows))
	for _, row := range rows {
		g := toProtoGoal(row)
		if p, ok := points[row.ID]; ok {
			g.TotalStoryPoint = p.TotalStoryPoint
			g.CompletedStoryPoint = p.CompletedStoryPoint
		}
		goals = append(goals, g)
	}

	return connect.NewResponse(&goalv1.ListGoalsByIdealResponse{Goals: goals}), nil
}

func (s *Service) UpdateGoal(ctx context.Context, req *connect.Request[goalv1.UpdateGoalRequest]) (*connect.Response[goalv1.UpdateGoalResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	existing, err := s.queries.GetGoalByID(ctx, sqlc.GetGoalByIDParams{
		ID:        req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("goal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get goal: %w", err))
	}

	dependsOn := existing.DependsOnGoalID
	if req.Msg.DependsOnGoalId != nil {
		dependsOn = util.ToPgInt8(*req.Msg.DependsOnGoalId)
		if err := s.validateDependsOn(ctx, accountID, existing.IdealID, req.Msg.GoalId, *req.Msg.DependsOnGoalId); err != nil {
			return nil, err
		}
	}

	row, err := s.queries.UpdateGoal(ctx, sqlc.UpdateGoalParams{
		ID:              req.Msg.GoalId,
		AccountID:       accountID,
		Title:           req.Msg.Title,
		Description:     util.ToPgText(req.Msg.Description),
		Status:          goalStatusToDB(req.Msg.Status),
		SortOrder:       req.Msg.SortOrder,
		DependsOnGoalID: dependsOn,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("goal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update goal: %w", err))
	}

	return connect.NewResponse(&goalv1.UpdateGoalResponse{Goal: toProtoGoal(row)}), nil
}

func (s *Service) DeleteGoal(ctx context.Context, req *connect.Request[goalv1.DeleteGoalRequest]) (*connect.Response[goalv1.DeleteGoalResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := s.queries.DeleteGoal(ctx, sqlc.DeleteGoalParams{
		ID:        req.Msg.GoalId,
		AccountID: accountID,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete goal: %w", err))
	}

	return connect.NewResponse(&goalv1.DeleteGoalResponse{}), nil
}

func (s *Service) CancelGoal(ctx context.Context, req *connect.Request[goalv1.CancelGoalRequest]) (*connect.Response[goalv1.CancelGoalResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	taskCount, err := s.queries.CountTasksByGoal(ctx, sqlc.CountTasksByGoalParams{
		GoalID:    req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("count tasks: %w", err))
	}
	if taskCount > 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("goal with tasks cannot be cancelled"))
	}

	row, err := s.queries.CancelGoal(ctx, sqlc.CancelGoalParams{
		ID:        req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("goal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cancel goal: %w", err))
	}

	return connect.NewResponse(&goalv1.CancelGoalResponse{Goal: toProtoGoal(row)}), nil
}

func (s *Service) AchieveGoal(ctx context.Context, req *connect.Request[goalv1.AchieveGoalRequest]) (*connect.Response[goalv1.AchieveGoalResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	points, err := s.queries.GetGoalPoints(ctx, sqlc.GetGoalPointsParams{
		GoalID:    req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("aggregate goal points: %w", err))
	}
	if points.TotalStoryPoint == 0 || points.CompletedStoryPoint < points.TotalStoryPoint {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("goal is not fully completed"))
	}

	row, err := s.queries.AchieveGoal(ctx, sqlc.AchieveGoalParams{
		ID:        req.Msg.GoalId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("goal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("achieve goal: %w", err))
	}

	return connect.NewResponse(&goalv1.AchieveGoalResponse{Goal: toProtoGoal(row)}), nil
}

func (s *Service) validateDependsOn(ctx context.Context, accountID, idealID, goalID, dependsOnID int64) error {
	if dependsOnID == 0 {
		return nil
	}
	if goalID != 0 && dependsOnID == goalID {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("goal cannot depend on itself"))
	}

	dep, err := s.queries.GetGoalByID(ctx, sqlc.GetGoalByIDParams{
		ID:        dependsOnID,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("depends_on goal not found"))
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("verify depends_on goal: %w", err))
	}
	if dep.IdealID != idealID {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("depends_on goal must belong to the same ideal"))
	}
	if goalID != 0 {
		if err := s.ensureNoCycle(ctx, accountID, goalID, dependsOnID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureNoCycle(ctx context.Context, accountID, goalID, dependsOnID int64) error {
	visited := make(map[int64]bool)
	current := dependsOnID
	for current != 0 {
		if current == goalID {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("circular goal dependency"))
		}
		if visited[current] {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("circular goal dependency"))
		}
		visited[current] = true

		goal, err := s.queries.GetGoalByID(ctx, sqlc.GetGoalByIDParams{
			ID:        current,
			AccountID: accountID,
		})
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("walk dependency: %w", err))
		}
		current = util.PgInt8Value(goal.DependsOnGoalID)
	}
	return nil
}

func toProtoGoal(row sqlc.TGoal) *goalv1.Goal {
	goal := &goalv1.Goal{
		GoalId:      row.ID,
		IdealId:     row.IdealID,
		Title:       row.Title,
		Description: util.PgTextString(row.Description),
		Status:      goalStatusFromDB(row.Status),
		Achievement: goalAchievementFromDB(row.Achievement),
		SortOrder:   row.SortOrder,
		GoalCd:      row.GoalCd,
		CreatedAt:   util.FormatPgTimestamp(row.CreatedAt),
		UpdatedAt:   util.FormatPgTimestamp(row.UpdatedAt),
	}
	if row.DependsOnGoalID.Valid {
		id := row.DependsOnGoalID.Int64
		goal.DependsOnGoalId = &id
	}
	return goal
}

func goalStatusToDB(status goalv1.GoalStatus) string {
	switch status {
	case goalv1.GoalStatus_GOAL_STATUS_IN_PROGRESS:
		return "2"
	case goalv1.GoalStatus_GOAL_STATUS_COMPLETED:
		return "3"
	case goalv1.GoalStatus_GOAL_STATUS_CANCELLED:
		return "4"
	default:
		return "1"
	}
}

func goalStatusFromDB(status string) goalv1.GoalStatus {
	switch status {
	case "2":
		return goalv1.GoalStatus_GOAL_STATUS_IN_PROGRESS
	case "3":
		return goalv1.GoalStatus_GOAL_STATUS_COMPLETED
	case "4":
		return goalv1.GoalStatus_GOAL_STATUS_CANCELLED
	default:
		return goalv1.GoalStatus_GOAL_STATUS_NOT_STARTED
	}
}

func goalAchievementFromDB(achievement string) goalv1.GoalAchievement {
	switch achievement {
	case "2":
		return goalv1.GoalAchievement_GOAL_ACHIEVEMENT_ACHIEVED
	default:
		return goalv1.GoalAchievement_GOAL_ACHIEVEMENT_NOT_ACHIEVED
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
