package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var DB *MongoDB

// MongoDB struct holds the client and database instances
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Config holds MongoDB connection configuration
type DBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// NewMongoDB creates a new MongoDB connection
func ConnectMongdb(config DBConfig) error {
	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Configure client options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB")

	// Get database instance
	DB = &MongoDB{
		Client:   client,
		Database: client.Database(config.Database),
	}

	return nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close() error {
	if m.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect MongoDB: %w", err)
	}

	log.Println("Successfully disconnected from MongoDB")
	return nil
}

// HealthCheck performs a health check on the database connection
func (m *MongoDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.Client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	return nil
}

// GetCollection returns a MongoDB collection
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
