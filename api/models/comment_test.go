package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/godiscourse/godiscourse/api/session"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestCommentCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer session.Database(ctx).Close()
	defer teardownTestContext(ctx)

	user := createTestUser(ctx, "im.yuqlee@gmail.com", "username", "password")
	assert.NotNil(user)
	category, _ := CreateCategory(ctx, "name", "alias", "Description", 0)
	assert.NotNil(category)
	topic, _ := user.CreateTopic(ctx, "title", "body", category.CategoryID)
	assert.NotNil(topic)

	commentCases := []struct {
		topicID string
		body    string
		valid   bool
	}{
		{topic.TopicID, "", false},
		{topic.TopicID, "      ", false},
		{uuid.Must(uuid.NewV4()).String(), "comment body", false},
		{topic.TopicID, "comment body", true},
	}

	for _, tc := range commentCases {
		t.Run(fmt.Sprintf("comment body %s", tc.body), func(t *testing.T) {
			if !tc.valid {
				comment, err := user.CreateComment(ctx, tc.topicID, tc.body)
				assert.NotNil(err)
				assert.Nil(comment)
				return
			}

			comment, err := user.CreateComment(ctx, tc.topicID, tc.body)
			assert.Nil(err)
			assert.NotNil(comment)
			assert.Equal(tc.body, comment.Body)
			new, err := readTestComment(ctx, comment.CommentID)
			assert.Nil(err)
			assert.NotNil(new)
			new, err = user.UpdateComment(ctx, uuid.Must(uuid.NewV4()).String(), "comment body")
			assert.NotNil(err)
			assert.Nil(new)
			new, err = user.UpdateComment(ctx, comment.CommentID, "    ")
			assert.NotNil(err)
			assert.Nil(new)
			new, err = user.UpdateComment(ctx, comment.CommentID, "new comment body")
			assert.Nil(err)
			assert.NotNil(new)
			assert.Equal("new comment body", new.Body)
			comments, err := topic.ReadComments(ctx, time.Time{})
			assert.Nil(err)
			assert.Len(comments, 1)
			comments, err = user.ReadComments(ctx, time.Time{})
			assert.Nil(err)
			assert.Len(comments, 1)
			topic, _ = ReadTopic(ctx, topic.TopicID)
			assert.NotNil(topic)
			assert.Equal(int64(1), topic.CommentsCount)
			err = user.DeleteComment(ctx, comment.CommentID)
			assert.Nil(err)
			topic, err = ReadTopic(ctx, topic.TopicID)
			assert.Nil(err)
			assert.NotNil(topic)
			assert.Equal(int64(0), topic.CommentsCount)
			comments, err = topic.ReadComments(ctx, time.Time{})
			assert.Nil(err)
			assert.Len(comments, 0)
			comments, err = user.ReadComments(ctx, time.Time{})
			assert.Nil(err)
			assert.Len(comments, 0)
			new, err = readTestComment(ctx, comment.CommentID)
			assert.Nil(err)
			assert.Nil(new)
		})
	}
}

func readTestComment(ctx context.Context, id string) (*Comment, error) {
	comment := &Comment{CommentID: id}
	if err := session.Database(ctx).Model(comment).Column(commentColumns...).WherePK().Select(); err == pg.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return comment, nil
}
