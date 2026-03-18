package domain

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{},       // first column
		{k.Quit}, // second column
	}
}
