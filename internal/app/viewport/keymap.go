package viewport

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
)

type viewportKeyMap = viewport.KeyMap

type KeyMap struct {
	Esc   key.Binding
	Enter key.Binding
	Slash key.Binding
	viewportKeyMap
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close filter box"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "filter apex log"),
		),
		Slash: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "open filter box"),
		),
		viewportKeyMap: viewport.DefaultKeyMap(),
	}
}

var keys = DefaultKeyMap()
