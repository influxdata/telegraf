package sampleconfig_ui

import (
	"fmt"
	"sort"
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

const (
	inputIndex = iota
	outputIndex
	aggregatorIndex
	processorIndex
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

	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	checked = lipgloss.NewStyle().SetString("✓").Foreground(special).PaddingRight(1).String()
)

type Item struct {
	DisplayTitle, RenderedTitle, ItemTitle string
	Desc, SampleConfig                     string
	Index                                  int
}

func (i Item) Title() string       { return i.DisplayTitle }
func (i Item) Description() string { return i.Desc }
func (i Item) FilterValue() string { return i.DisplayTitle }

type PluginPage struct {
	Tabs         []string
	activatedTab int
	PluginLists  [][]Item
	TabContent   []list.Model
	help         help.Model

	inputPlugins       map[string]Item
	outputPlugins      map[string]Item
	aggregatorsPlugins map[string]Item
	processorsPlugins  map[string]Item

	width int

	keys *pluginKeyMap
}

func (p *PluginPage) createPluginList(content []Item, width int, height int) list.Model {
	c := make([]list.Item, len(content))
	for i, arg := range content {
		c[i] = arg
	}

	pluginList := list.NewModel(c, newItemDelegate(p.keys), 0, 0)
	pluginList.SetShowStatusBar(false)
	pluginList.SetShowTitle(false)
	pluginList.KeyMap.PrevPage = key.NewBinding(
		key.WithKeys("h", "pgup"),
		key.WithHelp("h/pgup", "prev page"),
	)
	pluginList.KeyMap.NextPage = key.NewBinding(
		key.WithKeys("l", "pgdown"),
		key.WithHelp("l/pgdn", "next page"),
	)

	pluginList.SetSize(width, height-1)

	return pluginList
}

func NewPluginPage() PluginPage {
	tabs := []string{
		"Inputs",
		"Outputs",
		"Aggregators",
		"Processors",
	}

	var inputContent, outputContent, aggregatorContent, processorContent []Item
	titleColor := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})
	for name, creator := range inputs.Inputs {
		inputContent = append(inputContent, Item{
			DisplayTitle:  name,
			ItemTitle:     name,
			RenderedTitle: fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			Desc:          creator().Description(),
			SampleConfig:  creator().SampleConfig(),
		})
	}

	sort.Slice(inputContent, func(i, j int) bool {
		return strings.ToLower(inputContent[i].ItemTitle) < strings.ToLower(inputContent[j].ItemTitle)
	})

	for i := range inputContent {
		inputContent[i].Index = i
	}

	for name, creator := range outputs.Outputs {
		outputContent = append(outputContent, Item{
			DisplayTitle:  name,
			ItemTitle:     name,
			RenderedTitle: fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			Desc:          creator().Description(),
			SampleConfig:  creator().SampleConfig(),
		})
	}

	sort.Slice(outputContent, func(i, j int) bool {
		return strings.ToLower(outputContent[i].ItemTitle) < strings.ToLower(outputContent[j].ItemTitle)
	})

	for i := range outputContent {
		outputContent[i].Index = i
	}

	for name, creator := range aggregators.Aggregators {
		aggregatorContent = append(aggregatorContent, Item{
			DisplayTitle:  name,
			ItemTitle:     name,
			RenderedTitle: fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			Desc:          creator().Description(),
			SampleConfig:  creator().SampleConfig(),
		})
	}

	for i := range aggregatorContent {
		aggregatorContent[i].Index = i
	}

	sort.Slice(aggregatorContent, func(i, j int) bool {
		return strings.ToLower(aggregatorContent[i].ItemTitle) < strings.ToLower(aggregatorContent[j].ItemTitle)
	})

	for name, creator := range processors.Processors {
		processorContent = append(processorContent, Item{
			DisplayTitle:  name,
			ItemTitle:     name,
			RenderedTitle: fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			Desc:          creator().Description(),
			SampleConfig:  creator().SampleConfig(),
		})
	}

	for i := range processorContent {
		processorContent[i].Index = i
	}

	sort.Slice(processorContent, func(i, j int) bool {
		return strings.ToLower(processorContent[i].ItemTitle) < strings.ToLower(processorContent[j].ItemTitle)
	})

	var t [][]Item
	t = append(t, inputContent)
	t = append(t, outputContent)
	t = append(t, aggregatorContent)
	t = append(t, processorContent)

	c := make([]list.Model, 4)

	return PluginPage{
		Tabs:               tabs,
		PluginLists:        t,
		TabContent:         c,
		keys:               newPluginKeyMap(),
		help:               help.NewModel(),
		inputPlugins:       make(map[string]Item),
		outputPlugins:      make(map[string]Item),
		aggregatorsPlugins: make(map[string]Item),
		processorsPlugins:  make(map[string]Item),
	}
}

func (p *PluginPage) Init(width int, height int) {
	verticalMargins := strings.Count(p.renderTabs(width), "\n")
	for i, l := range p.PluginLists {
		p.TabContent[i] = p.createPluginList(l, width, height-verticalMargins)
	}
}

func (p *PluginPage) Update(m tea.Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
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
		case key.Matches(msg, p.keys.Enter):
			i := p.TabContent[p.activatedTab].SelectedItem()

			// Change the selected state of the plugin
			if plugin, ok := i.(Item); ok {
				if strings.HasPrefix(plugin.DisplayTitle, checked) {
					plugin.DisplayTitle = plugin.ItemTitle
					switch p.activatedTab {
					case inputIndex:
						delete(p.inputPlugins, plugin.ItemTitle)
					case outputIndex:
						delete(p.outputPlugins, plugin.ItemTitle)
					case aggregatorIndex:
						delete(p.aggregatorsPlugins, plugin.ItemTitle)
					case processorIndex:
						delete(p.processorsPlugins, plugin.ItemTitle)
					}
				} else {
					// Add a checkmark next to the title
					plugin.DisplayTitle = plugin.RenderedTitle
					switch p.activatedTab {
					case inputIndex:
						p.inputPlugins[plugin.ItemTitle] = plugin
					case outputIndex:
						p.outputPlugins[plugin.ItemTitle] = plugin
					case aggregatorIndex:
						p.aggregatorsPlugins[plugin.ItemTitle] = plugin
					case processorIndex:
						p.processorsPlugins[plugin.ItemTitle] = plugin
					}
				}
				// Update the items title
				p.TabContent[p.activatedTab].SetItem(plugin.Index, plugin)

				// If filtering, exit filter state and jump to the selected plugin
				if p.TabContent[p.activatedTab].SettingFilter() {
					p.TabContent[p.activatedTab].ResetFilter()
					p.TabContent[p.activatedTab].Select(plugin.Index)
				}
			}
		case key.Matches(msg, p.keys.Save):

		}
	case tea.WindowSizeMsg:
		p.help.Width = msg.Width

		// Since this program is using the full size of the viewport we need
		// to wait until we've received the window dimensions before we
		// can initialize the viewport. The initial dimensions come in
		// quickly, though asynchronously, which is why we wait for them
		// here.
		p.width = msg.Width
		verticalMargins := strings.Count(p.renderTabs(msg.Width), "\n")
		p.TabContent[p.activatedTab].SetSize(msg.Width, msg.Height-verticalMargins-1)
	}

	var cmd tea.Cmd
	p.TabContent[p.activatedTab], cmd = p.TabContent[p.activatedTab].Update(msg)
	return m, cmd
}

// renderTabs will create the view for the tabs
// counting the new lines can help determine the height for other components
func (p *PluginPage) renderTabs(width int) string {
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
	row := p.renderTabs(p.width)
	_, err := doc.WriteString(row)
	if err != nil {
		return err.Error()
	}

	// List of plugins
	_, err = doc.WriteString(p.TabContent[p.activatedTab].View())
	if err != nil {
		return err.Error()
	}

	return docStyle.Render(doc.String())
}
