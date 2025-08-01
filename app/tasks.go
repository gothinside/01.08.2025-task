package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Task struct {
	TaskID int
	FM     IFileManager
}

func CreateTask(TaskID int) *Task {
	archive, _ := os.Create(fmt.Sprintf("Archiv%v.zip", TaskID))
	zw := zip.NewWriter(archive)
	FM := CreateFileManager(archive, zw)
	return &Task{
		TaskID: TaskID,
		FM:     FM,
	}
}

func (task *Task) GetStatus() string {
	count := task.FM.GetFileCount()
	if count == 0 {
		return "Files not added"
	} else if count > 0 && count < 3 {
		return task.FM.GetFileStatus()
	}
	task.FM.Close()
	status := task.FM.GetFileStatus()
	link := fmt.Sprintf("Archive is ready: localhost:8080/archives/Archiv%v.zip", task.TaskID)
	return status + "\n" + link
}

type TaskManager struct {
	Tasks    map[int]*Task
	count    int
	MaxCount int
}

func (TM *TaskManager) TaskStartDownloadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	taskID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}

	task, ok := TM.Tasks[taskID]
	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	FileCount := task.FM.GetFileCount()
	if FileCount == 3 {
		http.Error(w, "This task is full", http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	URLs := make(map[string][]string)
	err = json.Unmarshal(data, &URLs)
	if err != nil {
		http.Error(w, "Failed read json", http.StatusBadRequest)
		return
	}

	if FileCount+len(URLs["URLs"]) > 3 {
		http.Error(w, "Too many files", http.StatusBadRequest)
		return
	}
	for _, url := range URLs["URLs"] {
		if strings.HasSuffix(url, "jpeg") || strings.HasSuffix(url, "pdf") {
			go task.FM.DownloadFile(url)
			w.Write([]byte(fmt.Sprintf(url + " downloading")))
		} else {
			http.Error(w, "Only jpeg or pdf", http.StatusBadRequest)
		}
	}

}

func (TM *TaskManager) AddTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Now allowed method", http.StatusMethodNotAllowed)
	} else {
		count := 0
		for _, task := range TM.Tasks {
			if !task.FM.IsDownload() {
				count++
			}
		}
		if count == 3 {
			http.Error(w, "Сервер загружен", http.StatusBadRequest)
			return
		}

		task_id := TM.count + 1
		TM.count++
		TM.Tasks[task_id] = CreateTask(task_id)
		w.Write([]byte(fmt.Sprintf("Task with id %v created", task_id)))
	}
}

func (TM *TaskManager) GetFile(w http.ResponseWriter, r *http.Request) {
	filePath := r.PathValue("archive")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

func (TM *TaskManager) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	taskID, err := strconv.Atoi(id)
	_, ok := TM.Tasks[taskID]
	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}
	w.Write([]byte(TM.Tasks[taskID].GetStatus()))
}
