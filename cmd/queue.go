package cmd

import (
	"fmt"
	"sync"
)

// Queue is a queue of files to upload. It makes sure that the files are uploaded
// in the order they are added as it is important for SumoLogic
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
	DebugLogger.Println(purple(fmt.Sprintf("File %s added to the queue", filename)))
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
	DebugLogger.Println(purple(fmt.Sprintf("File %s returned to the queue", filename)))
}

// Next returns the next file in the queue
func (q *Queue) Next() string {
	q.Lock()
	defer q.Unlock()
	if len(q.filesToUpload) == 0 {
		DebugLogger.Println(purple("No files in the queue"))
		return ""
	}
	file := q.filesToUpload[0]
	q.filesToUpload = q.filesToUpload[1:]
	DebugLogger.Println(purple(fmt.Sprintf("File %s taken from the queue", file)))
	return file
}

// Len returns the length of the queue
func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()
	return len(q.filesToUpload)
}
