package user

import (
	"context"
	"errors"
	"time"

	db "github.com/amantyagi23/ai-assistance/internal/shared/infra/db/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session has expired")
	ErrInvalidSessionID    = errors.New("invalid session ID")
	ErrMaxSessionsExceeded = errors.New("maximum number of active sessions exceeded")
)

// SessionType defines the type of session
type SessionType string

const (
	SessionTypeLogin         SessionType = "login"
	SessionTypeRefresh       SessionType = "refresh"
	SessionTypeAPI           SessionType = "api"
	SessionTypePasswordReset SessionType = "password_reset"
)

// SessionStatus defines the status of a session
type SessionStatus string

const (
	SessionStatusActive     SessionStatus = "active"
	SessionStatusExpired    SessionStatus = "expired"
	SessionStatusRevoked    SessionStatus = "revoked"
	SessionStatusTerminated SessionStatus = "terminated"
)

// UserSession represents a user session in MongoDB
type UserSession struct {
	SessionID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	SessionToken string             `bson:"session_token" json:"session_token"`
	RefreshToken string             `bson:"refresh_token,omitempty" json:"-"`
	SessionType  SessionType        `bson:"session_type" json:"session_type"`
	Status       SessionStatus      `bson:"status" json:"status"`

	// Device Information
	DeviceInfo map[string]interface{} `bson:"device_info,omitempty" json:"device_info,omitempty"`
	IPAddress  string                 `bson:"ip_address" json:"ip_address"`
	UserAgent  string                 `bson:"user_agent" json:"user_agent"`

	// Location Information
	Location map[string]interface{} `bson:"location,omitempty" json:"location,omitempty"`

	// Timestamps
	IssuedAt     time.Time  `bson:"issued_at" json:"issued_at"`
	ExpiresAt    time.Time  `bson:"expires_at" json:"expires_at"`
	LastActivity time.Time  `bson:"last_activity" json:"last_activity"`
	RevokedAt    *time.Time `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}

// UserSessionRepo handles database operations for user sessions
type UserSessionRepo struct {
	collection         *mongo.Collection
	maxSessionsPerUser int
}

// NewSessionRepository creates a new session repository instance
func NewUserSessionRepo() *UserSessionRepo {
	return &UserSessionRepo{
		collection:         db.DB.GetCollection("user_sessions"),
		maxSessionsPerUser: 5, // Maximum 5 active sessions per user
	}
}

// CreateSession creates a new user session
func (r *UserSessionRepo) CreateSession(ctx context.Context, session *UserSession) error {
	// Check if user has exceeded maximum active sessions
	activeCount, err := r.countActiveSessions(ctx, session.UserID)
	if err != nil {
		return err
	}

	if activeCount >= int64(r.maxSessionsPerUser) {
		// Remove oldest session to make room for new one
		if err := r.removeOldestSession(ctx, session.UserID); err != nil {
			return err
		}
	}

	session.SessionID = primitive.NewObjectID()
	session.IssuedAt = time.Now()
	session.LastActivity = time.Now()
	session.Status = SessionStatusActive

	_, err = r.collection.InsertOne(ctx, session)
	return err
}

// FindBySessionToken finds a session by its token
func (r *UserSessionRepo) FindBySessionToken(ctx context.Context, token string) (*UserSession, error) {
	var session UserSession
	err := r.collection.FindOne(ctx, bson.M{
		"session_token": token,
		"status":        SessionStatusActive,
	}).Decode(&session)

	if err == mongo.ErrNoDocuments {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		r.ExpireSession(ctx, session.SessionID.Hex())
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// FindByRefreshToken finds a session by refresh token
func (r *UserSessionRepo) FindByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error) {
	var session UserSession
	err := r.collection.FindOne(ctx, bson.M{
		"refresh_token": refreshToken,
		"status":        SessionStatusActive,
	}).Decode(&session)

	if err == mongo.ErrNoDocuments {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// FindActiveSessionsByUserID finds all active sessions for a user
func (r *UserSessionRepo) FindActiveSessionsByUserID(ctx context.Context, userID primitive.ObjectID) ([]UserSession, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":    userID,
		"status":     SessionStatusActive,
		"expires_at": bson.M{"$gt": time.Now()},
	})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []UserSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *UserSessionRepo) FindLatestActiveSessionByUserID(ctx context.Context, userID primitive.ObjectID) (*UserSession, error) {
	// Add sort option and limit to 1
	opts := options.FindOne().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	var session UserSession
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":    userID,
		"status":     SessionStatusActive,
		"expires_at": bson.M{"$gt": time.Now()},
	}, opts).Decode(&session)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // or return nil, ErrNoActiveSession
		}
		return nil, err
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity timestamp
func (r *UserSessionRepo) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return ErrInvalidSessionID
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"last_activity": time.Now(),
			},
		},
	)

	return err
}

// RevokeSession revokes a specific session
func (r *UserSessionRepo) RevokeSession(ctx context.Context, sessionID string) error {
	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return ErrInvalidSessionID
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"status":     SessionStatusRevoked,
				"revoked_at": now,
			},
		},
	)

	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *UserSessionRepo) RevokeAllUserSessions(ctx context.Context, userID primitive.ObjectID, excludeSessionID ...string) error {
	filter := bson.M{
		"user_id": userID,
		"status":  SessionStatusActive,
	}

	// Exclude specific session if provided
	if len(excludeSessionID) > 0 && excludeSessionID[0] != "" {
		excludeObjID, err := primitive.ObjectIDFromHex(excludeSessionID[0])
		if err == nil {
			filter["_id"] = bson.M{"$ne": excludeObjID}
		}
	}

	now := time.Now()
	_, err := r.collection.UpdateMany(ctx,
		filter,
		bson.M{
			"$set": bson.M{
				"status":     SessionStatusRevoked,
				"revoked_at": now,
			},
		},
	)

	return err
}

// ExpireSession marks a session as expired
func (r *UserSessionRepo) ExpireSession(ctx context.Context, sessionID string) error {
	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return ErrInvalidSessionID
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"status": SessionStatusExpired,
			},
		},
	)

	return err
}

// CleanupExpiredSessions removes expired sessions older than the given duration
func (r *UserSessionRepo) CleanupExpiredSessions(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result, err := r.collection.DeleteMany(ctx, bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"status": SessionStatusExpired},
			{"status": SessionStatusRevoked},
		},
		"issued_at": bson.M{"$lt": cutoffTime},
	})

	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// GetSessionStats gets session statistics for a user
func (r *UserSessionRepo) GetSessionStats(ctx context.Context, userID primitive.ObjectID) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats[result.ID] = result.Count
		}
	}

	return stats, nil
}

// Helper Methods

// countActiveSessions counts active sessions for a user
func (r *UserSessionRepo) countActiveSessions(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"status":     SessionStatusActive,
		"expires_at": bson.M{"$gt": time.Now()},
	})
}

// removeOldestSession removes the oldest active session for a user
func (r *UserSessionRepo) removeOldestSession(ctx context.Context, userID primitive.ObjectID) error {
	opts := options.FindOne().
		SetSort(bson.D{{Key: "last_activity", Value: 1}})

	var oldestSession UserSession
	err := r.collection.FindOne(ctx,
		bson.M{
			"user_id": userID,
			"status":  SessionStatusActive,
		},
		opts,
	).Decode(&oldestSession)

	if err != nil {
		return err
	}

	return r.RevokeSession(ctx, oldestSession.SessionID.Hex())
}

// GetUserActiveSessions gets all active sessions for a user with device info
func (r *UserSessionRepo) GetUserActiveSessions(ctx context.Context, userID primitive.ObjectID) ([]map[string]interface{}, error) {
	sessions, err := r.FindActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, map[string]interface{}{
			"id":            session.SessionID.Hex(),
			"session_type":  session.SessionType,
			"device_info":   session.DeviceInfo,
			"ip_address":    session.IPAddress,
			"user_agent":    session.UserAgent,
			"location":      session.Location,
			"last_activity": session.LastActivity,
			"expires_at":    session.ExpiresAt,
		})
	}

	return result, nil
}

// CreateSessionIndexes creates necessary indexes for the sessions collection
func (r *UserSessionRepo) CreateSessionIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "session_token", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "refresh_token", Value: 1},
			},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "expires_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
