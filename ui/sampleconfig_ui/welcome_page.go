package sampleconfig_ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/wordwrap"
)

const (
	headerHeight = 3
	footerHeight = 3
)

type WelcomePage struct {
	content string
	keys    welcomeKeyMap

	viewport viewport.Model
	help     help.Model
}

type welcomeKeyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

func (k welcomeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit}
}
func (k welcomeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Quit}, // first column
	}
}

func NewWelcomePage() WelcomePage {

	keys := welcomeKeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎/enter", "start"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	welcome := "Welcome! Here you can create a Telegraf Config.\n"

	blinkingStyle := lipgloss.NewStyle().Background(lipgloss.Color("#22ADF6")).Blink(true)
	blinking := fmt.Sprintf("Press %s to get started", blinkingStyle.Render("ENTER"))

	description := `
On the following page, you will be able to select plugins to generate a new config.
When ready you can press S to generate the new config from the selected plugins.`

	content := fmt.Sprintf("%s\n%s\n%s", welcome, blinking, description)

	return WelcomePage{content: content, keys: keys, help: help.NewModel()}
}

func (w *WelcomePage) Init(width int, height int) {
	w.help.Width = width
	fullView := w.help.FullHelpView(w.keys.FullHelp())
	verticalMargins := headerHeight + footerHeight + strings.Count(fullView, "\n") + 1
	w.viewport = viewport.Model{Width: width, Height: height - verticalMargins}
	w.viewport.SetContent(wordwrap.String(w.content, width))
}

func (w *WelcomePage) Update(m tea.Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// These keys should exit the program.
		case key.Matches(msg, w.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, w.keys.Enter):
			currentPage = pluginSelection
		}
	case tea.WindowSizeMsg:
		w.help.Width = msg.Width
		fullView := w.help.FullHelpView(w.keys.FullHelp())
		verticalMargins := headerHeight + footerHeight + strings.Count(fullView, "\n") + 1

		w.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
		w.viewport.SetContent(wordwrap.String(w.content, msg.Width))
	}

	// Because we're using the viewport's default update function (with pager-
	// style navigation) it's important that the viewport's update function:
	//
	// * Receives messages from the Bubble Tea runtime
	// * Returns commands to the Bubble Tea runtime
	//
	w.viewport, _ = w.viewport.Update(msg)

	return m, tea.Batch(cmds...)
}

func (w *WelcomePage) View() string {
	headerTop := "╭─────────────────────╮"
	headerMid := "│ Sample Config UI 🐯 ├"
	headerBot := "╰─────────────────────╯"
	headerMid += strings.Repeat("─", max(0, w.viewport.Width-runewidth.StringWidth(headerMid)))
	header := fmt.Sprintf("%s\n%s\n%s", headerTop, headerMid, headerBot)

	var footer string

	footerTop := strings.Repeat(" ", max(0, w.viewport.Width))
	footerMid := strings.Repeat("─", max(0, w.viewport.Width))
	footerBot := strings.Repeat(" ", max(0, w.viewport.Width))
	footer = fmt.Sprintf("%s\n%s\n%s", footerTop, footerMid, footerBot)

	helpView := w.help.View(w.keys)

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, w.viewport.View(), footer, helpView)
}
