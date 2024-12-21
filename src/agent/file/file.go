package file

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnzipFolder(zipFilePath, destDir string) error {
	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipFile.Close()

	fileInfo, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat zip file: %v", err)
	}

	zipReader, err := zip.NewReader(zipFile, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %v", err)
	}

	for _, file := range zipReader.File {
		destPath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer destFile.Close()

		zipFileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}
		defer zipFileReader.Close()

		_, err = io.Copy(destFile, zipFileReader)
		if err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}

	return nil
}

func CreateBashScript(commands []string, scriptPath string) error {
	file, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to create script file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString("#!/bin/bash\n")
	if err != nil {
		return fmt.Errorf("failed to write shebang: %v", err)
	}

	for _, command := range commands {
		_, err = file.WriteString(command + "\n")
		if err != nil {
			return fmt.Errorf("failed to write command to script: %v", err)
		}
	}

	err = os.Chmod(scriptPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to set file permissions: %v", err)
	}

	return nil
}