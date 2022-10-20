package tui

import tea "github.com/charmbracelet/bubbletea"

// type sessionState int

// const (
// 	idleState sessionState = iota
// 	showCodeState
// 	showImageState
// 	showMarkdownState
// 	showPdfState
// )

// Bubble represents the properties of the UI.
type Bubble struct {
	// state sessionState
	keys KeyMap
}

func New(file string) Bubble {
	return Bubble{
		keys: DefaultKeyMap(),
	}
}

// Init intializes the UI.
func (b Bubble) Init() tea.Cmd {
	return nil
}

// Update handles all UI interactions.
func (b Bubble) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

// View returns a string representation of the UI.
func (b Bubble) View() string {
	return ""
}
