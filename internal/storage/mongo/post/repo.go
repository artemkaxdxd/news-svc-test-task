package post

import (
	"context"
	"errors"
	"news-svc/config"
	"news-svc/internal/entity/post"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repo struct {
	db *mongo.Database
}

func New(db *mongo.Database) repo {
	return repo{db}
}

func (r repo) Create(ctx context.Context, p *post.Post) (string, error) {
	coll := r.db.Collection(post.CollectionName)

	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	result, err := coll.InsertOne(ctx, p)
	if err != nil {
		return "", err
	}

	oid, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return "", errors.New("failed to get inserted ID")
	}

	return oid.Hex(), nil
}

func (r repo) GetAll(ctx context.Context, page, limit int64) (posts []*post.Post, total int64, err error) {
	coll := r.db.Collection(post.CollectionName)

	skip := max((page-1)*limit, 0)

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	total, err = coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, 0, err
	}

	cursor, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &posts)
	return
}

func (r repo) GetByID(ctx context.Context, id string) (*post.Post, error) {
	coll := r.db.Collection(post.CollectionName)

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var post post.Post
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, config.ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}

func (r repo) Update(ctx context.Context, p *post.Post) error {
	coll := r.db.Collection(post.CollectionName)

	objID, err := bson.ObjectIDFromHex(p.ID)
	if err != nil {
		return err
	}

	p.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"title":      p.Title,
			"content":    p.Content,
			"updated_at": p.UpdatedAt,
		},
	}

	result, err := coll.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return config.ErrPostNotFound
	}

	return nil
}

func (r repo) Delete(ctx context.Context, id string) error {
	coll := r.db.Collection(post.CollectionName)

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return config.ErrPostNotFound
	}

	return nil
}

func (r repo) Search(ctx context.Context, query string, page, limit int64) (posts []*post.Post, total int64, err error) {
	coll := r.db.Collection(post.CollectionName)

	skip := max((page-1)*limit, 0)

	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"content": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	total, err = coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &posts)
	return
}

func (r repo) GetRecent(ctx context.Context, limit int64) (posts []*post.Post, err error) {
	coll := r.db.Collection(post.CollectionName)

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &posts)
	return
}

func (r repo) EnsureIndexes(ctx context.Context) error {
	coll := r.db.Collection(post.CollectionName)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "content", Value: "text"},
			},
			Options: options.Index().SetName("title_content_text"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_desc"),
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}
