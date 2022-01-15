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

type PluginInfo struct {
	pluginPage *PluginPage
	plugin     Item
	pluginType string

	keys infoKeyMap

	viewport viewport.Model
	help     help.Model

	currentWidth  int
	currentHeight int

	content string
}

type infoKeyMap struct {
	Backspace key.Binding
	Quit      key.Binding
}

func (k infoKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Backspace, k.Quit}
}
func (k infoKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Backspace, k.Quit}, // first column
	}
}

func NewPluginInfo(p *PluginPage, pluginType string, plugin Item) PluginInfo {
	keys := infoKeyMap{
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "back to plugin selection"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc/ctrl+c", "quit"),
		),
	}

	return PluginInfo{pluginPage: p, plugin: plugin, pluginType: pluginType, keys: keys, help: help.NewModel()}
}

func (i *PluginInfo) Init(width int, height int) {
	i.help.Width = width
	i.currentWidth = width
	i.currentHeight = height

	titleStyle := lipgloss.NewStyle().Foreground(special)

	title := fmt.Sprintf("%s %s", titleStyle.Render(i.pluginType+" Plugin:"), i.plugin.ItemTitle)
	desc := fmt.Sprintf("%s \n%s", titleStyle.Render("Description:"), i.plugin.pluginDescriber.Description())
	config := fmt.Sprintf("%s \n%s", titleStyle.Render("Sample Config:"), i.plugin.pluginDescriber.SampleConfig())

	i.content = fmt.Sprintf("%s\n\n%s\n\n%s", title, desc, config)

	fullView := i.help.FullHelpView(i.keys.FullHelp())
	verticalMargins := headerHeight + footerHeight + strings.Count(fullView, "\n") + 1

	i.viewport = viewport.Model{Width: width, Height: height - verticalMargins}
	i.viewport.SetContent(wordwrap.String(i.content, i.currentWidth))
}

func (i *PluginInfo) Update(m tea.Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// These keys should exit the program.
		case key.Matches(msg, i.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, i.keys.Backspace):
			i.pluginPage.infoPageActive = false
			return m, nil
		}
	case tea.WindowSizeMsg:
		i.currentWidth = msg.Width
		i.help.Width = msg.Width
		fullView := i.help.FullHelpView(i.keys.FullHelp())
		verticalMargins := headerHeight + footerHeight + strings.Count(fullView, "\n") + 1

		i.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
		i.viewport.SetContent(wordwrap.String(i.content, msg.Width))
	}

	i.viewport, _ = i.viewport.Update(msg)

	return m, tea.Batch(cmds...)
}

func (i *PluginInfo) View() string {
	headerTop := "╭────────────────────╮"
	headerMid := "│   Plugin Details   ├"
	headerBot := "╰────────────────────╯"
	headerMid += strings.Repeat("─", max(0, i.viewport.Width-runewidth.StringWidth(headerMid)))
	header := fmt.Sprintf("%s\n%s\n%s", headerTop, headerMid, headerBot)

	var footer string

	footerTop := strings.Repeat(" ", max(0, i.viewport.Width))
	footerMid := strings.Repeat("─", max(0, i.viewport.Width))
	footerBot := strings.Repeat(" ", max(0, i.viewport.Width))
	footer = fmt.Sprintf("%s\n%s\n%s", footerTop, footerMid, footerBot)

	helpView := i.help.View(i.keys)

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, i.viewport.View(), footer, helpView)
}
