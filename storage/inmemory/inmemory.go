package inmemory

import (
	"fmt"
	"posts/graph/model"
	"posts/storage"
	"sync"
)

const StorageTypeInMemory = "IN_MEMORY"

const maxCommentLength = 2000

type InMemoryStorage struct {
	posts    *postStorage
	comments *commentStorage
}

type postStorage struct {
	sync.RWMutex
	counter int
	data    []model.Post
}

type commentStorage struct {
	sync.RWMutex
	counter int
	data    []model.Comment
}

func NewInMemoryStorage() (*InMemoryStorage, error) {
	return &InMemoryStorage{
		posts:    &postStorage{},
		comments: &commentStorage{},
	}, nil
}

func (s *InMemoryStorage) CreatePost(post *model.NewPost) (*model.Post, error) {
	s.posts.Lock()
	defer s.posts.Unlock()

	s.posts.counter += 1
	createdPost := model.Post{
		ID:            fmt.Sprint(s.posts.counter),
		Title:         post.Title,
		Body:          post.Body,
		AllowComments: post.AllowComments,
		Author:        post.Author,
	}

	if len(post.Author) == 0 {
		return nil, fmt.Errorf("author has to be non-empty")
	}
	if len(post.Title) == 0 {
		return nil, fmt.Errorf("title has to be non-empty")
	}
	if len(post.Body) == 0 {
		return nil, fmt.Errorf("body has to be non-empty")
	}

	s.posts.data = append(s.posts.data, createdPost)
	return &createdPost, nil
}

func (s *InMemoryStorage) CreateComment(comment *model.NewComment) (*model.Comment, error) {
	s.comments.counter += 1
	createdComment := model.Comment{
		ID:       fmt.Sprint(s.comments.counter),
		PostID:   comment.PostID,
		ParentID: comment.ParentID,
		Body:     comment.Body,
		Author:   comment.Author,
	}

	if len(comment.Author) == 0 {
		return nil, fmt.Errorf("author has to be non-empty")
	}
	if len(comment.Body) == 0 {
		return nil, fmt.Errorf("comment has to be non-empty")
	}
	if len(comment.Body) > maxCommentLength {
		return nil, fmt.Errorf("comment is too long")
	}

	post, err := s.Post(fmt.Sprint(comment.PostID))
	if err != nil {
		return nil, fmt.Errorf("error finding Post: %v", err)
	}

	if !post.AllowComments {
		return nil, fmt.Errorf("comments are not allowed for this post")
	}

	if comment.ParentID != nil {
		_, err := s.Comment(fmt.Sprint(*comment.ParentID))
		if err != nil {
			return nil, fmt.Errorf("error finding parent comment: %v", err)
		}
	}

	s.comments.data = append(s.comments.data, createdComment)
	return &createdComment, nil
}

func (s *InMemoryStorage) Posts(page *int32, limit int32) ([]*model.Post, error) {
	var pageStart int
	if page == nil {
		pageStart = 0
	} else {
		pageStart = int(limit * *page)
	}
	pageEnd := pageStart + int(limit)

	s.posts.RLock()
	defer s.posts.RUnlock()

	if pageStart > len(s.posts.data) {
		return nil, fmt.Errorf("no more posts")
	}
	if pageEnd > len(s.posts.data) {
		pageEnd = len(s.posts.data)
	}

	var paginatedPosts []*model.Post
	for _, post := range s.posts.data[pageStart:pageEnd] {
		paginatedPosts = append(paginatedPosts, &post)
	}
	return paginatedPosts, nil
}

func (s *InMemoryStorage) Post(id string) (*model.Post, error) {
	s.posts.RLock()
	defer s.posts.RUnlock()

	var intId int
	_, err := fmt.Sscan(id, &intId)
	if err != nil {
		return nil, fmt.Errorf("error getting int from id: %v", err)
	}

	post := s.posts.data[intId]

	if post.ID != id {
		panic(fmt.Sprintf("Storage inconsistent, found id %s on the %d place", post.ID, intId))
	}

	return &post, nil
}

func (s *InMemoryStorage) Comment(id string) (*model.Comment, error) {
	s.comments.RLock()
	defer s.comments.RUnlock()

	var intId int
	_, err := fmt.Sscan(id, &intId)
	if err != nil {
		return nil, fmt.Errorf("error getting int from id: %v", err)
	}

	comment := s.comments.data[intId]

	if comment.ID != id {
		panic(fmt.Sprintf("Storage inconsistent, found id %s on the %d place", comment.ID, intId))
	}

	return &comment, nil
}

func (s *InMemoryStorage) Comments(postID string, first int32, after *string) (*model.CommentsConnection, error) {
	s.comments.RLock()
	defer s.comments.RUnlock()

	var comments model.CommentsConnection
	comments.PageInfo = &model.PageInfo{
		HasNextPage: new(bool),
	}

	alreadyAfter := after == nil
	for _, comment := range s.comments.data {
		if fmt.Sprint(comment.PostID) != postID || comment.ParentID != nil {
			continue
		}
		if !alreadyAfter {
			if comment.ID == *after {
				alreadyAfter = true
			}
			continue
		}
		if len(comments.Edges) == int(first) {
			*comments.PageInfo.HasNextPage = true
			break
		}
		comments.Edges = append(comments.Edges, &model.CommentsEdge{
			Cursor: fmt.Sprint(comment.ID),
			Node:   &comment,
		})
		comments.PageInfo.EndCursor = fmt.Sprint(comment.ID)
	}

	return &comments, nil
}

func (s *InMemoryStorage) Replies(ParentID string, first int32, after *string) (*model.CommentsConnection, error) {
	s.comments.RLock()
	defer s.comments.RUnlock()

	var comments model.CommentsConnection
	comments.PageInfo = &model.PageInfo{
		HasNextPage: new(bool),
	}

	alreadyAfter := after == nil
	for _, comment := range s.comments.data {
		if comment.ParentID == nil || fmt.Sprint(*comment.ParentID) != ParentID {
			continue
		}
		if !alreadyAfter {
			if comment.ID == *after {
				alreadyAfter = true
			}
			continue
		}
		if len(comments.Edges) == int(first) {
			*comments.PageInfo.HasNextPage = true
			break
		}
		comments.Edges = append(comments.Edges, &model.CommentsEdge{
			Cursor: fmt.Sprint(comment.ID),
			Node:   &comment,
		})
		comments.PageInfo.EndCursor = fmt.Sprint(comment.ID)
	}

	return &comments, nil
}

var _ storage.Storage = (*InMemoryStorage)(nil)
