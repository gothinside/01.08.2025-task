package main

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type IFileManager interface {
	IsDownload() bool
	GetFileStatus() string
	GetFileCount() int
	DownloadFile(string)
	Close()
}

type FileManager struct {
	ZipWriter  *zip.Writer
	Archive    *os.File
	FileStatus map[string]string
	mu         *sync.Mutex
}

func (FM *FileManager) IsDownload() bool {
	if FM.GetFileCount() < 3 {
		return false
	}
	FM.mu.Lock()
	defer FM.mu.Unlock()
	for _, status := range FM.FileStatus {
		if status == "Status: Downloading" {
			return false
		}
	}
	return true
}

func CreateFileManager(Archive *os.File, ZipWriter *zip.Writer) *FileManager {
	return &FileManager{
		ZipWriter:  ZipWriter,
		Archive:    Archive,
		FileStatus: make(map[string]string),
		mu:         &sync.Mutex{},
	}
}

func (FM *FileManager) GetFileStatus() string {
	FM.mu.Lock()
	defer FM.mu.Unlock()

	var sb strings.Builder
	for fileURL, status := range FM.FileStatus {
		sb.WriteString(fileURL)
		sb.WriteString(": ")
		sb.WriteString(status)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (FM *FileManager) Close() {
	FM.ZipWriter.Close()
	FM.Archive.Close()
}

func (FM *FileManager) GetFileCount() int {
	return len(FM.FileStatus)
}

func (FM *FileManager) DownloadFile(FileURL string) {
	FM.mu.Lock()
	FM.FileStatus[FileURL] = "Status: Downloading"
	FM.mu.Unlock()

	resp, err := http.Get(FileURL)
	if err != nil {
		FM.mu.Lock()
		FM.FileStatus[FileURL] = "Status: Error: " + err.Error()
		FM.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		FM.mu.Lock()
		FM.FileStatus[FileURL] = "Status: Invalid resource"
		FM.mu.Unlock()
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		FM.mu.Lock()
		FM.FileStatus[FileURL] = "Status: Error: " + err.Error()
		FM.mu.Unlock()
		return
	}
	FM.mu.Lock()
	parts := strings.Split(FileURL, "/")
	filename := parts[len(parts)-1]

	zwr, err := FM.ZipWriter.Create(filename)
	if err != nil {
		FM.mu.Lock()
		FM.FileStatus[FileURL] = "Status: Error: " + err.Error()
		FM.mu.Unlock()
		return
	}

	_, err = zwr.Write(data)
	FM.mu.Unlock()
	if err != nil {
		FM.mu.Lock()
		FM.FileStatus[FileURL] = "Status: Error: " + err.Error()
		FM.mu.Unlock()
		return
	}

	FM.mu.Lock()
	FM.FileStatus[FileURL] = "Status: Downloaded"
	FM.mu.Unlock()
}
