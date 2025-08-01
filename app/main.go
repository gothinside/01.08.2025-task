package main

import (
	"fmt"
	"net/http"
)

type ITaskManager interface {
	TaskStartDownloadFiles(w http.ResponseWriter, r *http.Request)
	AddTask(w http.ResponseWriter, r *http.Request)
	GetTaskStatus(w http.ResponseWriter, r *http.Request)
	GetFile(w http.ResponseWriter, r *http.Request)
}

func main() {
	TM := &TaskManager{Tasks: make(map[int]*Task), count: 0, MaxCount: 3}
	var ITM ITaskManager
	ITM = TM
	mux := http.NewServeMux()
	mux.HandleFunc("/CreateTask", ITM.AddTask)
	mux.HandleFunc("/task/{id}/download", ITM.TaskStartDownloadFiles)
	mux.HandleFunc("/task/{id}/status", ITM.GetTaskStatus)
	mux.HandleFunc("/archives/{archive}", ITM.GetFile)
	http.ListenAndServe(":8080", mux)
	for _, task := range TM.Tasks {
		fmt.Println(task)
		task.FM.Close()
	}
}
