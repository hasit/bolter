package tui

import "go.etcd.io/bbolt"

// Model represents the properties of the UI.
type Model struct {
	cursor  int
	file    string
	db      *bbolt.DB
	buckets []bucket
}

type bucket struct {
	name   string
	bucket *bbolt.Bucket
}

func New(file string) Model {
	return Model{
		file:    file,
		buckets: make([]bucket, 0),
		cursor:  0,
	}
}
