package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	//"image"
	//"archive/zip"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	world "github.com/BigCitrusFruit/StructureChest/world"
)

type AppUI struct {
	Tabs *container.AppTabs
	StructureTab struct {
		Split *container.Split
		StructureSearch *widget.Entry
		StructureTree *widget.Tree
		WorldSearch *widget.Entry
		SortSelect *widget.Select
		WorldList *widget.List
		WorldSort *widget.Select
	}
}

type sortMethod int

type treeControl struct {
	Directory string
	RootNodes []*structNode
	SearchQuery string
	SortMethod sortMethod
	NodeMap map[string]*structNode
	Increment atomic.Uint64
}

type worldsControl struct {
	SelectedFolder string
	Worlds []*worldCard
	SearchQuery string
	SortMethod sortMethod
}

type structNode struct {
	Path string
	Name string
	IsDir bool
	ModTime time.Time
	Children []*structNode
	Uid string
}

type worldCard struct {
	PathToFolder string
	WorldName string
	LastPlayed string
	Structures []world.StructureName
	Packs []string
	Card *CardItem
}

type CardItem struct {
	widget.BaseWidget
	Image *canvas.Image
	Name *widget.Label
	Date *widget.Label
	StructuresLabel *widget.Label
	StructuresList *widget.Label
	PacksLabel *widget.Label
	PacksList *widget.Label
	Root *fyne.Container
}

const (
	DateA sortMethod = iota
	DateD
	Alphabetical
)

var (
	UI AppUI
	StructTree treeControl
	WorldList worldsControl
)

func main() {
	StructTree.Directory = "./files/structures"
	WorldList.SelectedFolder = "~/.var/app/io.mrarm.mcpelauncher/data/mcpelauncher/games/com.mojang"
	MainApp := app.NewWithID("com.bigcitrusfruit.structurechest")
	MainWindow := MainApp.NewWindow("StructureChestMC")

	PlaceholderLabel := widget.NewLabel("")
	BuildUI()

	headerTabs := container.NewAppTabs(
		container.NewTabItem("1 - Structures", UI.StructureTab.Split),
		container.NewTabItem("2 - Worlds", PlaceholderLabel),
		container.NewTabItem("3 - Packs", PlaceholderLabel),
	)
	WorldList.Worlds = buildWorldsList()
	UI.StructureTab.WorldList.Refresh()
	StructTree.RootNodes = buildTree(StructTree.Directory)
	StructTree.RefreshTree()

	MainWindow.SetContent(headerTabs)
	MainWindow.ShowAndRun()
}

func (tree *treeControl) RefreshTree() {
	StructTree.NodeMap = make(map[string]*structNode)
	for _, node := range StructTree.RootNodes {
		addToMap(node)
	}
	UI.StructureTab.StructureTree.Refresh()
}

func addToMap(node *structNode) {
	if node.IsDir {
		node.Uid = node.Path
	} else {
		node.Uid = node.Path + "#" + strconv.Itoa(int(StructTree.Increment.Add(1)))
	}
	StructTree.NodeMap[node.Uid] = node
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

func buildWorldsList() []*worldCard {
	path := resolvePath(filepath.Join(WorldList.SelectedFolder, "/minecraftWorlds"))
	entries, err := os.ReadDir(path) 
	if err != nil { 
		return []*worldCard{} 
	}
	var worlds []*worldCard
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worldPath := filepath.Join(path, entry.Name())
		worldName, err := os.ReadFile(filepath.Join(worldPath, "/levelname.txt"))
		if err != nil {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		structures := world.ListWorldStructures(worldPath)
		world := &worldCard{
			PathToFolder: worldPath,
			WorldName: string(worldName),
			LastPlayed: info.ModTime().Format(time.DateTime),
			Card: NewWorldCard(),
			Structures: structures,
		}
		worlds = append(worlds, world)
	}
	return worlds
}

func filterWorlds(worlds []*worldCard, query string) []*worldCard {
	return worlds
}

func resolvePath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = filepath.Join(home, path[1:])
	}
	return os.ExpandEnv(path)
}
