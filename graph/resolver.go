package graph

import (
	"posts/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const maxCommentLength = 2000

type Resolver struct {
	storage storage.Storage
}

func NewResolver(storage storage.Storage) *Resolver {
	return &Resolver{storage}
}
