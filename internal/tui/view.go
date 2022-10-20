package tui

import "fmt"

// View returns a string representation of the UI.
func (m Model) View() string {
	s := fmt.Sprintf("Viewing boltdb file: %s\n\n", m.file)

	for i, bucket := range m.buckets {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s", cursor, bucket)
	}

	s += "\nPress q to quit.\n"
	return s
}
