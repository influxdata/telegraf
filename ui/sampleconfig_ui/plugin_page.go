package sampleconfig_ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}
	highlight = lipgloss.AdaptiveColor{Light: "#13002D", Dark: "#22ADF6"}
	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}
	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)
	activeTab = tab.Copy().Border(activeTabBorder, true)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	docStyle = lipgloss.NewStyle().Padding(1, 2, 0, 2)
)

type Item struct {
	ItemTitle, Desc string
}

func (i Item) Title() string       { return i.ItemTitle }
func (i Item) Description() string { return i.Desc }
func (i Item) FilterValue() string { return i.ItemTitle }

// pluginKeyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type pluginKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Help      key.Binding
	Backspace key.Binding
	Filter    key.Binding
	Enter     key.Binding
	Info      key.Binding
	Save      key.Binding
	Quit      key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k pluginKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Backspace, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k pluginKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Filter, k.Enter, k.Info, k.Save},
		{k.Help, k.Backspace, k.Quit},
	}
}

type PluginPage struct {
	Tabs         []string
	activatedTab int
	PluginLists  [][]list.Item
	TabContent   []list.Model
	help         help.Model

	width int

	keys pluginKeyMap
}

func createPluginList(content []list.Item, width int, height int) list.Model {
	pluginList := list.NewModel(content, list.NewDefaultDelegate(), width, height-1)
	pluginList.SetShowStatusBar(false)
	pluginList.SetShowTitle(false)
	pluginList.SetShowHelp(false)

	return pluginList
}

func NewPluginPage() PluginPage {
	tabs := []string{
		"Inputs",
		"Outputs",
		"Aggregators",
		"Processors",
	}

	var inputContent, outputContent, aggregatorContent, processorContent []list.Item

	for name, creator := range inputs.Inputs {
		inputContent = append(inputContent, Item{ItemTitle: name, Desc: creator().Description()})
	}

	for name, creator := range outputs.Outputs {
		outputContent = append(outputContent, Item{ItemTitle: name, Desc: creator().Description()})
	}

	for name, creator := range aggregators.Aggregators {
		aggregatorContent = append(aggregatorContent, Item{ItemTitle: name, Desc: creator().Description()})
	}

	for name, creator := range processors.Processors {
		processorContent = append(processorContent, Item{ItemTitle: name, Desc: creator().Description()})
	}

	var t [][]list.Item
	t = append(t, inputContent)
	t = append(t, outputContent)
	t = append(t, aggregatorContent)
	t = append(t, processorContent)

	c := make([]list.Model, 4)

	keys := pluginKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll list up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll list down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move to left tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move to right tab"),
		),
		Filter: key.NewBinding(
			key.WithKeys("filter"),
			key.WithHelp("/", "filter the list"),
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
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "go back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	return PluginPage{Tabs: tabs, PluginLists: t, TabContent: c, keys: keys, help: help.NewModel()}
}

func (p *PluginPage) Init(width int, height int) {
	p.help.Width = width
	fullView := p.help.FullHelpView(p.keys.FullHelp())
	verticalMargins := strings.Count(p.createTabs(width), "\n") + strings.Count(fullView, "\n") + 2
	for i, l := range p.PluginLists {
		p.TabContent[i] = createPluginList(l, width, height-verticalMargins)
	}
}

func (p *PluginPage) Update(m tea.Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// These keys should exit the program.
		case key.Matches(msg, p.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, p.keys.Right):
			if p.activatedTab < len(p.Tabs)-1 {
				p.activatedTab++
			}
			return m, nil
		case key.Matches(msg, p.keys.Left):
			if p.activatedTab > 0 {
				p.activatedTab--
			}
			return m, nil
		case key.Matches(msg, p.keys.Backspace):
			if p.TabContent[p.activatedTab].FilterState() != list.Filtering {
				currentPage = welcomePage
				return m, nil
			}
		case key.Matches(msg, p.keys.Help):
			p.help.ShowAll = !p.help.ShowAll
		}
	case tea.WindowSizeMsg:
		p.help.Width = msg.Width

		// Since this program is using the full size of the viewport we need
		// to wait until we've received the window dimensions before we
		// can initialize the viewport. The initial dimensions come in
		// quickly, though asynchronously, which is why we wait for them
		// here.
		p.width = msg.Width
		fullView := p.help.FullHelpView(p.keys.FullHelp())
		verticalMargins := strings.Count(p.createTabs(msg.Width), "\n") + strings.Count(fullView, "\n") + 2
		p.TabContent[p.activatedTab] = createPluginList(p.PluginLists[p.activatedTab], msg.Width, msg.Height-verticalMargins)
	}

	var cmd tea.Cmd
	p.TabContent[p.activatedTab], cmd = p.TabContent[p.activatedTab].Update(msg)
	return m, cmd
}

func (p *PluginPage) createTabs(width int) string {
	var renderedTabs []string

	for i, t := range p.Tabs {
		if i == p.activatedTab {
			renderedTabs = append(renderedTabs, activeTab.Render(t))
		} else {
			renderedTabs = append(renderedTabs, tab.Render(t))
		}
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedTabs...,
	)
	gap := tabGap.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(row)-2)))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap) + "\n\n"
}

func (p *PluginPage) View() string {
	doc := strings.Builder{}

	// Tabs
	{
		row := p.createTabs(p.width)
		_, err := doc.WriteString(row)
		if err != nil {
			return err.Error()
		}
	}

	//list
	_, err := doc.WriteString(p.TabContent[p.activatedTab].View())
	if err != nil {
		return err.Error()
	}

	_, err = doc.WriteString("\n\n" + p.help.View(p.keys))
	if err != nil {
		return err.Error()
	}

	return docStyle.Render(doc.String())
}
