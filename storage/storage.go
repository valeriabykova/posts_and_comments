package storage

import (
	"posts/graph/model"
)

type Storage interface {
	CreatePost(post *model.NewPost) (*model.Post, error)
	CreateComment(comment *model.NewComment) (*model.Comment, error)
	Posts(page *int32, limit int32) ([]*model.Post, error)
	Post(id string) (*model.Post, error)
	Comment(id string) (*model.Comment, error)
	Comments(postID string, first int32, after *string) (*model.CommentsConnection, error)
	Replies(ParentID string, first int32, after *string) (*model.CommentsConnection, error)
}
