package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type AppUI struct {
	Tabs *container.AppTabs
	StructureTab struct {
		Split *container.Split
		StructureSearch *widget.Entry
		StructureTree *widget.Tree
		WorldSearch *widget.Entry
		SortSelect *widget.Select
	}
}

type sortMethod int

type treeControl struct {
	Directory string
	RootNodes []*structNode
	SearchQuery string
	SortMethod sortMethod
	NodeMap map[string]*structNode
}

type structNode struct {
	Path string
	Name string
	IsDir bool
	ModTime time.Time
	Children []*structNode
}

const (
	DateA sortMethod = iota
	DateD
	Alphabetical
)

var (
	UI AppUI
	StructTree treeControl
)

func main() {
	StructTree.Directory = "./structures"
	MainApp := app.NewWithID("com.bigcitrusfruit.structurechest")
	MainWindow := MainApp.NewWindow("StructureChestMC")

	PlaceholderLabel := widget.NewLabel("")
	buildUI()

	headerTabs := container.NewAppTabs(
		container.NewTabItem("1 - Structures", UI.StructureTab.Split),
		container.NewTabItem("2 - Worlds", PlaceholderLabel),
		container.NewTabItem("3 - Packs", PlaceholderLabel),
	)
	StructTree.RootNodes = buildTree(StructTree.Directory)
	StructTree.RefreshTree()

	MainWindow.SetContent(headerTabs)
	MainWindow.ShowAndRun()
}

func buildUI()  {
	UI.StructureTab.StructureSearch = widget.NewEntry()
	UI.StructureTab.StructureSearch.SetPlaceHolder("Search structures...")
	UI.StructureTab.StructureSearch.OnChanged = func(search string) {
		StructTree.SearchQuery = search
		StructTree.RefreshTree()
	}
	UI.StructureTab.StructureSearch.ActionItem = widget.NewButton("X", func() {
		UI.StructureTab.StructureSearch.SetText("")
		StructTree.RefreshTree()
	})
	UI.StructureTab.StructureTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID { // get child id
			var (
				children []widget.TreeNodeID
				folders []*structNode
				files []*structNode
			)
			if id == "" {
				if StructTree.SearchQuery == "" {
					for _, node := range StructTree.RootNodes {
						if node.IsDir {
							folders = append(folders, node)
						} else {
							files = append(files, node)
						}
					}
				} else {
					folders = []*structNode{}
					files = filterNodeList(StructTree.RootNodes)
				}
			} else {
				for _, child := range StructTree.NodeMap[id].Children {
					if child.IsDir {
						folders = append(folders, child)
					} else {
						files = append(files, child)
					}
				}
			}
			sort.Slice(folders, func(i int, j int) bool {
					return folders[i].Name < folders[j].Name
			})
			sort.Slice(files, func(i int, j int) bool {
				switch StructTree.SortMethod {
				case Alphabetical:
					return files[i].Name < files[j].Name
				case DateA:
					return files[i].ModTime.After(files[j].ModTime)
				case DateD:
					fallthrough
				default:
					return files[i].ModTime.Before(files[j].ModTime)
				}
			})
			for _, folder := range folders {
				children = append(children, folder.Path)
			}
			for _, file := range files {
				children = append(children, file.Path)
			}
			return children
		},
		func(id widget.TreeNodeID) bool { // IsBranch
			if id == "" {
				return true
			}
			return StructTree.NodeMap[id].IsDir
		},
		func(isBranch bool) fyne.CanvasObject { // templates
			if isBranch {
				return widget.NewLabel("Branch template")
			}
			return widget.NewLabel("Leaf template")
		},
		func(id widget.TreeNodeID, isBranch bool, object fyne.CanvasObject) { // update
			node := StructTree.NodeMap[id]
			text := node.Name
			if isBranch {
				text = "🗀  " + text
			}
			object.(*widget.Label).SetText(text)
			object.(*widget.Label).Refresh()
		})
	UI.StructureTab.SortSelect = widget.NewSelect([]string{"Newest First", "Oldest First", "Alphabetical"}, func(option string) {
		switch option {
		case "Newest First":
			StructTree.SortMethod = DateD
		case "Oldest First":
			StructTree.SortMethod = DateA
		case "Alphabetical":
			StructTree.SortMethod = Alphabetical
		default:
			StructTree.SortMethod = DateA
		}
		StructTree.RefreshTree()
	})
	UI.StructureTab.Split = container.NewHSplit(
		container.NewBorder(
			container.NewVBox(
				container.NewHBox(
					widget.NewLabel("Sort by: "),
					UI.StructureTab.SortSelect,
					widget.NewButton("Export as...", func() {}),
				),
				UI.StructureTab.StructureSearch,
			),
			nil, nil, nil,
			container.NewScroll(
				UI.StructureTab.StructureTree,
			),
		),
		container.NewBorder(
			nil, nil, nil, nil, nil,
		),
	)
	UI.StructureTab.Split.Offset = 0.3
}

func (tree *treeControl) RefreshTree() {
	StructTree.NodeMap = make(map[string]*structNode)
	for _, node := range StructTree.RootNodes {
		addToMap(node)
	}
	UI.StructureTab.StructureTree.Refresh()
}

func addToMap(node *structNode) {
	StructTree.NodeMap[node.Path] = node
	if node.IsDir {
		for _, child := range node.Children {
			addToMap(child)
		}
	}
}

func buildTree(path string) []*structNode {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	var nodes []*structNode
	for _, entry := range entries {
		
		if !entry.IsDir() && filepath.Ext(entry.Name()) != ".mcstructure" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		newNode := &structNode{
			Name: strings.TrimSuffix(entry.Name(), ".mcstructure"),
			Path: filepath.Join(path, entry.Name()),
			IsDir: entry.IsDir(),
			ModTime: info.ModTime(),
		}
		if newNode.IsDir {
			newNode.Children = buildTree(newNode.Path)
		}
		nodes = append(nodes, newNode)
	}
	return nodes
}

func filterNodeList(nodes []*structNode) []*structNode {
	var filteredNodes []*structNode
	for _, node := range nodes {
		if node.IsDir {
			filteredNodes = append(filteredNodes, filterNodeList(node.Children)...)
		} else {
			if strings.Contains(strings.ToLower(node.Name), strings.ToLower(StructTree.SearchQuery)) {
				filteredNodes = append(filteredNodes, node)
			}
		}
	}
	return filteredNodes
}
