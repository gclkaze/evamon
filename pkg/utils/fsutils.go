package utils

import "os"

func FolderExists(folder string) bool {
	info, err := os.Stat(folder)
	return !os.IsNotExist(err) && info.IsDir()
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func CreateFolder(folder string) error {
	if folder == "." {
		return nil
	}
	err := os.Mkdir(folder, os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
	}
	return err
}
