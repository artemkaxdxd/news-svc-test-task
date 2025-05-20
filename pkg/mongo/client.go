package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Database interface {
	// Instance returns the primary Mongo database instance.
	Instance() *mongo.Database
	// Close disconnects the Mongo client.
	Close(ctx context.Context) error
	// Ping checks connectivity to Mongo.
	Ping(ctx context.Context) error
}

type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

// New creates a new Mongo instance using provided config and slog logger.
func New(ctx context.Context, user, password, host, port, name string) (*Mongo, error) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/?authSource=admin",
		user, password, host, port,
	)

	clientOpts := options.Client().ApplyURI(uri)

	ctxConn, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctxConn, nil); err != nil {
		return nil, err
	}

	database := client.Database(name)

	return &Mongo{client: client, db: database}, nil
}

// Instance returns the Mongo database.
func (m *Mongo) Instance() *mongo.Database {
	return m.db
}

// Close disconnects the Mongo client.
func (m *Mongo) Close(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("mongo client already disconnected")
	}
	return m.client.Disconnect(ctx)
}

// Ping checks the connection to MongoDB.
func (m *Mongo) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}
