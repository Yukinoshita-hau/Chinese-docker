package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"configuraiton"
)

type Replica struct {
	Agent  string `json:"agent"`
	Status string `json:"status"` // "running", "stopped", etc.
}

type ServiceInfo struct {
	Name        string    `json:"name"`
	ReplicaCount int       `json:"replica-count"`
	Replicas    []Replica `json:"replicas"`
}

type AgentStatus struct {
	Services []ServiceInfo `json:"services"`
}

func GetServiceInfoFromAgent(agentURL string) ([]ServiceInfo, error) {
	statusURL := fmt.Sprintf("%s/status", agentURL)

	resp, err := http.Get(statusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch status from agent %s: %v", agentURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get valid response from %s: %s", agentURL, resp.Status)
	}

	var agentStatus AgentStatus
	if err := json.NewDecoder(resp.Body).Decode(&agentStatus); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %v", agentURL, err)
	}

	return agentStatus.Services, nil
}

func GetProcessInfoForAllAgents() (map[string][]ServiceInfo, error) {
	configuration, err := configuration.ReadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %v", err)
	}

	var wg sync.WaitGroup
	agentResults := make(map[string][]ServiceInfo)
	var mu sync.Mutex

	for _, agent := range configuration.Agents {
		wg.Add(1)
		go func(agentURL string) {
			defer wg.Done()

			services, err := GetServiceInfoFromAgent(agentURL)
			if err != nil {
				log.Printf("Error getting service info from agent %s: %v", agentURL, err)
				return
			}

			mu.Lock()
			agentResults[agentURL] = services
			mu.Unlock()
		}(agent)
	}

	wg.Wait()

	return agentResults, nil
}