package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
)

type keyMap struct {
	quit         key.Binding
	enter        key.Binding
	tab          key.Binding
	help         key.Binding
	refresh      key.Binding
	showTable    bool
	showViewport bool
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.tab, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	var ks [][]key.Binding
	tk := table.DefaultKeyMap()
	if k.showTable {
		ks = append(ks, []key.Binding{
			k.enter,
			k.refresh,
			tk.LineUp,
			tk.LineDown,
			tk.PageUp,
			tk.PageDown,
			tk.HalfPageUp,
			tk.HalfPageDown,
			tk.GotoTop,
			tk.GotoBottom,
		})
	}
	if k.showViewport {
		vk := viewport.DefaultKeyMap()
		ks = append(ks, []key.Binding{
			vk.PageDown,
			vk.PageUp,
			vk.HalfPageUp,
			vk.HalfPageDown,
			vk.Down,
			vk.Up,
		})
	}
	ks = append(ks, []key.Binding{k.tab, k.help, k.quit})
	return ks
}

var keys = keyMap{
	quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open selected log"),
	),
	refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("refresh", "refresh apex logs"),
	),
	tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),
	help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}
