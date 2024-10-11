package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Mrakorez/perm-fs/common"
)

func main() {
	log.SetFlags(log.Ltime + log.Lshortfile)

	if len(os.Args) != 3 {
		fmt.Println("Usage: perm-fs <target-dir> <fs-config>")
		os.Exit(1)
	}

	targetDirPath, err := common.ExpandUser(filepath.Clean(os.Args[1]))
	if err != nil {
		log.Fatalln(err)
	}
	fsConfigPath, err := common.ExpandUser(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}

	fsConfigReader, err := getConfigReader(fsConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	defaultPermsMap, err := newDefaultPermsMap(targetDirPath, fsConfigReader)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = fsConfigReader.Seek(0, 0)
	if err != nil {
		log.Fatalln(err)
	}

	fsConfigFileMap, err := loadConfig(fsConfigReader)
	if err != nil {
		log.Fatalln(err)
	}

	currentFileMap, err := loadFiles(targetDirPath, defaultPermsMap)
	if err != nil {
		log.Fatalln(err)
	}

	updateConfig(fsConfigFileMap, currentFileMap)

	rootBase := common.RootToBase(targetDirPath, targetDirPath)
	info := currentFileMap[rootBase]

	delete(currentFileMap, rootBase)
	delete(fsConfigFileMap, rootBase)

	currentFileMap["/"] = info
	fsConfigFileMap["/"] = info

	err = saveConfig(fsConfigPath, fsConfigFileMap)
	if err != nil {
		log.Fatalln(err)
	}
}
