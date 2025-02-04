package postgres

import (
	"fmt"
	"posts/graph/model"
	"posts/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const StorageTypePostgres = "POSTGRES"

type PostgreSqlStorage struct {
	db *gorm.DB
}

func (p *Post) ToGraph() *model.Post {
	return &model.Post{
		ID:            fmt.Sprint(p.ID),
		Title:         p.Title,
		Body:          p.Body,
		AllowComments: p.AllowComments,
		Author:        p.Author,
	}
}

func (p *Comment) ToGraph() *model.Comment {
	return &model.Comment{
		ID:       fmt.Sprint(p.ID),
		PostID:   p.PostID,
		ParentID: p.ParentID,
		Body:     p.Body,
		Author:   p.Author,
	}
}

func NewPostgreStorage(host, user, password, dbname string, port string) (*PostgreSqlStorage, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Post{}, &Comment{})

	return &PostgreSqlStorage{db: db}, nil
}

func (s *PostgreSqlStorage) CreatePost(post *model.NewPost) (*model.Post, error) {
	newPost := &Post{
		Title:         post.Title,
		Body:          post.Body,
		Author:        post.Author,
		AllowComments: post.AllowComments,
	}
	if err := s.db.Create(newPost).Error; err != nil {
		return nil, err
	}
	return newPost.ToGraph(), nil
}

func (s *PostgreSqlStorage) CreateComment(comment *model.NewComment) (*model.Comment, error) {
	newComment := &Comment{
		Body:     comment.Body,
		Author:   comment.Author,
		PostID:   comment.PostID,
		ParentID: comment.ParentID,
	}

	if err := s.db.Create(newComment).Error; err != nil {
		return nil, err
	}

	if err := s.db.Save(newComment).Error; err != nil {
		return nil, err
	}

	return newComment.ToGraph(), nil
}

func (s *PostgreSqlStorage) Posts(page *int32, limit int32) ([]*model.Post, error) {
	var posts []*Post
	offset := 0
	if page != nil {
		offset = int((*page - 1) * limit)
	}

	if err := s.db.Limit(int(limit)).Offset(offset).Find(&posts).Error; err != nil {
		return nil, err
	}

	var graphPosts []*model.Post
	for _, post := range posts {
		graphPosts = append(graphPosts, post.ToGraph())
	}
	return graphPosts, nil
}

func (s *PostgreSqlStorage) Post(id string) (*model.Post, error) {
	var post Post
	if err := s.db.First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return post.ToGraph(), nil
}

func (s *PostgreSqlStorage) Comment(id string) (*model.Comment, error) {
	var comment Comment
	if err := s.db.First(&comment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return comment.ToGraph(), nil
}

func (s *PostgreSqlStorage) Comments(postID string, first int32, after *string) (*model.CommentsConnection, error) {
	var replies []*Comment
	query := s.db.Where("parent_id IS NULL AND post_id = ?", postID).
		Order("id ASC").
		Limit(int(first))

	if after != nil {
		query = query.Where("id > ?", *after)
	}

	if err := query.Find(&replies).Error; err != nil {
		return nil, err
	}

	edges := []*model.CommentsEdge{}
	for _, comment := range replies {
		edges = append(edges, &model.CommentsEdge{
			Cursor: fmt.Sprint(comment.ID),
			Node:   comment.ToGraph(),
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: new(bool),
	}
	if len(edges) > 0 {
		pageInfo.EndCursor = edges[len(edges)-1].Cursor
	}
	*pageInfo.HasNextPage = len(edges) == int(first)

	return &model.CommentsConnection{Edges: edges, PageInfo: pageInfo}, nil
}

func (s *PostgreSqlStorage) Replies(parentID string, first int32, after *string) (*model.CommentsConnection, error) {
	var replies []*Comment
	query := s.db.Where("parent_id = ?", parentID).
		Order("id ASC").
		Limit(int(first))

	if after != nil {
		query = query.Where("id > ?", *after)
	}

	if err := query.Find(&replies).Error; err != nil {
		return nil, err
	}

	graphReplies := []*model.CommentsEdge{}
	for _, reply := range replies {
		graphReplies = append(graphReplies, &model.CommentsEdge{
			Cursor: fmt.Sprint(reply.ID),
			Node:   reply.ToGraph(),
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: new(bool),
	}
	if len(graphReplies) > 0 {
		pageInfo.EndCursor = graphReplies[len(graphReplies)-1].Cursor
	}
	*pageInfo.HasNextPage = len(graphReplies) == int(first)

	return &model.CommentsConnection{Edges: graphReplies, PageInfo: pageInfo}, nil
}

var _ storage.Storage = (*PostgreSqlStorage)(nil)
