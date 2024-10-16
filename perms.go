package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mrakorez/perm-fs/common"
)

var (
	cachedFileFrequentPerms common.FilePerms
	cachedDirFrequentPerms  common.FilePerms
)

type defaultPerms struct {
	dir  common.FilePerms
	file common.FilePerms
	link common.FilePerms
}

type defaultPermsMap map[string]defaultPerms

func getFrequentPerms(files common.FileMap, fileType string) common.FilePerms {
	if cachedFileFrequentPerms == (common.FilePerms{}) {
		fileBuff := make([]common.FilePerms, 0, len(files))

		for _, info := range files {
			if info.Type == common.File {
				fileBuff = append(fileBuff, info.Perms)
			}
		}

		cachedFileFrequentPerms = common.FrequentItem(fileBuff)
		if cachedFileFrequentPerms == (common.FilePerms{}) {
			panic("unable to determine default file permissions: no valid file info found")
		}
	}

	if cachedDirFrequentPerms == (common.FilePerms{}) {
		dirBuff := make([]common.FilePerms, 0, common.TmpCapacity)

		for _, info := range files {
			if info.Type == common.Dir {
				dirBuff = append(dirBuff, info.Perms)
			}
		}

		cachedDirFrequentPerms = common.FrequentItem(dirBuff)
		if cachedDirFrequentPerms == (common.FilePerms{}) {
			panic("unable to determine default directory permissions: no valid directory info found")
		}
	}

	if fileType == common.File {
		return cachedFileFrequentPerms
	}
	if fileType == common.Dir {
		return cachedDirFrequentPerms
	}

	panic(fmt.Sprintf("unknown file type: %s. Unable to determine default permissions", fileType))
}

func newCustomFsConfig(target string, fsConfig io.Reader) (common.FileMap, error) {
	customFsConfig := make(common.FileMap, common.CfgCapacity)

	targetDir := filepath.Dir(target)

	scanner := bufio.NewScanner(fsConfig)
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		if fields[0] == "/" {
			fields[0] = filepath.Base(target)
		}

		path := filepath.Join(targetDir, fields[0])

		fileType, err := common.GetFileType(path)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		if fileType == common.Link {
			continue
		}

		info := common.FileInfo{
			Type: fileType,
			Perms: common.FilePerms{
				Owner: fields[1],
				Group: fields[2],
				Perms: fields[3],
			},
		}
		customFsConfig[common.RootToBase(path, target)] = info
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return customFsConfig, nil
}

func newDefaultPermsMap(targetDir string, fsConfig io.Reader) (defaultPermsMap, error) {
	customConfig, err := newCustomFsConfig(targetDir, fsConfig)
	if err != nil {
		return nil, err
	}

	fsDirs := make(common.FileMap, common.TmpCapacity)

	for path, info := range customConfig {
		if info.Type == common.Dir {
			fsDirs[path] = info
		}
	}

	permsMap := make(defaultPermsMap, common.TmpCapacity)

	for path := range fsDirs {

		perms := defaultPerms{
			link: common.FilePerms{
				Owner: "0",
				Group: "0",
				Perms: "0777",
			},
		}

		foundFiles := make([]common.FileInfo, 0, common.TmpCapacity)
		foundDirs := make([]common.FileInfo, 0, common.TmpCapacity)

		fileSearchPath, dirSearchPath := path, path

		for len(foundFiles) == 0 {
			if fileSearchPath == "." {
				perms.file = getFrequentPerms(customConfig, common.File)
				break
			}

			foundFiles = customConfig.FindByType(fileSearchPath, common.File)
			fileSearchPath = filepath.Dir(fileSearchPath)
		}

		for len(foundDirs) == 0 {
			if dirSearchPath == "." {
				perms.dir = getFrequentPerms(customConfig, common.Dir)
				break
			}

			foundDirs = customConfig.FindByType(dirSearchPath, common.Dir)
			dirSearchPath = filepath.Dir(dirSearchPath)
		}

		if len(foundFiles) != 0 {
			perms.file = common.FrequentItem(foundFiles).Perms
		}
		if len(foundDirs) != 0 {
			perms.dir = common.FrequentItem(foundDirs).Perms
		}

		permsMap[path] = perms
	}

	return permsMap, nil
}
