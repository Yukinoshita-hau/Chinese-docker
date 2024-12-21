package isolated

import (
	"bytes"
	"configuration"
	"file"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

// Функция для запуска процесса с приложением и получения его метрик
func RunServiceWithApp(serviceID, appDir string) (string, error) {
	logFile := fmt.Sprintf("/tmp/%s.log", serviceID)
	pidFile := fmt.Sprintf("/tmp/%s.pid", serviceID)
	serviceArrScript, err := configuration.FindServiceByName(serviceID)
	if err != nil {
		return "", fmt.Errorf("incorrect configuration")
	}
	file.CreateBashScript(serviceArrScript.StartScript, "../payload-service/"+serviceID+"/start.sh")

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Application directory does not exist: %s", appDir)
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		"mkdir -p /mnt/app && "+
			"mount --bind %s /mnt/app && "+
			"echo 'Mounted app directory to /mnt/app' && "+
			"cd /mnt/app && "+
			"echo 'Running service %s' && "+
			"echo $$ > %s && "+
			"./start.sh >> %s 2>&1",
		appDir, serviceID, pidFile, logFile))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	log.Printf("Running command: %v", cmd)

	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("Error while starting service %s: %v\n", serviceID, err)
	}

	pid := cmd.Process.Pid
	log.Printf("Service %s started with PID: %d", serviceID, pid)
	arrPid, err := GetMainProcessPID(pid)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	fmt.Println("-------------------", arrPid)

	err = ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	if err != nil {
		return "", fmt.Errorf("Failed to write PID to file: %v", err)
	}

	log.Printf("Command Output: %s", out.String())
	log.Printf("Command Error Output: %s", stderr.String())

	log.Printf("Service %s has started successfully", serviceID)


	lastCommand, err := getLastCommandFromStartScript("../payload-service/" + serviceID + "/start.sh")
	if err != nil {
		return "", fmt.Errorf("не удалось получить последнюю команду из start.sh: %v", err)
	}

	time.Sleep(2 * time.Second)
	cmd = exec.Command("pgrep", "-f", lastCommand)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("не удалось выполнить pgrep для поиска процессов: %v", err)
	}

	pids := strings.Split(string(output), "\n")
	for _, pidStr := range pids {
		fmt.Println(pidStr, pids)
		pid, _ := strconv.Atoi(pidStr)
		cpuPercent, memoryUsage, err := GetProcessMetrics(pid)
		if err != nil {
			log.Printf("Error getting metrics for process %d: %v", pid, err)
		} else {
			log.Printf("Process %d metrics: CPU usage: %.2f%%, Memory usage: %.2f MB", pid, cpuPercent, memoryUsage)
		}
	}

	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		log.Printf("PID file not found: %s", pidFile)
	} else {
		log.Printf("PID file created: %s", pidFile)
	}

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		log.Printf("Log file not found: %s", logFile)
	} else {
		log.Printf("Log file created: %s", logFile)
	}

	return logFile, nil
}

// Функция для получения метрик использования CPU и памяти процесса по его PID
func GetProcessMetrics(pid int) (float64, float64, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось создать объект процесса для PID %d: %v", pid, err)
	}

	cpuPercent, err := p.CPUPercent()
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось получить данные по CPU для процесса с PID %d: %v", pid, err)
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось получить данные по памяти для процесса с PID %d: %v", pid, err)
	}

	return cpuPercent, float64(memInfo.RSS) / (1024 * 1024), nil 
}

// Функция для поиска всех процессов, запущенных определенной командой
func findProcessByCommand(command string) ([]int, error) {
	cmd := exec.Command("ps", "-eo", "pid,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить команду ps: %v", err)
	}

	var pids []int
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if strings.Contains(parts[1], command) {
			pids = append(pids, pid)
			log.Printf("Найден процесс с командой %s и PID: %d", command, pid)
		}
	}

	return pids, nil
}

// Функция для получения последней команды из start.sh
func getLastCommandFromStartScript(startScriptPath string) (string, error) {
	data, err := ioutil.ReadFile(startScriptPath)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл start.sh: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	var lastCommand string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			lastCommand = line
		}
	}

	if lastCommand == "" {
		return "", fmt.Errorf("файл start.sh не содержит команд")
	}

	return lastCommand, nil
}

// Функция для завершения процесса по PID и очистки следов
func StopService(serviceID, startScriptPath string) error {
	lastCommand, err := getLastCommandFromStartScript(startScriptPath)
	if err != nil {
		return fmt.Errorf("не удалось получить последнюю команду из start.sh: %v", err)
	}

	cmd := exec.Command("pgrep", "-f", lastCommand)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("не удалось выполнить pgrep для поиска процессов: %v", err)
	}

	pids := strings.Split(string(output), "\n")
	for _, pidStr := range pids {
		if pidStr == "" {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Printf("Не удалось преобразовать PID %s: %v", pidStr, err)
			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Printf("Не удалось найти процесс с PID %d: %v", pid, err)
			continue
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Printf("Не удалось отправить SIGTERM процессу %d: %v", pid, err)
			err = process.Signal(syscall.SIGKILL)
			if err != nil {
				log.Printf("Не удалось отправить SIGKILL процессу %d: %v", pid, err)
			}
		}

		time.Sleep(2 * time.Second)

		processState, err := process.Wait()
		if err != nil {
			log.Printf("Процесс с PID %d не завершился: %v", pid, err)
		} else {
			log.Printf("Процесс с PID %d завершен: %v", pid, processState)
		}
	}

	pidFile := fmt.Sprintf("/tmp/%s.pid", serviceID)
	pidData, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл PID: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return fmt.Errorf("неверный формат PID в файле %s: %v", pidFile, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("не удалось найти процесс с PID %d: %v", pid, err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Printf("Не удалось отправить SIGTERM процессу %d: %v", pid, err)
		err = process.Signal(syscall.SIGKILL)
		if err != nil {
			log.Printf("Не удалось отправить SIGKILL процессу %d: %v", pid, err)
		}
	}

	time.Sleep(2 * time.Second)

	processState, err := process.Wait()
	if err != nil {
		log.Printf("Процесс с PID %d не завершился: %v", pid, err)
	} else {
		log.Printf("Процесс с PID %d завершен: %v", pid, processState)
	}


	return nil
}

func terminateProcess(pid int) error {
	cmd := exec.Command("kill", "-TERM", strconv.Itoa(pid))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", pid, err)
	}
	return nil
}

func IsProcessRunning(pid int) bool {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid))
	err := cmd.Run()

	return err == nil
}

func GetMainProcessPID(ppid int) ([]string, error) {
	pidFile := "/tmp/main.pid"
	pidData, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read main PID from file: %v", err)
	}

	pid := strings.TrimSpace(string(pidData))
	log.Printf("Main process PID: %s", pid)

	cmd := exec.Command("ps", "--ppid", pid, "-o", "pid,cmd")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute ps command: %v", err)
	}

	fmt.Println("Output of ps command:\n", out.String())

	var childPIDs []string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "PID") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 1 && strings.Contains(parts[1], "./main") {
			childPID := parts[0]
			childPIDs = append(childPIDs, childPID)
		}
	}

	return childPIDs, nil
}

func GetPID(serviceID string) (int, error) {
	pidFile := fmt.Sprintf("/tmp/%s.pid", serviceID)

	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file %s: %v", pidFile, err)
	}

	pidStr := string(data)
	pidStr = strings.TrimSpace(pidStr)

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file %s: %v", pidFile, err)
	}

	return pid, nil
}
