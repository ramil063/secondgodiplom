package auth

import (
	"context"
	"errors"
	"time"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
	"github.com/ramil063/secondgodiplom/internal/hash"
	"github.com/ramil063/secondgodiplom/internal/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Auth) SaveAccessToken(ctx context.Context, userID int, token string) (int, error) {
	tokenHash := hash.GetTokenHash(token)

	expiresAt := time.Now().Add(time.Duration(auth.TokenExpiredSeconds) * time.Second)
	row := s.Repository.Pool.QueryRow(
		ctx,
		"INSERT INTO oauth_access_token (token_hash, user_id, expires_at) VALUES ($1, $2, $3) RETURNING id",
		tokenHash,
		userID,
		expiresAt)

	var accessTokenId int
	err := row.Scan(&accessTokenId)

	if err != nil {
		return 0, errors.New("SaveAccessToken error in sql empty result")
	}

	return accessTokenId, nil
}

func (s *Auth) SaveRefreshToken(ctx context.Context, accessTokenId int, token string) error {
	tokenHash := hash.GetTokenHash(token)

	expiresAt := time.Now().Add(time.Duration(auth.TokenExpiredSeconds) * time.Second)
	exec, err := s.Repository.Pool.Exec(
		ctx,
		"INSERT INTO oauth_refresh_token (token_hash, access_token_id, expires_at) VALUES ($1, $2, $3)",
		tokenHash,
		accessTokenId,
		expiresAt)

	if err != nil {
		return errors.New("RegisterUser error in sql empty result")
	}

	if exec == nil {
		logger.WriteErrorLog("RegisterUser error in sql empty result")
		return errors.New("RegisterUser error in sql empty result")
	}

	rows := exec.RowsAffected()
	if rows != 1 {
		logger.WriteErrorLog("RegisterUser error expected to affect 1 row")
		return errors.New("RegisterUser expected to affect 1 row")
	}

	return nil
}

func (s *Auth) GetRefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshToken, error) {
	refreshTokenHash := hash.GetTokenHash(refreshToken)

	row := s.Repository.Pool.QueryRow(
		ctx,
		`SELECT
				ort.token_hash,
				oat.token_hash AS access_token_hash,
				oat.user_id,
				ort.expires_at
			FROM oauth_refresh_token ort
			    LEFT JOIN oauth_access_token oat on oat.id = ort.access_token_id
			WHERE ort.is_revoked=FALSE AND ort.token_hash = $1
			ORDER BY ort.created_at DESC
			LIMIT 1`,
		refreshTokenHash)

	var tokenHash []byte
	var accessTokenHash []byte
	var userID int
	var expiresAt time.Time

	err := row.Scan(&tokenHash, &accessTokenHash, &userID, &expiresAt)
	if err != nil {
		return nil, status.Error(codes.NotFound, "token not found")
	}

	token := &auth.RefreshToken{
		UserID:          userID,
		TokenHash:       string(tokenHash),
		AccessTokenHash: string(accessTokenHash),
		ExpiresAt:       expiresAt,
	}
	return token, nil
}

func (s *Auth) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	refreshTokenHash := hash.GetTokenHash(refreshToken)

	exec, err := s.Repository.Pool.Exec(
		ctx,
		"UPDATE oauth_refresh_token SET is_revoked=TRUE WHERE token_hash = $1",
		refreshTokenHash)

	if err != nil {
		return errors.New("RevokeRefreshToken error in sql empty result")
	}
	if exec == nil {
		logger.WriteErrorLog("RevokeRefreshToken error in sql empty result")
		return errors.New("RevokeRefreshToken error in sql empty result")
	}

	rows := exec.RowsAffected()
	if rows != 1 {
		logger.WriteErrorLog("RevokeRefreshToken error expected to affect 1 row")
		return errors.New("RevokeRefreshToken expected to affect 1 row")
	}
	return nil
}
