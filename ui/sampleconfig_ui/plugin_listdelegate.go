package sampleconfig_ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

// newItemDelegate is used to provide the additional custom key bindings
// this allows the custom keys to be shown in the same help model
func newItemDelegate(keys *pluginKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.ShortHelpFunc = keys.ShortHelp
	d.FullHelpFunc = keys.FullHelp
	return d
}

// pluginKeyMap defines a set of keybindings for the plugin selection page
// These key bindings are an addition to the keys provided by the default list
type pluginKeyMap struct {
	Left      key.Binding
	Right     key.Binding
	Backspace key.Binding
	Enter     key.Binding
	Info      key.Binding
	Save      key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k pluginKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Backspace}
}

// FullHelp returns keybindings for the expanded help view
func (k pluginKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Enter, k.Info, k.Save},
	}
}

func newPluginKeyMap() *pluginKeyMap {
	return &pluginKeyMap{
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "next left tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next right tab"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎ enter", "select plugin"),
		),
		Info: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "ⓘ plugin details"),
		),
		Save: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "💾 save config"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "go back"),
		),
	}
}
