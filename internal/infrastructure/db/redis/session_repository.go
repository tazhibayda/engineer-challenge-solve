package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/auth"
)

const (
	sessionPrefix = "session:"
	familyPrefix  = "family:"
)

type SessionRepository struct {
	client *redis.Client
}

const userSessionsPrefix = "user_sessions:"

func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

func (r *SessionRepository) Create(ctx context.Context, s *auth.Session) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	sessionKey := sessionPrefix + s.ID.String()
	familyKey := familyPrefix + s.FamilyID.String()
	userSessionsKey := userSessionsPrefix + s.UserID.String()

	pipe := r.client.Pipeline()
	ttl := time.Until(s.ExpiresAt)

	pipe.Set(ctx, sessionKey, data, ttl)
	pipe.SAdd(ctx, familyKey, s.ID.String())
	pipe.Expire(ctx, familyKey, ttl)

	pipe.SAdd(ctx, userSessionsKey, s.ID.String())
	pipe.Expire(ctx, userSessionsKey, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *SessionRepository) SessionByID(ctx context.Context, id uuid.UUID) (*auth.Session, error) {
	key := sessionPrefix + id.String()
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, auth.ErrSessionExpired
		}
		return nil, err
	}

	var s auth.Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) RevokeRefresh(ctx context.Context, id uuid.UUID) error {
	s, err := r.SessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, auth.ErrSessionExpired) {
			return nil
		}
		return err
	}

	if s.IsRevoked {
		return nil
	}

	s.IsRevoked = true

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	ttl := time.Until(s.ExpiresAt)
	if ttl <= 0 {
		return nil
	}

	return r.client.Set(ctx, sessionPrefix+id.String(), data, ttl).Err()
}

func (r *SessionRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	familyKey := familyPrefix + familyID.String()

	sessionIDs, err := r.client.SMembers(ctx, familyKey).Result()
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	keysToDelete := make([]string, 0, len(sessionIDs)+1)
	for _, id := range sessionIDs {
		keysToDelete = append(keysToDelete, sessionPrefix+id)
	}
	keysToDelete = append(keysToDelete, familyKey)

	return r.client.Del(ctx, keysToDelete...).Err()
}

func (r *SessionRepository) MarkRotated(ctx context.Context, id uuid.UUID) error {
	s, err := r.SessionByID(ctx, id)
	if err != nil {
		return err
	}
	s.IsRevoked = true

	data, _ := json.Marshal(s)
	ttl := time.Until(s.ExpiresAt)

	return r.client.Set(ctx, sessionPrefix+id.String(), data, ttl).Err()
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	userSessionsKey := userSessionsPrefix + userID.String()

	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	keysToDelete := make([]string, 0, len(sessionIDs)+1)
	for _, id := range sessionIDs {
		keysToDelete = append(keysToDelete, sessionPrefix+id)
	}
	keysToDelete = append(keysToDelete, userSessionsKey)

	return r.client.Del(ctx, keysToDelete...).Err()
}
