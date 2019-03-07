package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/godiscourse/godiscourse/api/durable"
	"github.com/godiscourse/godiscourse/api/session"
	"github.com/gofrs/uuid"
)

const (
	minCommentBodySize = 6
)

const commentsDDL = `
CREATE TABLE IF NOT EXISTS comments (
	comment_id            VARCHAR(36) PRIMARY KEY,
	body                  TEXT NOT NULL,
  topic_id              VARCHAR(36) NOT NULL REFERENCES topics ON DELETE CASCADE,
	user_id               VARCHAR(36) NOT NULL REFERENCES users ON DELETE CASCADE,
	score                 INTEGER NOT NULL DEFAULT 0,
	created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX ON comments (topic_id, created_at);
CREATE INDEX ON comments (user_id, created_at);
CREATE INDEX ON comments (score DESC, created_at);
`

// Comment is struct for comment of topic
type Comment struct {
	CommentID string    `sql:"comment_id,pk"`
	Body      string    `sql:"body"`
	TopicID   string    `sql:"topic_id"`
	UserID    string    `sql:"user_id"`
	Score     int       `sql:"score,notnull"`
	CreatedAt time.Time `sql:"created_at"`
	UpdatedAt time.Time `sql:"updated_at"`

	User *User
}

var commentCols = []string{"comment_id", "body", "topic_id", "user_id", "score", "created_at", "updated_at"}

func (c *Comment) values() []interface{} {
	return []interface{}{c.CommentID, c.Body, c.TopicID, c.UserID, c.Score, c.CreatedAt, c.UpdatedAt}
}

// CreateComment create a new comment
func (user *User) CreateComment(ctx context.Context, topicID, body string) (*Comment, error) {
	body = strings.TrimSpace(body)
	if len(body) < minCommentBodySize {
		return nil, session.BadDataError(ctx)
	}
	t := time.Now()
	c := &Comment{
		CommentID: uuid.Must(uuid.NewV4()).String(),
		Body:      body,
		UserID:    user.UserID,
		CreatedAt: t,
		UpdatedAt: t,
	}
	err := runInTransaction(ctx, func(tx *sql.Tx) error {
		topic, err := findTopic(ctx, tx, topicID)
		if err != nil {
			return err
		} else if topic == nil {
			return session.NotFoundError(ctx)
		}
		count, err := commentsCountByTopic(ctx, tx, topicID)
		if err != nil {
			return err
		}
		topic.CommentsCount = count + 1
		topic.UpdatedAt = t
		c.TopicID = topic.TopicID
		cols, params := prepareColumnsWithValues(commentCols)
		_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO comments (%s) VALUES (%s)", cols, params), c.values()...)
		if err != nil {
			return err
		}
		tcols, tparams := prepareColumnsWithValues([]string{"comments_count", "updated_at"})
		_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE topics SET (%s)=(%s) WHERE topic_id='%s'", tcols, tparams, topic.TopicID), topic.CommentsCount, topic.UpdatedAt)
		if err != nil {
			return err
		}
		_, err = upsertStatistic(ctx, tx, "comments")
		return err
	})
	if err != nil {
		if _, ok := err.(session.Error); ok {
			return nil, err
		}
		return nil, session.TransactionError(ctx, err)
	}
	c.User = user
	return c, nil
}

// UpdateComment update the comment by id
func (user *User) UpdateComment(ctx context.Context, id, body string) (*Comment, error) {
	body = strings.TrimSpace(body)
	if len(body) < minCommentBodySize {
		return nil, session.BadDataError(ctx)
	}
	var comment *Comment
	err := runInTransaction(ctx, func(tx *sql.Tx) error {
		var err error
		comment, err = findComment(ctx, tx, id)
		if err != nil {
			return err
		} else if comment == nil {
			return session.NotFoundError(ctx)
		} else if comment.UserID != user.UserID && !user.isAdmin() {
			return session.AuthorizationError(ctx)
		}
		comment.Body = body
		comment.UpdatedAt = time.Now()
		cols, params := prepareColumnsWithValues([]string{"body", "updated_at"})
		_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE comments SET (%s)=(%s) WHERE comment_id='%s'", cols, params, comment.CommentID), comment.Body, comment.UpdatedAt)
		return err
	})
	if err != nil {
		if _, ok := err.(session.Error); ok {
			return nil, err
		}
		return nil, session.TransactionError(ctx, err)
	}
	return comment, nil
}

// ReadComments read comments by topicID, parameters: offset
func (topic *Topic) ReadComments(ctx context.Context, offset time.Time) ([]*Comment, error) {
	if offset.IsZero() {
		offset = time.Now()
	}
	rows, err := session.Database(ctx).QueryContext(ctx, fmt.Sprintf("SELECT %s FROM comments WHERE topic_id=$1 AND created_at<$2 ORDER BY created_at DESC LIMIT $3", strings.Join(commentCols, ",")), topic.TopicID, offset, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c, err := commentFromRows(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return comments, nil
}

// ReadComments read comments by userID, parameters: offset
func (user *User) ReadComments(ctx context.Context, offset time.Time) ([]*Comment, error) {
	if offset.IsZero() {
		offset = time.Now()
	}
	rows, err := session.Database(ctx).QueryContext(ctx, fmt.Sprintf("SELECT %s FROM comments WHERE user_id=$1 AND created_at<$2 ORDER BY created_at DESC LIMIT $3", strings.Join(commentCols, ",")), user.UserID, offset, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c, err := commentFromRows(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return comments, nil
}

//DeleteComment delete a comment by ID
func (user *User) DeleteComment(ctx context.Context, id string) error {
	err := runInTransaction(ctx, func(tx *sql.Tx) error {
		comment, err := findComment(ctx, tx, id)
		if err != nil || comment == nil {
			return err
		}
		if !user.isAdmin() && user.UserID != comment.UserID {
			return session.ForbiddenError(ctx)
		}
		topic, err := findTopic(ctx, tx, comment.TopicID)
		if err != nil {
			return err
		} else if topic == nil {
			return session.BadDataError(ctx)
		}
		count, err := commentsCountByTopic(ctx, tx, comment.TopicID)
		if err != nil {
			return err
		}
		topic.CommentsCount = count - 1
		topic.UpdatedAt = time.Now()
		cols, params := prepareColumnsWithValues([]string{"comments_count", "updated_at"})
		_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE topics SET (%s)=(%s) WHERE topic_id='%s'", cols, params, topic.TopicID), topic.CommentsCount, topic.UpdatedAt)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE comment_id=$1", comment.CommentID)
		return err
	})
	if err != nil {
		if _, ok := err.(session.Error); ok {
			return err
		}
		return session.TransactionError(ctx, err)
	}
	return nil
}

func findComment(ctx context.Context, tx *sql.Tx, id string) (*Comment, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, nil
	}

	rows, err := tx.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM comments WHERE comment_id=$1", strings.Join(commentCols, ",")), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return commentFromRows(rows)
}

func commentsCountByTopic(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	var count int64
	err := tx.QueryRowContext(ctx, "SELECT count(*) FROM comments WHERE topic_id=$1", id).Scan(&count)
	return count, err
}

func commentsCount(ctx context.Context, tx *sql.Tx) (int64, error) {
	var count int64
	err := tx.QueryRowContext(ctx, "SELECT count(*) FROM comments").Scan(&count)
	return count, err
}

func commentFromRows(row durable.Row) (*Comment, error) {
	var c Comment
	err := row.Scan(&c.CommentID, &c.Body, &c.TopicID, &c.UserID, &c.Score, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
