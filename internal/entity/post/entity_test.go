package post

import (
	"news-svc/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestValidate(t *testing.T) {
	p := &Post{}
	err := p.Validate()
	assert.ErrorIs(t, err, config.ErrEmptyTitle)

	p.Title = "Title"
	err = p.Validate()
	assert.ErrorIs(t, err, config.ErrEmptyContent)

	p.Content = "Content"
	err = p.Validate()
	assert.NoError(t, err)
}

func TestMarshalUnmarshalBSON(t *testing.T) {
	orig := &Post{
		ID:        "",
		Title:     "Hello",
		Content:   "World",
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Millisecond),
	}

	data, err := orig.MarshalBSON()
	assert.NoError(t, err)

	var doc bson.D
	err = bson.Unmarshal(data, &doc)
	assert.NoError(t, err)
	for _, elem := range doc {
		assert.NotEqual(t, "_id", elem.Key)
	}

	hexID := bson.NewObjectID().Hex()
	orig.ID = hexID
	dataWithID, err := orig.MarshalBSON()
	assert.NoError(t, err)

	var round Post
	err = round.UnmarshalBSON(dataWithID)
	assert.NoError(t, err)

	assert.Equal(t, orig.ID, round.ID)
	assert.Equal(t, orig.Title, round.Title)
	assert.Equal(t, orig.Content, round.Content)
	assert.WithinDuration(t, orig.CreatedAt, round.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, orig.UpdatedAt, round.UpdatedAt, time.Millisecond)
}

func TestMarshalBSONInvalidHex(t *testing.T) {
	p := &Post{ID: "invalid-hex", Title: "T", Content: "C", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_, err := p.MarshalBSON()
	assert.Error(t, err)
}
