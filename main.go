package main

import (
	"os"
	"path/filepath"
	"sort"
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
		ComMojangEntry *widget.Entry
		WorldListContainer fyne.CanvasObject
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
	FilteredWorlds []*worldCard
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
	LastPlayed time.Time
	Structures []world.StructureName
	Packs []world.BehaviorPack
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
	WorldPointer *worldCard
}

type TreeItemLabel struct {
	widget.Label
	NodeID widget.TreeNodeID
	SplitOffset float64
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
	WorldList.UpdateWorlds()
	UI.StructureTab.WorldList.Refresh()
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
		_, packIDs := world.ListWorldPacks(worldPath)
		var packs []world.BehaviorPack	
		for _, packID := range packIDs {
			pack, err := world.GetBPFromWorld(worldPath, packID)
			if err != nil || pack.Name == "" {
				continue
			}
			packs = append(packs, pack)
		}
		world := &worldCard{
			PathToFolder: worldPath,
			WorldName: string(worldName),
			LastPlayed: info.ModTime(),
			Structures: structures,
			Packs: packs,
		}
		world.Card = NewWorldCard(world)
		worlds = append(worlds, world)
	}
	return worlds
}

func (worldList *worldsControl) UpdateWorlds() {
	worldList.FilteredWorlds = []*worldCard{}
	for _, world := range worldList.Worlds {
		if strings.Contains(strings.ToLower(world.WorldName), strings.ToLower(worldList.SearchQuery)) {
			worldList.FilteredWorlds = append(worldList.FilteredWorlds, world)
		}
		sort.Slice(worldList.FilteredWorlds, func(i int, j int) bool {
			switch worldList.SortMethod {
			case DateA:
				return worldList.FilteredWorlds[i].LastPlayed.After(worldList.FilteredWorlds[j].LastPlayed)
			case DateD:
				return worldList.FilteredWorlds[i].LastPlayed.Before(worldList.FilteredWorlds[j].LastPlayed)
			case Alphabetical:
				return worldList.FilteredWorlds[i].WorldName < worldList.FilteredWorlds[j].WorldName
			default:
				return worldList.FilteredWorlds[i].LastPlayed.Before(worldList.FilteredWorlds[j].LastPlayed)
			}
		})
	}
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
