package sampleconfig_ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/wordwrap"
)

type SaveConfigPage struct {
	pluginPage      *PluginPage
	selectedPlugins map[int]PluginTab

	keys saveKeyMap

	viewport viewport.Model
	help     help.Model

	currentWidth  int
	currentHeight int

	content string
}

type saveKeyMap struct {
	Backspace key.Binding
	Enter     key.Binding
	Quit      key.Binding
}

func (k saveKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Backspace, k.Enter, k.Quit}
}

func (k saveKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

func NewSaveConfigPage(p *PluginPage, selectedPlugins map[int]PluginTab) SaveConfigPage {
	keys := saveKeyMap{
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "back to plugin selection"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎ enter", "write sample config"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc/ctrl+c", "quit"),
		),
	}

	return SaveConfigPage{pluginPage: p, keys: keys, selectedPlugins: selectedPlugins, help: help.NewModel()}
}

func (s *SaveConfigPage) Init(width int, height int) {
	s.help.Width = width
	s.currentWidth = width
	s.currentHeight = height

	titleStyle := lipgloss.NewStyle().Foreground(special)

	var renderedSelectedPlugins string
	for _, t := range s.selectedPlugins {
		var list string
		for _, i := range t.Selected {
			list += fmt.Sprintf("• %s\n", i.ItemTitle)
		}
		if list != "" {
			renderedSelectedPlugins += fmt.Sprintf("%s\n\n%s\n", titleStyle.Render(t.Name+":"), list)
		}
	}

	instructions := "If the listed plugins are correct, hit ENTER to save to telegraf.conf"
	s.content = fmt.Sprintf("%s\n\n%s", instructions, renderedSelectedPlugins)

	fullView := s.help.FullHelpView(s.keys.FullHelp())
	verticalMargins := headerHeight + footerHeight + strings.Count(fullView, "\n") + 1

	s.viewport = viewport.Model{Width: width, Height: height - verticalMargins}
	s.viewport.SetContent(wordwrap.String(s.content, s.currentWidth))
}

func (s *SaveConfigPage) Update(m tea.Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keys.Enter):
			var sampleConfig string

			for _, t := range s.selectedPlugins {
				for _, i := range t.Selected {
					sampleConfig += fmt.Sprintf("#%s\n[[%s.%s]]\n", i.pluginDescriber.Description(), t.Name, i.ItemTitle)
					sampleConfig += fmt.Sprintf("%s\n", i.pluginDescriber.SampleConfig())
				}
			}
			_ = os.WriteFile("telegraf.conf", []byte(sampleConfig), 0644)

			return m, tea.Quit
		case key.Matches(msg, s.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, s.keys.Backspace):
			s.pluginPage.savePageActive = false
			return m, nil
		}
	case tea.WindowSizeMsg:
		s.currentWidth = msg.Width
		s.help.Width = msg.Width
		fullView := s.help.FullHelpView(s.keys.FullHelp())
		verticalMargins := headerHeight + footerHeight + strings.Count(fullView, "\n") + 1

		s.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
		s.viewport.SetContent(wordwrap.String(s.content, msg.Width))
	}

	s.viewport, _ = s.viewport.Update(msg)

	return m, tea.Batch(cmds...)
}

func (s *SaveConfigPage) View() string {
	headerTop := "╭─────────────────────────────────────╮"
	headerMid := "│   Review Plugins before Saving 💾   ├"
	headerBot := "╰─────────────────────────────────────╯"
	headerMid += strings.Repeat("─", max(0, s.viewport.Width-runewidth.StringWidth(headerMid)))
	header := fmt.Sprintf("%s\n%s\n%s", headerTop, headerMid, headerBot)

	var footer string

	footerTop := strings.Repeat(" ", max(0, s.viewport.Width))
	footerMid := strings.Repeat("─", max(0, s.viewport.Width))
	footerBot := strings.Repeat(" ", max(0, s.viewport.Width))
	footer = fmt.Sprintf("%s\n%s\n%s", footerTop, footerMid, footerBot)

	helpView := s.help.View(s.keys)

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, s.viewport.View(), footer, helpView)
}
