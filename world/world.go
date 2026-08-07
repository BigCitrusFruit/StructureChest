package world

import (
	"fmt"
	"os"
	"path/filepath"
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
	Version string
	UUID string
	Description string
	EmbeddedStructures []RawStructure
}

type ResourcePack struct {
	Name string
	Version string
	UUID string
	Description string
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

func GetStructureFromWorld(path string, name StructureName) (RawStructure, error) {
	path = filepath.Join(resolvePath(path)) + "/db"
	if path == "/db" {
		return RawStructure{}, fmt.Errorf("Invalid path to world")
	}
	database, err := leveldb.OpenFile(path, &dbopt.Options{ReadOnly:true})
	if err != nil {
		return RawStructure{}, err
	}
	key := []byte("structuretemplate_" + name.Prefix + name.Name)
	value, err := database.Get(key, nil)
	return RawStructure(value), nil
}


func ListWorldPacks(path string) []string {
	path = resolvePath(path)
	return nil
}
