package ideal

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
	idealv1 "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/ideal"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateIdeal(ctx context.Context, req *connect.Request[idealv1.CreateIdealRequest]) (*connect.Response[idealv1.CreateIdealResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	created, err := s.queries.CreateIdeal(ctx, sqlc.CreateIdealParams{
		AccountID:   accountID,
		Title:       req.Msg.Title,
		Description: util.ToPgText(req.Msg.Description),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create ideal: %w", err))
	}

	return connect.NewResponse(&idealv1.CreateIdealResponse{Ideal: toProtoIdeal(created)}), nil
}

func (s *Service) ListIdeals(ctx context.Context, _ *connect.Request[idealv1.ListIdealsRequest]) (*connect.Response[idealv1.ListIdealsResponse], error) {
	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	rows, err := s.queries.ListIdealsByAccount(ctx, accountID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list ideals: %w", err))
	}

	ideals := make([]*idealv1.Ideal, 0, len(rows))
	for _, row := range rows {
		ideals = append(ideals, toProtoIdeal(row))
	}

	return connect.NewResponse(&idealv1.ListIdealsResponse{Ideals: ideals}), nil
}

func (s *Service) GetIdeal(ctx context.Context, req *connect.Request[idealv1.GetIdealRequest]) (*connect.Response[idealv1.GetIdealResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	row, err := s.queries.GetIdealByID(ctx, sqlc.GetIdealByIDParams{
		ID:        req.Msg.IdealId,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("ideal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get ideal: %w", err))
	}

	return connect.NewResponse(&idealv1.GetIdealResponse{Ideal: toProtoIdeal(row)}), nil
}

func (s *Service) UpdateIdeal(ctx context.Context, req *connect.Request[idealv1.UpdateIdealRequest]) (*connect.Response[idealv1.UpdateIdealResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	row, err := s.queries.UpdateIdeal(ctx, sqlc.UpdateIdealParams{
		ID:          req.Msg.IdealId,
		AccountID:   accountID,
		Title:       req.Msg.Title,
		Description: util.ToPgText(req.Msg.Description),
		Status:      idealStatusToDB(req.Msg.Status),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("ideal not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update ideal: %w", err))
	}

	return connect.NewResponse(&idealv1.UpdateIdealResponse{Ideal: toProtoIdeal(row)}), nil
}

func (s *Service) DeleteIdeal(ctx context.Context, req *connect.Request[idealv1.DeleteIdealRequest]) (*connect.Response[idealv1.DeleteIdealResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := s.queries.DeleteIdeal(ctx, sqlc.DeleteIdealParams{
		ID:        req.Msg.IdealId,
		AccountID: accountID,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete ideal: %w", err))
	}

	return connect.NewResponse(&idealv1.DeleteIdealResponse{}), nil
}

func toProtoIdeal(row sqlc.TIdeal) *idealv1.Ideal {
	return &idealv1.Ideal{
		IdealId:     row.ID,
		Title:       row.Title,
		Description: util.PgTextString(row.Description),
		Status:      idealStatusFromDB(row.Status),
		CreatedAt:   util.FormatPgTimestamp(row.CreatedAt),
		UpdatedAt:   util.FormatPgTimestamp(row.UpdatedAt),
	}
}

func idealStatusToDB(status idealv1.IdealStatus) string {
	switch status {
	case idealv1.IdealStatus_IDEAL_STATUS_COMPLETED:
		return "2"
	case idealv1.IdealStatus_IDEAL_STATUS_ARCHIVED:
		return "3"
	default:
		return "1"
	}
}

func idealStatusFromDB(status string) idealv1.IdealStatus {
	switch status {
	case "2":
		return idealv1.IdealStatus_IDEAL_STATUS_COMPLETED
	case "3":
		return idealv1.IdealStatus_IDEAL_STATUS_ARCHIVED
	default:
		return idealv1.IdealStatus_IDEAL_STATUS_IN_PROGRESS
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
