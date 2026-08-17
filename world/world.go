package world

import (
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	"strings"

	leveldb "github.com/df-mc/goleveldb/leveldb"
	dbopt "github.com/df-mc/goleveldb/leveldb/opt"
	util "github.com/df-mc/goleveldb/leveldb/util"
)

type RawStructure []byte

type StructureName struct {
	Prefix string
	Name string
}

type BehaviorPack struct {
	Name string
	Description string
	ID PackID
	Path string
	EmbeddedStructureNames []string
}

type ResourcePack struct {
	Name string
	Description string
	ID PackID
	Path string
}

type PackID struct {
	UUID string `json:"pack_id"`
	Version [3]int `json:"version"`
}

type manifestHeader struct {
	Header struct {
		Name string `json:"name"`
		UUID string `json:"uuid"`
		Version [3]int `json:"version"`
		Description string `json:"description"`
	} `json:"header"`
}

type WorldSession struct {
	WorldPath string
	StagedStructures map[StructureName]RawStructure
	DeletedStructures map[StructureName]bool
	StagedBPs []BehaviorPack
	StagedRPs []ResourcePack
	DeletedPacks map[PackID]bool
}

type Addon struct {
	BP BehaviorPack
	RP ResourcePack
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

func NewWorldSession(worldPath string) *WorldSession  {
	return &WorldSession{
		WorldPath: resolvePath(worldPath),
		StagedStructures: make(map[StructureName]RawStructure),
		DeletedStructures:  make(map[StructureName]bool),
		DeletedPacks: make(map[PackID]bool),
	}
}

func (session *WorldSession) AddStructure(name StructureName, data RawStructure) {
	delete(session.DeletedStructures, name)
	session.StagedStructures[name] = data
}

func (session *WorldSession) DiscardChanges() {
	session.StagedStructures = make(map[StructureName]RawStructure)
	session.DeletedStructures = make(map[StructureName]bool)
	session.DeletedPacks = make(map[PackID]bool)
	session.StagedBPs = nil
	session.StagedRPs = nil
}

func (session *WorldSession) ListStructures() []StructureName {
	diskStructures := ListWorldStructures(session.WorldPath)
	var merged []StructureName
	for _, structure := range diskStructures {
		if !session.DeletedStructures[structure] {
			merged = append(merged, structure)
		}
	}
	for name := range session.StagedStructures {
		merged = append(merged, name)
	}
	return merged
}

func (session *WorldSession) Commit() error {
	dbPath := filepath.Join(session.WorldPath, "/db")
	database, err := leveldb.OpenFile(dbPath, &dbopt.Options{ReadOnly: false})
	if err != nil {
		return fmt.Errorf("World database failed to open: %v", err)
	}
	defer database.Close()
	batch := new(leveldb.Batch)
	for name, data := range session.StagedStructures {
		prefix := name.Prefix
		if prefix != "" {
			prefix = "mystructure"
		}
		key := "structuretemplate_" + prefix + ":" + name.Name
		batch.Put([]byte(key), data)
	}
	err = database.Write(batch, nil)
	if err != nil {
		return fmt.Errorf("Failed to write changes to levelDB: %v", err)
	}
	session.DiscardChanges()
	return nil
}

// Lists the structures inside a world's database.
// Returns only the name and ID of each structure.
func ListWorldStructures(path string) []StructureName  {
	path = filepath.Join(resolvePath(path), "/db")
	if path == "/db" {
		return []StructureName{}
	}
	structurePrefix := "structuretemplate_"
	excludedPrefixes := []string{"bot", "wedit"}
	database, err := leveldb.OpenFile(path, &dbopt.Options{ReadOnly:true})
	if err != nil {
		return []StructureName{}
	}
	structIterator := database.NewIterator(util.BytesPrefix([]byte(structurePrefix)), nil)
	defer structIterator.Release()
	var structNames []StructureName
	for structIterator.Next() {
		key := string(structIterator.Key())
		key = strings.TrimPrefix(key, structurePrefix)
		isExcluded := false
		for _, excludedPrefix := range excludedPrefixes {
			if strings.HasPrefix(key, excludedPrefix) {
				isExcluded = true
				break
			}
		}
		if isExcluded {
			continue
		}
		var (
			structName string
			structPrefix string
		)
		structPrefix, structName, wasFound := strings.Cut(key, ":")
		if !wasFound {
			structName = key
			structPrefix = ""
		}
		structNames = append(structNames, StructureName{Name: structName, Prefix: structPrefix})
	}
	return structNames
}

// Gets a structure from a world's level database by its name.
// Requires a name and prefix, although most useful structures will have the prefix "mystructure"
func GetStructureFromWorld(path string, name StructureName) (RawStructure, error) {
	path = filepath.Join(resolvePath(path)) + "/db"
	if path == "/db" {
		return RawStructure{}, fmt.Errorf("Invalid path to world")
	}
	database, err := leveldb.OpenFile(path, &dbopt.Options{ReadOnly:true})
	if err != nil {
		return RawStructure{}, err
	}
	defer database.Close()
	key := []byte("structuretemplate_" + name.Prefix + ":" + name.Name)
	value, err := database.Get(key, nil)
	return RawStructure(value), nil
}

// Gets a structure file from a pack specified by a path.
func GetStructureFromPack(path string, name string) (RawStructure, error) {
	var (
		structure RawStructure
		err error
	)
	structure, err = os.ReadFile(filepath.Join(path, "/structures", "/" + name + ".mcstructure"))
	if err != nil {
		return RawStructure{}, err
	}
	return structure, nil
}

// Lists the UUIDs and versions defined in the world_resource_packs.json and world_behavior_packs.json files for a world.
// The existence of an entry for a pack here does not garuantee that the pack exists in the world.
func ListWorldPacks(path string) (RP []PackID, BP []PackID) {
	path = resolvePath(path)
	RPfile, err := os.ReadFile(filepath.Join(path, "/world_resource_packs.json"))
	if err != nil {
		return []PackID{}, []PackID{}
	}
	BPfile, err := os.ReadFile(filepath.Join(path, "/world_behavior_packs.json"))
	if err != nil {
		return []PackID{}, []PackID{}
	}
	err = json.Unmarshal(RPfile, &RP)
	if err != nil {
		RP = []PackID{}
	}
	err = json.Unmarshal(BPfile, &BP)
	if err != nil {
		BP = []PackID{}
	}	
	return
}

//  Accepts the ID of a pack and attempts to grab it from the world. If it doesnt exist, returns an empty struct.
func GetBPFromWorld(worldPath string, ID PackID) (BehaviorPack, error) {
	BPpath := filepath.Join(worldPath, "/behavior_packs")
	BPfolders, err := os.ReadDir(BPpath)
	if err != nil {
		return BehaviorPack{}, err
	}
	var pack BehaviorPack
	for _, BPfolder := range BPfolders {
		if !BPfolder.IsDir() {
			continue
		}
		manifest, err := os.ReadFile(filepath.Join(BPpath, BPfolder.Name(), "/manifest.json"))
		if err != nil {
			continue
		}
		var header manifestHeader
		err = json.Unmarshal(manifest, &header)
		if err != nil {
			continue
		}
		if header.Header.UUID == ID.UUID && header.Header.Version == ID.Version {
			pack = BehaviorPack{
				ID: ID,
				Name: header.Header.Name,
				Description: header.Header.Description,
				Path: filepath.Join(BPpath, BPfolder.Name()),
			}
			break
		}
	}
	if pack.Name == "" {
		return BehaviorPack{}, nil
	}
	embeddedStructures, err := os.ReadDir(filepath.Join(pack.Path, "/structures"))
	if err == nil {
		for _, structure := range embeddedStructures {
			if filepath.Ext(structure.Name()) != ".mcstructure" {
				continue
			}
			pack.EmbeddedStructureNames = append(pack.EmbeddedStructureNames, strings.TrimSuffix(structure.Name(), ".mcstructure"))
		}
	}
	return pack, nil
}

// Accepts the ID of a pack and attempts to grab it from the world. If it doesnt exist, returns an empty struct.
func GetRPFromWorld(worldPath string, ID PackID) (ResourcePack, error) {
	RPpath := filepath.Join(worldPath, "/resource_packs")
	RPfolders, err := os.ReadDir(RPpath)
	if err != nil {
		return ResourcePack{}, err
	}
	var pack ResourcePack
	for _, RPfolder := range RPfolders {
		if !RPfolder.IsDir() {
			continue
		}
		manifest, err := os.ReadFile(filepath.Join(RPpath, RPfolder.Name(), "/manifest.json"))
		if err != nil {
			continue
		}
		var header manifestHeader
		err = json.Unmarshal(manifest, &header)
		if err != nil {
			continue
		}
		if header.Header.UUID == ID.UUID && header.Header.Version == ID.Version {
			pack = ResourcePack{
				ID: ID,
				Name: header.Header.Name,
				Description: header.Header.Description,
				Path: filepath.Join(RPpath, RPfolder.Name()),
			}
			break
		}
	}
	return pack, nil
}

/*
Holding the following until new changes are confirmed

func AddRPToWorld(packPath string, worldPath string) error { // TODO
	return nil
}


func AddBPToWorld(packPath string, worldPath string) error { // TODO
	return nil
}


func RemovePackFromWorld(id PackID, worldPath string, isBP bool) error { // TODO
	return nil
}

// Deletes a structure by the given name from a world.
func DeleteStructureFromWorld(structure StructureName, worldPath string) error {
	db, err := leveldb.OpenFile(filepath.Join(worldPath, "/db"), &dbopt.Options{ReadOnly: false})
	if err != nil {
		return err
	}
	defer db.Close()
	key := "structuretemplate_" + structure.Prefix + ":" + structure.Name
	return db.Delete([]byte(key), nil)
}

// Adds a structure to the world database. The key defaults to
// `"structuretemplate_mystructure:" + name`
// as if it were saved or imported by the player in-game
func (structure *RawStructure) AddToWorld(path string, name string) error {
	db, err := leveldb.OpenFile(filepath.Join(path, "/db"), &dbopt.Options{ReadOnly: false})
	if err != nil {
		return err
	}
	defer db.Close()
	key := "structuretemplate_mystructure:" + name
	return db.Put([]byte(key), *structure, nil)
}

func (structure *RawStructure) AddToWorldPack(packPath string, name string) error {
	structurePath := filepath.Join(packPath, "/"+ name + ".mcstructure")
	return os.WriteFile(structurePath, *structure, 0666)
}
*/
