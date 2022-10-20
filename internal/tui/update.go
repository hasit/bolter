package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all UI interactions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.buckets)-1 {
				m.cursor++
			}
		case "enter", " ":
		}
	}
	return m, nil
}
