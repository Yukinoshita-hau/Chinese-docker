package main

import (
	"configuration"
	"file"
	"fmt"
	"io"
	"isolated"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	saveDir := "../payload-service/" + serviceID
	fmt.Println("Attempting to create directory:", saveDir)

	stat, err := os.Stat(saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(saveDir, 0777)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
				fmt.Println("Error creating directory:", err)
				return
			}
			fmt.Println("Directory created:", saveDir)
		} else {
			http.Error(w, fmt.Sprintf("Failed to check if directory exists: %v", err), http.StatusInternalServerError)
			return
		}
	} else if !stat.IsDir() {
		http.Error(w, "A file exists with the same name as the directory", http.StatusInternalServerError)
		fmt.Println("A file exists where the directory should be:", saveDir)
		return
	} else {
		fmt.Println("Directory already exists:", saveDir)
	}

	fileF, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}
	defer fileF.Close()

	destFilePath := filepath.Join(saveDir, "received.zip")
	destFile, err := os.Create(destFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create destination file: %v", err), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, fileF)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("File saved:", destFile.Name())

	err = file.UnzipFolder(destFilePath, saveDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to unzip file: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("File unzipped:", saveDir)

	w.Write([]byte("Folder received, unpacked, and sent successfully"))
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	pidFile := "/tmp/" + serviceID + ".pid"

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read PID file: %v", err), http.StatusInternalServerError)
		return
	}

	pid := strings.TrimSpace(string(pidBytes))

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid PID in file: %v", err), http.StatusInternalServerError)
		return
	}

	if isolated.IsProcessRunning(pidInt) {
		w.Write([]byte("running"))
	} else {
		w.Write([]byte("stopped"))
	}
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	configuration, err := configuration.ReadConfiguration()
	if err != nil {
		fmt.Errorf(err.Error())
		return

	}

	for i := 0; i < len(configuration.Agents); i++ {
		http.Get("http://" + configuration.Agents[i] + "/upload?serviceID=" + serviceID)
	}
}

func StopHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	isolated.StopService(serviceID, "../payload-service/"+serviceID+"/start.sh")
	w.Write([]byte("Service has been stopped"))
}

func StartHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	_, err := isolated.RunServiceWithApp(serviceID, "../payload-service/"+serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write([]byte("Service started now!"))
}

func RemoveHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	err := os.RemoveAll("../payload-service/" + serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write([]byte(serviceID + ": has been removed"))
}

func GetLogHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	file, err := os.ReadFile("/tmp/" + serviceID + ".log")
	if err != nil {
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
	}
	w.Write(file)
}

func main() {
	http.HandleFunc("/add", AddHandler)
	http.HandleFunc("/remove", RemoveHandler)
	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/stop", StopHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/status", StatusHandler)
	http.HandleFunc("/logs", GetLogHandler)

	fmt.Println("Starting agent on port 8083...")
	log.Fatal(http.ListenAndServe(":8083", nil))
}