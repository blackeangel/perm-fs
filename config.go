package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/Mrakorez/perm-fs/common"
)

func loadConfig(config io.Reader) (common.FileMap, error) {
	fileMap := make(common.FileMap, common.CfgCapacity)

	scanner := bufio.NewScanner(config)
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)

		if len(fields) == 5 {
			// if there are 5 fields, it's a symbolic link
			info := common.FileInfo{
				Target: fields[4],
				Perms: common.FilePerms{
					Group: fields[2],
					Owner: fields[1],
					Perms: fields[3],
				},
			}

			fileMap[fields[0]] = info
			continue
		}

		info := common.FileInfo{
			Perms: common.FilePerms{
				Group: fields[2],
				Owner: fields[1],
				Perms: fields[3],
			},
		}

		fileMap[fields[0]] = info
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return fileMap, nil
}

func updateConfig(fsConfigFileMap, currentFileMap common.FileMap) {
	// delete non-existent files from fs_config
	for file := range fsConfigFileMap {
		if _, ok := currentFileMap[file]; !ok {
			delete(fsConfigFileMap, file)
		}
	}

	// add new files to fs_config
	for file, info := range currentFileMap {
		if _, ok := fsConfigFileMap[file]; !ok {
			fsConfigFileMap[file] = info
		}
	}
}

func getConfigReader(path string) (*bytes.Reader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func saveConfig(path string, config common.FileMap) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(config.String())
	if err != nil {
		return err
	}

	return nil
}
