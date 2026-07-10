package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/t-kimura1988/kihenya-daic-app/internal/db/sqlc"
	accountv1 "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/account"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	queries *sqlc.Queries
}

func NewHandler(queries *sqlc.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) CreateAccount(ctx context.Context, req *connect.Request[accountv1.CreateAccountRequest]) (*connect.Response[accountv1.CreateAccountResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	existing, err := h.queries.GetAccountByUID(ctx, userID)
	if err == nil {
		return connect.NewResponse(&accountv1.CreateAccountResponse{
			Account: toProtoAccount(existing),
		}), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get account: %w", err))
	}

	created, err := h.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		Uid:         userID,
		DisplayName: req.Msg.DisplayName,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create account: %w", err))
	}

	return connect.NewResponse(&accountv1.CreateAccountResponse{
		Account: toProtoAccount(created),
	}), nil
}

func (h *Handler) GetAccount(ctx context.Context, _ *connect.Request[accountv1.GetAccountRequest]) (*connect.Response[accountv1.GetAccountResponse], error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	acct, err := h.queries.GetAccountByUID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("account not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get account: %w", err))
	}

	return connect.NewResponse(&accountv1.GetAccountResponse{
		Account: toProtoAccount(acct),
	}), nil
}

func toProtoAccount(acct sqlc.TDaicAccount) *accountv1.Account {
	return &accountv1.Account{
		AccountId:   acct.ID,
		Uid:         acct.Uid,
		DisplayName: acct.DisplayName,
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
