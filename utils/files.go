package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetMyPath() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

func AbsFilePath(filename string) (string, error) {
	return filepath.Abs(filename)
}

func CleanPathInput(path string) string {
	return filepath.Clean(path)
}

func FilenameWoExt(path string) string {
	path = BaseFilename(path)
	return path[:len(path)-len(filepath.Ext(path))]
}

func BaseFilename(path string) string {
	return filepath.Base(path)
}

func FileExists(filename string) bool {
	i, err := os.Stat(filename)
	return err == nil && !i.IsDir()
}

func DirExists(dirname string) bool {
	i, err := os.Stat(dirname)
	return err == nil && i.IsDir()
}

func RecursiveSearchDirForFile(currDir, searchFilename string) (string, error) {
	i, err := os.Stat(currDir)
	if err != nil {
		return "", err
	}

	if i.IsDir() {
		if files, err := os.ReadDir(currDir); err == nil {
			for _, file := range files {
				newPath := filepath.Join(currDir, file.Name())
				path, err := RecursiveSearchDirForFile(newPath, searchFilename)
				if path != "" {
					return path, nil
				} else if err != nil {
					return "", err
				}
			}
		} else {
			return "", err
		}
	} else {
		if i.Name() == searchFilename {
			return currDir, nil
		}
	}

	return "", nil
}
