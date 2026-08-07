package main

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"

	//"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func BuildUI()  {
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
				children = append(children, folder.Uid)
			}
			for _, file := range files {
				children = append(children, file.Uid)
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
	UI.StructureTab.WorldSearch = widget.NewEntry()
	UI.StructureTab.WorldSearch.SetPlaceHolder("Search worlds...")
	UI.StructureTab.WorldSearch.ActionItem = widget.NewButton("X", func() {
		UI.StructureTab.WorldSearch.SetText("")
	})
	UI.StructureTab.WorldSort = widget.NewSelect([]string{"Newest First", "Oldest First", "Alphabetical"}, func(option string) {})
	UI.StructureTab.WorldList = widget.NewList(
		func() int { // length
			return len(WorldList.Worlds)
		},
		func() fyne.CanvasObject { // template
			return NewWorldCard()
		},
		func(id widget.ListItemID, object fyne.CanvasObject) { // update
			world := WorldList.Worlds[id]
			item := object.(*CardItem)
			world.Card = item
			world.setContent()
		},
	)
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
			container.NewVBox(
				container.NewBorder(
					widget.NewLabel("placeholder"),
					nil,
					container.NewHBox(
						widget.NewButton("Import world...", func() {}),
						widget.NewLabel("Sort by: "),
						UI.StructureTab.WorldSort,
					),
					nil,
					UI.StructureTab.WorldSearch,
				),
			),
			nil, nil, nil, 
			container.NewScroll(UI.StructureTab.WorldList),
		),
	)
	UI.StructureTab.Split.Offset = 0.3
}

func NewWorldCard() *CardItem {
	card := &CardItem{
		Image: canvas.NewImageFromResource(theme.FolderIcon()),
		Name: widget.NewLabelWithStyle("Template", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		Date: widget.NewLabel(""),
		PacksLabel: widget.NewLabel(""),
		PacksList: widget.NewLabel(""),
		StructuresLabel: widget.NewLabel(""),
		StructuresList: widget.NewLabel(""),
	}
	card.Image.FillMode = canvas.ImageFillContain
	card.Image.CornerRadius = 6
	card.Image.SetMinSize(fyne.NewSize(64, 64))

	details := container.NewVBox(
		container.NewHBox(card.Name, widget.NewSeparator(), card.Date),
		container.NewBorder(
			nil, nil,
			card.PacksLabel,
			nil,
			card.PacksList,
		),
		container.NewBorder(
			nil, nil,
			card.StructuresLabel,
			nil,
			card.StructuresList,
		),
	)
	leftSide := container.NewStack(card.Image)
	cardContent := container.NewBorder(
		nil, nil,
		leftSide,
		nil,
		details,
	)
	card.Root = container.NewVBox(widget.NewCard("","", cardContent))
	card.ExtendBaseWidget(card)
	return card
}

func (card *CardItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(card.Root)
}

func (card *worldCard) setContent() {
	card.Card.Name.SetText(card.WorldName)
	card.Card.Date.SetText(card.LastPlayed)
	var structList []string
	if len(card.Structures) > 5 {
		for i := range 5 {
			structList = append(structList, card.Structures[i].Name)
		}
		structList = append(structList, fmt.Sprintf("and %d more.", len(card.Structures) - 5))
	} else {
		for _, structure := range card.Structures {
			structList = append(structList, structure.Name)
		}
	}
	card.Card.StructuresLabel.SetText(fmt.Sprintf("Structures (%d): ", len(card.Structures)))
	card.Card.StructuresList.SetText(strings.Join(structList, ", "))
	card.Card.PacksLabel.SetText(fmt.Sprintf("Packs (%d)", len(card.Packs)))
}

