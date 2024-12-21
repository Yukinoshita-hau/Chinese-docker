package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"configuration"
)

var (
	mutex         sync.Mutex
)

type ReplicaStatus struct {
	Agent  string `json:"agent"`
	Status string `json:"status"`
}

type ServiceStatus struct {
	Name     string           `json:"name"`
	Replicas []ReplicaStatus  `json:"replicas"`
}

func checkAgentStatus(agentURL string, serviceID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/status?serviceID=%s", agentURL, serviceID))
	if resp.StatusCode == http.StatusInternalServerError {
		return "", fmt.Errorf("bad response status")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get status from agent %s: %v", agentURL, err)
	}
	defer resp.Body.Close()

	status, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from agent %s: %v", agentURL, err)
	}

	return string(status), nil
}

func CollectStatusData() []ServiceStatus {
	configuration, err := configuration.ReadConfiguration()
	if err != nil {
		fmt.Errorf(err.Error())
	}

	arrServiceStatus := make([]ServiceStatus, 0)
	configServices := configuration.Services
	configAgents := configuration.Agents
	for i := 0; i < len(configServices); i++ {
		serviceStatus := ServiceStatus{
			Name: configServices[i].Name,
			Replicas: make([]ReplicaStatus, 0),
		}
		for j := 0; j < len(configAgents); j++ {
			status, err := checkAgentStatus(configAgents[j], configServices[i].Name)
			if err != nil {
				serviceStatus.Replicas = append(serviceStatus.Replicas, ReplicaStatus{
					Agent:  configAgents[j],
					Status: "error",
				})
			} else {
				serviceStatus.Replicas = append(serviceStatus.Replicas, ReplicaStatus{
					Agent:  configAgents[j],
					Status: status,
				})
			}
		}
		arrServiceStatus = append(arrServiceStatus, serviceStatus)
	}


	return arrServiceStatus
}