package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	CfgCapacity int = 5000
	TmpCapacity int = 200

	Dir  string = "D"
	File string = "F"
	Link string = "L"
)

type FilePerms struct {
	Group string
	Owner string
	Perms string
}

type FilePermsMap map[string]FilePerms

type FileInfo struct {
	Perms  FilePerms
	Type   string
	Target string
}

type FileMap map[string]FileInfo

func (f FileInfo) string(path *string) string {
	if f.Target == "" {
		return fmt.Sprintf("%s %s %s %s\n", *path, f.Perms.Owner, f.Perms.Group, f.Perms.Perms)
	}
	return fmt.Sprintf("%s %s %s %s %s\n", *path, f.Perms.Owner, f.Perms.Group, f.Perms.Perms, f.Target)
}

func (f FileMap) FindBytype(root, ftype string) []FileInfo {
	files := make([]FileInfo, 0, TmpCapacity)

	for path, finfo := range f {
		if strings.HasPrefix(path, root) && path != root && finfo.Type == ftype {
			files = append(files, finfo)
		}
	}
	return files
}

func (f FileMap) String() string {
	var str strings.Builder

	for path, info := range f {
		str.WriteString(info.string(&path))
	}
	return str.String()
}

func FrequentItem[T comparable](slice []T) T {
	freqMap := make(map[T]int, len(slice))

	var maxItem T
	var maxCount int

	for _, item := range slice {
		freqMap[item]++
		if freqMap[item] > maxCount {
			maxItem = item
			maxCount = freqMap[item]
		}
	}

	return maxItem
}

func GetFileType(path string) (string, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	switch stat.Mode() & os.ModeType {
	case os.ModeDir:
		return Dir, nil
	case os.ModeSymlink:
		return Link, nil
	default:
		return File, nil
	}
}

func RootToBase(path, root string) string {
	return strings.Replace(path, root, filepath.Base(root), 1)
}

func ExpandUser(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}
