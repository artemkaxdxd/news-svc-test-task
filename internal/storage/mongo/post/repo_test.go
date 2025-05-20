package post

import (
	"context"
	"fmt"
	"log/slog"
	"news-svc/config"
	"news-svc/internal/entity/post"

	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Also integration tests for handler <-> service <-> repo could be added,
// but considering there are already present tests for repo with dockertest,
// and unit tests with mocked service and repo layer are present,
// i think there is no need to overcomplicate. Imo these are enough.

var testMongoClient *mongo.Client

func setupDockerMongoDB() (*dockertest.Pool, *dockertest.Resource, string, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not connect to docker: %w", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "6.0",
		Env:        []string{},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return nil, nil, "", fmt.Errorf("could not start resource: %w", err)
	}

	mongoURI := fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp"))

	if err = pool.Retry(func() error {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		testMongoClient, err = mongo.Connect(options.Client().ApplyURI(mongoURI))
		if err != nil {
			return err
		}
		return testMongoClient.Ping(ctx, nil)
	}); err != nil {
		return nil, nil, "", fmt.Errorf("could not connect to docker: %w", err)
	}

	return pool, resource, mongoURI, nil
}

func setupTest(t *testing.T) (*mongo.Database, repo, func()) {
	require.NotNil(t, testMongoClient, "MongoDB client not initialized")

	ctx := context.Background()

	dbName := "test_db_" + bson.NewObjectID().Hex()
	db := testMongoClient.Database(dbName)

	repo := New(db)

	err := repo.EnsureIndexes(ctx)
	require.NoError(t, err)

	cleanup := func() {
		err := db.Drop(ctx)
		assert.NoError(t, err)
	}

	return db, repo, cleanup
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, resource, mongoURI, err := setupDockerMongoDB()
	if err != nil {
		slog.Error("could not setup Docker MongoDB", "err", err)
		return
	}

	testMongoClient, err = mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		slog.Error("could not connect to MongoDB", "err", err)
		return
	}

	code := m.Run()

	if err := testMongoClient.Disconnect(ctx); err != nil {
		slog.Error("rrror disconnecting from MongoDB", "err", err)
		return
	}

	if err := pool.Purge(resource); err != nil {
		slog.Error("could not purge resource", "err", err)
		return
	}

	slog.Info("exiting main", "code", code)
}

func createSamplePost() *post.Post {
	return &post.Post{
		Title:   "Test Title",
		Content: "Test Content",
	}
}

func createMultiplePosts(ctx context.Context, repo repo, count int) error {
	for i := range count {
		p := &post.Post{
			Title:   fmt.Sprintf("Title %s", string(rune(i+65))),
			Content: fmt.Sprintf("Content %s", string(rune(i+65))),
		}
		_, err := repo.Create(ctx, p)
		if err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	p := createSamplePost()
	id, err := repo.Create(ctx, p)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotZero(t, p.CreatedAt)
	assert.NotZero(t, p.UpdatedAt)
	assert.Equal(t, p.CreatedAt, p.UpdatedAt)
}

func TestGetByID(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	p := createSamplePost()
	id, err := repo.Create(ctx, p)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, p.Title, retrieved.Title)
	assert.Equal(t, p.Content, retrieved.Content)
	assert.NotZero(t, retrieved.CreatedAt)
	assert.NotZero(t, retrieved.UpdatedAt)

	nonExistentID := bson.NewObjectID().Hex()
	_, err = repo.GetByID(ctx, nonExistentID)
	assert.ErrorIs(t, err, config.ErrPostNotFound)

	_, err = repo.GetByID(ctx, "invalid-id")
	assert.Error(t, err)
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	numPosts := 5
	err := createMultiplePosts(ctx, repo, numPosts)
	require.NoError(t, err)

	posts, total, err := repo.GetAll(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(numPosts), total)
	assert.Len(t, posts, numPosts)

	posts, total, err = repo.GetAll(ctx, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(numPosts), total)
	assert.Len(t, posts, 2)

	posts, total, err = repo.GetAll(ctx, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(numPosts), total)
	assert.Len(t, posts, 2)

	posts, total, err = repo.GetAll(ctx, 10, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(numPosts), total)
	assert.Len(t, posts, 0)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	p := createSamplePost()
	id, err := repo.Create(ctx, p)
	require.NoError(t, err)

	retrievedBefore, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	updatedPost := &post.Post{
		ID:      id,
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	time.Sleep(10 * time.Millisecond)
	err = repo.Update(ctx, updatedPost)
	require.NoError(t, err)

	retrievedAfter, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrievedAfter.Title)
	assert.Equal(t, "Updated Content", retrievedAfter.Content)
	assert.Equal(t, retrievedBefore.CreatedAt, retrievedAfter.CreatedAt)
	assert.True(t, retrievedAfter.UpdatedAt.After(retrievedBefore.UpdatedAt))

	nonExistentPost := &post.Post{
		ID:      bson.NewObjectID().Hex(),
		Title:   "Non-existent Post",
		Content: "This post doesn't exist",
	}
	err = repo.Update(ctx, nonExistentPost)
	assert.ErrorIs(t, err, config.ErrPostNotFound)

	invalidPost := &post.Post{
		ID:      "invalid-id",
		Title:   "Invalid ID",
		Content: "This post has an invalid ID",
	}
	err = repo.Update(ctx, invalidPost)
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	p := createSamplePost()
	id, err := repo.Create(ctx, p)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, id)
	assert.ErrorIs(t, err, config.ErrPostNotFound)

	nonExistentID := bson.NewObjectID().Hex()
	err = repo.Delete(ctx, nonExistentID)
	assert.ErrorIs(t, err, config.ErrPostNotFound)

	err = repo.Delete(ctx, "invalid-id")
	assert.Error(t, err)
}

func TestSearch(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	//
	posts := []*post.Post{
		{Title: "Go Programming", Content: "Learning Go programming language"},
		{Title: "MongoDB Tutorial", Content: "Working with MongoDB in Go"},
		{Title: "RESTful API", Content: "Building RESTful APIs with Go"},
		{Title: "GraphQL API", Content: "Building GraphQL APIs with Go"},
		{Title: "Testing in Go", Content: "Writing tests for Go applications"},
	}

	for _, p := range posts {
		_, err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	foundPosts, total, err := repo.Search(ctx, "Go", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, foundPosts, 5)

	foundPosts, total, err = repo.Search(ctx, "MongoDB", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, foundPosts, 1)

	foundPosts, total, err = repo.Search(ctx, "API", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, foundPosts, 2)

	foundPosts, total, err = repo.Search(ctx, "Go", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, foundPosts, 2)

	foundPosts, total, err = repo.Search(ctx, "NonExistentTerm", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, foundPosts, 0)
}

func TestGetRecent(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTest(t)
	defer cleanup()

	numPosts := 5
	err := createMultiplePosts(ctx, repo, numPosts)
	require.NoError(t, err)

	recentPosts, err := repo.GetRecent(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, recentPosts, 3)

	assert.Equal(t, "Title E", recentPosts[0].Title)
	assert.Equal(t, "Title D", recentPosts[1].Title)
	assert.Equal(t, "Title C", recentPosts[2].Title)

	allRecentPosts, err := repo.GetRecent(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, allRecentPosts, 5)
}

func TestEnsureIndexes(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupTest(t)
	defer cleanup()

	cursor, err := db.Collection(post.CollectionName).Indexes().List(ctx)
	require.NoError(t, err)

	var indexes []bson.M
	err = cursor.All(ctx, &indexes)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(indexes), 3)

	err = repo.EnsureIndexes(ctx)
	assert.NoError(t, err)
}
