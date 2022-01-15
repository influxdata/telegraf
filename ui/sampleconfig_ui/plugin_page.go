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
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

// Describes the styles for the plugins page
// Could be re-used by other pages in the future to keep a consisten look
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

// Item describes a plugin item in the bubbles list
// For an example of lists: https://github.com/charmbracelet/bubbletea/blob/master/examples/list-simple/main.go
type Item struct {
	DisplayTitle, RenderedTitle, ItemTitle string
	pluginDescriber                        telegraf.PluginDescriber
	Index                                  int
}

func (i Item) Title() string       { return i.DisplayTitle }
func (i Item) Description() string { return i.pluginDescriber.Description() }
func (i Item) FilterValue() string { return i.DisplayTitle }

type PluginPage struct {
	// Tabs for the plugin selection page, maps to the selected plugins
	Tabs         map[int]PluginTab
	activatedTab int
	TabContent   []list.Model
	PluginLists  [][]Item

	// Used to pass the current width,height between Update, view and subpages
	width  int
	height int

	// Holds possible keys for the page (additional to the ones that come with default list)
	keys *pluginKeyMap
	help help.Model

	// PluginPage has two sub-pages, info for each plugin and a final save screen
	// These pages are sub-pages of the plugin page to allow passing info to them
	// SampleconfigUI is only pass by value, so it can't pass info unless made global
	infoPage       Pages
	infoPageActive bool
	savePage       Pages
	savePageActive bool
}

// createPluginList will create a list.Model for the plugin lists
func (p *PluginPage) createPluginList(content []Item, width int, height int) list.Model {
	c := make([]list.Item, len(content))
	for i, arg := range content {
		c[i] = arg
	}

	pluginList := list.NewModel(c, newItemDelegate(p.keys), 0, 0)
	pluginList.SetShowStatusBar(false)
	pluginList.SetShowTitle(false)
	// The default allows left/right arrow keys but we are using them to navigate tabs
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

// processPlugin prepares a itemized plugin list for display
func processPlugin(items []Item) []Item {
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].ItemTitle) < strings.ToLower(items[j].ItemTitle)
	})

	for i := range items {
		items[i].Index = i
	}

	return items
}

// PluginTab keeps track of the type of plugin with the selected
type PluginTab struct {
	Name     string
	Selected map[string]Item
}

func NewPluginPage() PluginPage {
	var inputContent, outputContent, aggregatorContent, processorContent []Item
	titleColor := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})

	// Each input type has its own creator type, so have to duplicate the init code
	for name, creator := range inputs.Inputs {
		inputContent = append(inputContent, Item{
			DisplayTitle:    name,
			ItemTitle:       name,
			RenderedTitle:   fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			pluginDescriber: creator(),
		})
	}

	for name, creator := range outputs.Outputs {
		outputContent = append(outputContent, Item{
			DisplayTitle:    name,
			ItemTitle:       name,
			RenderedTitle:   fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			pluginDescriber: creator(),
		})
	}

	for name, creator := range aggregators.Aggregators {
		aggregatorContent = append(aggregatorContent, Item{
			DisplayTitle:    name,
			ItemTitle:       name,
			RenderedTitle:   fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			pluginDescriber: creator(),
		})
	}

	for name, creator := range processors.Processors {
		processorContent = append(processorContent, Item{
			DisplayTitle:    name,
			ItemTitle:       name,
			RenderedTitle:   fmt.Sprintf("%s%s", checked, titleColor.Render(name)),
			pluginDescriber: creator(),
		})
	}

	var t [][]Item
	t = append(t, processPlugin(inputContent))
	t = append(t, processPlugin(outputContent))
	t = append(t, processPlugin(aggregatorContent))
	t = append(t, processPlugin(processorContent))

	tabs := map[int]PluginTab{
		0: {
			Name:     "Input",
			Selected: make(map[string]Item),
		},
		1: {
			Name:     "Output",
			Selected: make(map[string]Item),
		},
		2: {
			Name:     "Aggregator",
			Selected: make(map[string]Item),
		},
		3: {
			Name:     "Processor",
			Selected: make(map[string]Item),
		},
	}

	return PluginPage{
		Tabs:        tabs,
		PluginLists: t,
		TabContent:  make([]list.Model, 4),
		keys:        newPluginKeyMap(),
		help:        help.NewModel(),
	}
}

func (p *PluginPage) Init(width int, height int) {
	verticalMargins := strings.Count(p.renderTabs(width), "\n")
	for i, l := range p.PluginLists {
		p.TabContent[i] = p.createPluginList(l, width, height-verticalMargins)
	}
	p.width = width
	p.height = height
}

func (p *PluginPage) InfoPageIndex() int {
	return 1
}

func (p *PluginPage) Update(m tea.Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if p.infoPageActive {
		return p.infoPage.Update(m, msg)
	}
	if p.savePageActive {
		return p.savePage.Update(m, msg)
	}

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
					delete(p.Tabs[p.activatedTab].Selected, plugin.ItemTitle)
				} else {
					// Add a checkmark next to the title
					plugin.DisplayTitle = plugin.RenderedTitle
					p.Tabs[p.activatedTab].Selected[plugin.ItemTitle] = plugin
				}
				// Update the items title
				p.TabContent[p.activatedTab].SetItem(plugin.Index, plugin)

				p.TabContent[p.activatedTab].ResetFilter()
				p.TabContent[p.activatedTab].Select(plugin.Index)
			}
		case key.Matches(msg, p.keys.Info):
			if !p.TabContent[p.activatedTab].SettingFilter() {
				i := p.TabContent[p.activatedTab].SelectedItem()

				if plugin, ok := i.(Item); ok {
					p.infoPageActive = true
					currentTab := p.Tabs[p.activatedTab]
					infoPage := NewPluginInfo(p, currentTab.Name, plugin)
					infoPage.Init(p.width, p.height)
					p.infoPage = &infoPage
				}
			}
		case key.Matches(msg, p.keys.Save):
			if !p.TabContent[p.activatedTab].SettingFilter() {
				savePage := NewSaveConfigPage(p, p.Tabs)
				savePage.Init(p.width, p.height)
				p.savePage = &savePage
				p.savePageActive = true
			}
		}
	case tea.WindowSizeMsg:
		p.help.Width = msg.Width
		p.width = msg.Width
		p.height = msg.Height
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

	// Sort the keys to make sure tabs are in the same order everytime
	// Using a map helps with organizing selected plugins with plugin type
	keys := make([]int, 0, len(p.Tabs))
	for k := range p.Tabs {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		if k == p.activatedTab {
			renderedTabs = append(renderedTabs, activeTab.Render(p.Tabs[k].Name))
		} else {
			renderedTabs = append(renderedTabs, tab.Render(p.Tabs[k].Name))
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
	if p.infoPageActive {
		return p.infoPage.View()
	}
	if p.savePageActive {
		return p.savePage.View()
	}

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
