package post

import (
	"news-svc/config"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CollectionName = "posts"
)

type (
	Post struct {
		ID        string    `bson:"_id,omitempty" json:"id"`
		Title     string    `bson:"title" json:"title"`
		Content   string    `bson:"content" json:"content"`
		CreatedAt time.Time `bson:"created_at" json:"created_at"`
		UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	}

	mongoPost struct {
		ID        bson.ObjectID `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		Content   string        `bson:"content"`
		CreatedAt time.Time     `bson:"created_at"`
		UpdatedAt time.Time     `bson:"updated_at"`
	}
)

func (p Post) Validate() error {
	if p.Title == "" {
		return config.ErrEmptyTitle
	}
	if p.Content == "" {
		return config.ErrEmptyContent
	}
	return nil
}

func (p *Post) MarshalBSON() ([]byte, error) {
	if p.ID == "" {
		return bson.Marshal(bson.D{
			{Key: "title", Value: p.Title},
			{Key: "content", Value: p.Content},
			{Key: "created_at", Value: p.CreatedAt},
			{Key: "updated_at", Value: p.UpdatedAt},
		})
	}

	objectID, err := bson.ObjectIDFromHex(p.ID)
	if err != nil {
		return nil, err
	}

	return bson.Marshal(bson.D{
		{Key: "_id", Value: objectID},
		{Key: "title", Value: p.Title},
		{Key: "content", Value: p.Content},
		{Key: "created_at", Value: p.CreatedAt},
		{Key: "updated_at", Value: p.UpdatedAt},
	})
}

func (p *Post) UnmarshalBSON(data []byte) error {
	var tmp mongoPost
	if err := bson.Unmarshal(data, &tmp); err != nil {
		return err
	}

	p.ID = tmp.ID.Hex()
	p.Title = tmp.Title
	p.Content = tmp.Content
	p.CreatedAt = tmp.CreatedAt
	p.UpdatedAt = tmp.UpdatedAt

	return nil
}
