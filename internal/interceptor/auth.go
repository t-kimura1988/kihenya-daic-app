package interceptor

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/sirupsen/logrus"

	"github.com/t-kimura1988/kihenya-daic-app/internal/account"
	"github.com/t-kimura1988/kihenya-daic-app/internal/db/sqlc"
	accountv1 "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/account"
)

type AuthValidator interface {
	ValidateToken(tokenString string) (*validator.ValidatedClaims, error)
}

type AuthInterceptor struct {
	validator AuthValidator
	queries   *sqlc.Queries
}

func NewAuthInterceptor(validator AuthValidator, queries *sqlc.Queries) *AuthInterceptor {
	return &AuthInterceptor{
		validator: validator,
		queries:   queries,
	}
}

func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		authHeader := req.Header().Get("authorization")
		if authHeader == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header is missing"))
		}

		token, err := extractToken(authHeader)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token format"))
		}

		claims, err := i.validator.ValidateToken(token)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
		}

		userID := claims.RegisteredClaims.Subject
		ctx = account.SetUserIDInContext(ctx, userID)
		ctx = account.SetAuthHeaderInContext(ctx, authHeader)

		if isCreateAccountRequest(req) {
			return next(ctx, req)
		}

		acct, err := i.queries.GetAccountByUID(ctx, userID)
		if err != nil {
			logrus.WithError(err).WithField("uid", userID).Warn("account not found in daic")
			connErr := connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("account not found"))
			authErrorDetail := &accountv1.AuthErrorDetail{
				Code:    "E6",
				Message: "account not found",
			}
			errDetail, _ := connect.NewErrorDetail(authErrorDetail)
			connErr.AddDetail(errDetail)
			return nil, connErr
		}

		ctx = account.SetAccountIDInContext(ctx, acct.ID)
		return next(ctx, req)
	}
}

func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func isCreateAccountRequest(req connect.AnyRequest) bool {
	if req.Spec().Procedure == "/daic.v1.account.AccountService/CreateAccount" {
		return true
	}
	_, ok := req.Any().(*accountv1.CreateAccountRequest)
	return ok
}

func extractToken(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}
