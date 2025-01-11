package cmd

import (
	"os"
	"path"
	"strings"
	"sync"
)

type Queue struct {
	// queue is a slice of strings
	sync.Mutex
	filesToUpload []string
}

// AddFile adds a file to the queue
func (q *Queue) AddFile(filename string) {
	q.Lock()
	defer q.Unlock()
	// Verify if the file is already in the queue
	for _, f := range q.filesToUpload {
		if f == filename {
			return
		}
	}
	q.filesToUpload = append(q.filesToUpload, filename)
}

// ReturnFile returns a file to the queue
func (q *Queue) ReturnFile(filename string) {
	q.Lock()
	defer q.Unlock()
	// Verify if the file is already in the queue
	for _, f := range q.filesToUpload {
		if f == filename {
			return
		}
	}
	q.filesToUpload = append([]string{filename}, q.filesToUpload...)
}

// Next returns the next file in the queue
func (q *Queue) Next() string {
	q.Lock()
	defer q.Unlock()
	if len(q.filesToUpload) == 0 {
		return ""
	}
	file := q.filesToUpload[0]
	q.filesToUpload = q.filesToUpload[1:]
	return file
}

// Len returns the length of the queue
func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()
	return len(q.filesToUpload)
}

func UploadFilesToSumo(wd string, url string) error {
	// Get list of files in the working directory and check if there are any batch files
	files, err := os.ReadDir(wd)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), batchFilenamePrefix) {
			filename := path.Join(wd, file.Name())
			DebugLogger.Printf("Found batch file %s. Uploading to SumoLogic...\n", filename)
			if url != "" {
				err = uploadFileToSumoSource(filename, url)
				if err != nil {
					return err
				}
			} else {
				DebugLogger.Println("No SumoLogic receiver URL provided. Skipping upload.")
			}
			DebugLogger.Printf("Removing batch file %s\n", filename)
			err = os.Remove(filename)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
