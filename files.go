package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Mrakorez/perm-fs/common"
)

func loadFiles(root string, permsMap defaultPermsMap) (common.FileMap, error) {
	currentFileMap := make(common.FileMap, common.CfgCapacity)

	getFileInfo := func(path, rootBase string, d fs.DirEntry) (common.FileInfo, error) {
		var filePerms defaultPerms

		for filePerms == (defaultPerms{}) && rootBase != "." {
			filePerms, rootBase = permsMap[rootBase], filepath.Dir(rootBase)
		}

		if filePerms == (defaultPerms{}) {
			panic("failed to find appropriate permissions for file or directory: " + path)
		}

		var info common.FileInfo

		switch d.Type() & os.ModeType {
		case os.ModeDir:
			info.Type = common.Dir
			info.Perms = filePerms.dir
		case os.ModeSymlink:
			target, err := os.Readlink(path)
			if err != nil {
				return info, err
			}
			info.Type = common.Link
			info.Perms = filePerms.link
			info.Target = target
		default:
			info.Type = common.File
			info.Perms = filePerms.file
		}

		return info, nil
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rootBase := common.RootToBase(path, root)

		info, err := getFileInfo(path, rootBase, d)
		if err != nil {
			return err
		}

		currentFileMap[rootBase] = info

		return nil
	})
	if err != nil {
		return nil, err
	}

	return currentFileMap, nil
}
