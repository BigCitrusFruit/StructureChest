package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	fyne "fyne.io/fyne/v2"
	app "fyne.io/fyne/v2/app"
	container "fyne.io/fyne/v2/container"
	widget "fyne.io/fyne/v2/widget"
)

type structureNode struct {
	Name string
	Path string
	IsDir bool
	ModTime time.Time
	Children []*structureNode
}

type repositoryState struct {
	RootPath string
	Nodes []*structureNode
	SearchQuery string
	SortMode sortOption
	Container *fyne.Container
}

type sortOption int

const (
	SortByDateAsc sortOption = iota
	SortByDateDesc
	SortByNameAsc
	SortByNameDesc
)

func Run() {
	main := app.NewWithID("com.bigcitrusfruit.structurechest")
	mainWindow := main.NewWindow("StructureChestMC")

	const StructureRoot = "./structures"

	// Left panel
	structureList := container.NewStack()
	state := repositoryState{
		RootPath: StructureRoot,
		Nodes: buildStructureTree(StructureRoot),
		SortMode: SortByDateAsc,
		Container: structureList,
	}

	structureSearch := widget.NewEntry()
	structureSearch.SetPlaceHolder("Search structures...")
	structureSearch.OnChanged = func(text string) {
		state.SearchQuery = text
		state.Refresh()
	}
	sortSelect := widget.NewSelect([]string{"Date ↑", "Date ↓", "Name ↑", "Name ↓"}, func(selected string) {
		switch selected {
		case "Date ↑":
			state.SortMode = SortByDateAsc
		case "Name ↑":
			state.SortMode = SortByNameAsc
		case "Name ↓":
			state.SortMode = SortByNameDesc
		default:
			state.SortMode = SortByDateDesc
		}
		state.Refresh()
	})
	sortSelect.SetSelected("Date ↑")
	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Structures", fyne.TextAlignLeading, fyne.TextStyle{Bold: true,}),
			sortSelect,
			structureSearch,
		),
		widget.NewButton("Export selected as pack", func() {}),
		nil, nil,
		container.NewVScroll(state.Container),
	)

	// Right panel
	worldSearch := widget.NewEntry()
	worldSearch.SetPlaceHolder("Search worlds...")
	worldList := widget.NewLabel("Bow Wow\nGrrrrr\n")

	rightPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Worlds", fyne.TextAlignLeading, fyne.TextStyle{Bold: true,}),
			worldSearch,
		),
		widget.NewButton("Upload", func() {}),
		nil, nil,
		container.NewVScroll(worldList),
	)

	mainWindow.Resize(fyne.NewSize(900,600))

	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.2)

	mainWindow.SetContent(split)

	
	mainWindow.ShowAndRun()
}

func buildStructureTree(dir string) []*structureNode {
	if err := os.MkdirAll(dir, 0755); err != nil { 
		return nil
	}
	nodes, err := scanDirectory(dir)
	if err != nil {
		return nil
	}
	return nodes
}

func scanDirectory(dir string) ([]*structureNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var nodes []*structureNode

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		node := &structureNode{
			Name: entry.Name(),
			Path: fullPath,
			IsDir: entry.IsDir(),
			ModTime: info.ModTime(),
		}
		if node.IsDir {
			children, err := scanDirectory(fullPath)
			if err == nil {
				node.Children = children
			}
			nodes = append(nodes, node)
		} else if strings.ToLower(filepath.Ext(fullPath)) == ".mcstructure" {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func sortNodes(nodes []*structureNode, mode sortOption) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		switch mode {
		case SortByNameAsc:
			return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
		case SortByNameDesc:
			return strings.ToLower(nodes[i].Name) > strings.ToLower(nodes[j].Name)
		case SortByDateAsc:
			return nodes[i].ModTime.Before(nodes[j].ModTime)
		case SortByDateDesc:
			fallthrough
		default:
			return nodes[i].ModTime.After(nodes[j].ModTime)
		}
	})
	for _, child := range nodes {
		if child.IsDir && len(child.Children) > 0 {
			sortNodes(child.Children, mode)
		}
	}
}

func filterNodes(nodes []*structureNode, query string) []*structureNode {
	if strings.TrimSpace(query) == "" {
		return nodes
	}
	var result []*structureNode
	query = strings.ToLower(query)

	for _, node := range nodes {
		if node.IsDir {
			matchingChildren := filterNodes(node.Children, query)
			if len(matchingChildren) > 0 {
				dirCopy := *node
				dirCopy.Children = matchingChildren
				result = append(result, &dirCopy)
			}
		} else if strings.Contains(strings.ToLower(node.Name), query) {
			result = append(result, node)
		}
	}
	return result
}

func (state *repositoryState) Refresh() {
	filtered := filterNodes(state.Nodes, state.SearchQuery)
	sortNodes(filtered, state.SortMode)
	newUI := renderTree(filtered)

	state.Container.Objects = []fyne.CanvasObject{newUI}
	state.Container.Refresh()
}

func renderTree(nodes []*structureNode) fyne.CanvasObject {
	if len(nodes) == 0 {
		return widget.NewLabel("No structures to show.")
	}
	accordion := widget.NewAccordion()
	var looseItems []fyne.CanvasObject

	for _, node := range nodes {
		if node.IsDir {
			subContent := renderTree(node.Children)
			item := widget.NewAccordionItem(node.Name, subContent)
			accordion.Append(item)
		} else {
			button := widget.NewButton(node.Name, func() {
				// TODO make structure selection
			})
			button.Alignment = widget.ButtonAlignLeading
			looseItems = append(looseItems, button)
		}
	}
	if len(accordion.Items) > 0 {
		if len(looseItems) > 0 {
			return container.NewVBox(accordion, container.NewVBox(looseItems...))
		}
		return accordion
	}
	return container.NewVBox(looseItems...)
}
