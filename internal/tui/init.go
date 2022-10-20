package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"go.etcd.io/bbolt"
)

// Init intializes the UI.
func (m Model) Init() tea.Cmd {
	db, err := bbolt.Open(m.file, 0400, nil)
	if err != nil {
		return tea.Quit
	}
	m.db = db

	err = m.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucketName []byte, bucketPtr *bbolt.Bucket) error {
			b := bucket{
				name:   string(bucketName),
				bucket: bucketPtr,
			}
			m.buckets = append(m.buckets, b)
			return nil
		})
	})
	if err != nil {
		return tea.Quit
	}

	return nil
}
