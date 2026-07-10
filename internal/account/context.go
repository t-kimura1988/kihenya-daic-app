package account

import (
	"context"
	"errors"
)

type (
	accountIDKey  struct{}
	userIDKey     struct{}
	authHeaderKey struct{}
	emailKey      struct{}
)

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func GetAuthHeaderFromContext(ctx context.Context) (string, error) {
	authHeader, ok := ctx.Value(authHeaderKey{}).(string)
	if !ok {
		return "", errors.New("auth header not found in context")
	}
	return authHeader, nil
}

func SetAuthHeaderInContext(ctx context.Context, authHeader string) context.Context {
	return context.WithValue(ctx, authHeaderKey{}, authHeader)
}

func GetAccountIDFromContext(ctx context.Context) (int64, error) {
	accountID, ok := ctx.Value(accountIDKey{}).(int64)
	if !ok {
		return 0, errors.New("account ID not found in context")
	}
	return accountID, nil
}

func SetAccountIDInContext(ctx context.Context, accountID int64) context.Context {
	return context.WithValue(ctx, accountIDKey{}, accountID)
}

func GetEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(emailKey{}).(string)
	if !ok {
		return "", errors.New("email not found in context")
	}
	return email, nil
}

func SetEmailInContext(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailKey{}, email)
}
