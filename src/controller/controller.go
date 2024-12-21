package main

import (
	"configuration"
	"encoding/json"
	"file"
	"fmt"
	"log"
	"net/http"
	"service"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result := service.CollectStatusData()
	statusJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Unable to generate status", http.StatusInternalServerError)
		return
	}

	w.Write(statusJSON)
}

func StartHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	configuration, err := configuration.ReadConfiguration()
	if err != nil {
		fmt.Errorf(err.Error())
	}
	http.Get(configuration.OwnAgent + "/start?serviceID=" + serviceID)
}

func AddServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	if serviceID == "" {
		http.Error(w, "serviceID is required", http.StatusBadRequest)
		return
	}

	configuration, err := configuration.ReadConfiguration()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading configuration: %v", err), http.StatusInternalServerError)
		return
	}

	servicePath := "../payload-service/" + serviceID

	for i := 0; i < len(configuration.Agents); i++ {
		err = file.SendFolderToAgent(configuration.Agents[i] +"/upload?serviceID="+serviceID, servicePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to send service folder: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte("Service added successfully"))
}

func deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("serviceID")
	configuration, err := configuration.ReadConfiguration()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	for i := 0; i < len(configuration.Agents); i++ {
		http.Get(configuration.Agents[i] + "/remove?serviceID=" + serviceID)
	}
}

func main() {
	http.HandleFunc("/add", AddServiceHandler)
	http.HandleFunc("/status", StatusHandler)
	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/delete", deleteServiceHandler)

	fmt.Println("Starting controller on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
