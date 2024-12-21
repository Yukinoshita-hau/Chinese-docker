package file

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SendFolderToAgent(url, folderPath string) error {
	fmt.Println("Archiving folder...")

	zipFileName := "/tmp/folder.zip"
	err := ZipFolder(folderPath, zipFileName)
	if err != nil {
		return fmt.Errorf("failed to zip folder: %v", err)
	}

	zipFile, err := os.Open(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipFile.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(zipFileName))
	if err != nil {
		return fmt.Errorf("failed to create form file part: %v", err)
	}
	_, err = io.Copy(part, zipFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Folder successfully sent")
	} else {
		return fmt.Errorf("failed to send folder, server returned: %v", resp.Status)
	}

	return nil
}

func ZipFolder(folderPath, zipFileName string) error {
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == folderPath {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipEntry, err := zipWriter.Create(strings.TrimPrefix(path, folderPath+"/"))
		if err != nil {
			return err
		}

		_, err = io.Copy(zipEntry, file)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk folder: %v", err)
	}

	return nil
}
