package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type supabaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type TokenValidator struct {
	jwtSecret []byte
	issuer    string
	jwks      *keyfunc.JWKS
}

func NewTokenValidator(jwtSecret, issuer string) (*TokenValidator, error) {
	if issuer == "" {
		return nil, errors.New("SUPABASE_JWT_ISSUER is required")
	}

	tv := &TokenValidator{issuer: issuer}
	if jwtSecret != "" {
		tv.jwtSecret = []byte(jwtSecret)
	}

	jwksURL := strings.TrimRight(issuer, "/") + "/.well-known/jwks.json"
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshTimeout:  10 * time.Second,
	})
	if err != nil {
		if len(tv.jwtSecret) == 0 {
			return nil, fmt.Errorf("failed to load Supabase JWKS from %s: %w", jwksURL, err)
		}
		logrus.WithError(err).Warn("JWKS fetch failed; HS256 fallback only")
	} else {
		tv.jwks = jwks
	}

	return tv, nil
}

func (tv *TokenValidator) signingKey(token *jwt.Token) (interface{}, error) {
	switch token.Method.(type) {
	case *jwt.SigningMethodHMAC:
		if len(tv.jwtSecret) == 0 {
			return nil, errors.New("HS256 token received but SUPABASE_JWT_SECRET is not set")
		}
		return tv.jwtSecret, nil
	case *jwt.SigningMethodECDSA, *jwt.SigningMethodRSA:
		if tv.jwks == nil {
			return nil, errors.New("RS256/ECDSA token received but JWKS could not be loaded")
		}
		return tv.jwks.Keyfunc(token)
	default:
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
}

func (tv *TokenValidator) ValidateToken(tokenString string) (*validator.ValidatedClaims, error) {
	claims := &supabaseClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, tv.signingKey, jwt.WithLeeway(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != tv.issuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	audienceValid := false
	for _, aud := range claims.Audience {
		if aud == "authenticated" {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return nil, errors.New("invalid audience")
	}

	if claims.Subject == "" {
		return nil, errors.New("missing subject claim")
	}

	return &validator.ValidatedClaims{
		RegisteredClaims: validator.RegisteredClaims{
			Subject: claims.Subject,
		},
	}, nil
}
