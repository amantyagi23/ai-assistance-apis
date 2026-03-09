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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserID     = errors.New("invalid user ID")
)

// User represents the user model in MongoDB
type User struct {
	UserID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email           string             `bson:"email" json:"email"`
	Password        string             `bson:"password" json:"-"` // "-" excludes from JSON
	Name            string             `bson:"name" json:"name"`
	IsEmailVerified bool               `bson:"is_email_verified" json:"is_email_verified"`
	ProfilePic      *string            `bson:"profile_pic,omitempty" json:"profile_pic,omitempty"`
	LastLoginAt     *time.Time         `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time         `bson:"deleted_at,omitempty" json:"-"`
}

// CreateUserRequest represents the data needed to create a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// UpdateUserRequest represents the data needed to update a user
type UpdateUserRequest struct {
	Name       string `json:"name,omitempty"`
	ProfilePic string `json:"profile_pic,omitempty"`
}

// UserRepo handles database operations for users
type UserRepo struct {
	collection *mongo.Collection
}

// NewUserRepo creates a new user repository instance
func NewUserRepo() *UserRepo {
	return &UserRepo{
		collection: db.DB.GetCollection("users"),
	}
}

// Create inserts a new user into the database
func (r *UserRepo) Create(ctx context.Context, user *User) error {
	// Check if user already exists
	existingUser, err := r.FindByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return ErrUserAlreadyExists
	}

	user.UserID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsEmailVerified = false

	_, err = r.collection.InsertOne(ctx, user)
	return err
}

// FindByID retrieves a user by their ID
func (r *UserRepo) FindByID(ctx context.Context, id string) (*User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var user User
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": nil,
	}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmail retrieves a user by their email
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{
		"email":      email,
		"deleted_at": nil,
	}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates an existing user
func (r *UserRepo) Update(ctx context.Context, id string, updates *UpdateUserRequest) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidUserID
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Build update fields
	setFields := bson.M{}
	if updates.Name != "" {
		setFields["name"] = updates.Name
	}

	if updates.ProfilePic != "" {
		setFields["profile_pic"] = updates.ProfilePic
	}

	if len(setFields) > 0 {
		update["$set"] = setFields
	}

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objID, "deleted_at": nil},
		update,
	)

	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidUserID
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objID, "deleted_at": nil},
		bson.M{
			"$set": bson.M{
				"deleted_at": now,
				"updated_at": now,
				"is_active":  false,
			},
		},
	)

	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidUserID
	}

	now := time.Now()
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"last_login_at": now,
				"updated_at":    now,
			},
		},
	)

	return err
}

// VerifyEmail marks a user's email as verified
func (r *UserRepo) VerifyEmail(ctx context.Context, email string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"email": email},
		bson.M{
			"$set": bson.M{
				"is_email_verified": true,
				"updated_at":        time.Now(),
			},
		},
	)

	return err
}

// ChangePassword updates a user's password
func (r *UserRepo) ChangePassword(ctx context.Context, id, hashedPassword string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidUserID
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"password":   hashedPassword,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}

// Count returns the total number of active users
func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"deleted_at": nil})
}

// Search searches users by name or email
func (r *UserRepo) Search(ctx context.Context, query string, page, limit int64) ([]User, int64, error) {
	skip := (page - 1) * limit

	filter := bson.M{
		"deleted_at": nil,
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
