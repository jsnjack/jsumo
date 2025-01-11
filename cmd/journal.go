package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
)

// workingDir is the directory where the application stores files
const workingDir = ".local/jsumo/"

// journalctlCmdPrefix is the prefix of the journalctl command. It is meant to produce
// logs and the cursor (last log line)
const journalctlCmdPrefix = "journalctl --output=short-iso-precise --utc --show-cursor --quiet"

// postfixAfterCursor is the postfix of the journalctl command to get logs after the cursor
const postfixAfterCursor = "--after-cursor="

// postfixSinceStart is the postfix of the journalctl command to get logs since the start of the program
const postfixSinceStart = "--since="

// defaultInterval is the default interval to get logs from journalctl
const defaultInterval = 5 * time.Second

// cursorFile is the file where the cursor is stored
const cursorFilename = "jsumo-cursor"

// batchSize is the cutoff size for logs to be sent to SumoLogic, bytes. If the size of the logs
// is greater than this, they are split into batches
// Ref: https://help.sumologic.com/docs/send-data/hosted-collectors/http-source/troubleshooting/#request-timeouts
const batchSize = 500 * 1024 // 500 KB

// initialCounter is the initial counter for the batch files
const initialCounter = 1000000

// batchFilenamePrefix is the prefix of the batch files
const batchFilenamePrefix = "batch-"

type JournalReader struct {
	startedAt     time.Time
	workingDir    string      // Working directory
	counter       int         // Used for batching
	readyToUpload chan string // Channel to signal that a batch file is ready to be uploaded
}

// getJournalctlCmd returns the journalctl command to get logs
func (j *JournalReader) getJournalctlCmd() (string, error) {
	cursorFile := path.Join(j.workingDir, cursorFilename)
	cursor, err := j.readCursorFile(cursorFile)
	if err != nil {
		// If the cursor file doesn't exist, start logs from the time the program started
		if os.IsNotExist(err) {
			return fmt.Sprintf("%s %s\"%s\"", journalctlCmdPrefix, postfixSinceStart, j.startedAt.Format("2006-01-02 15:04:05")), nil
		}
		return "", err
	}
	return fmt.Sprintf("%s %s\"%s\"", journalctlCmdPrefix, postfixAfterCursor, cursor), nil
}

// readCursorFile reads the cursor file and returns the cursor
func (j *JournalReader) readCursorFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeCursorFile writes the cursor to the cursor file
func (j *JournalReader) writeCursorFile(filename, cursor string) error {
	return os.WriteFile(filename, []byte(cursor), 0644)
}

// ReadLogs reads logs from journalctl and prepares them for sending to SumoLogic
func (j *JournalReader) ReadLogs() error {
	startedAt := time.Now()
	DebugLogger.Println("Reading logs from journalctl...")

	if !j.shouldReadNewLogs() {
		return nil
	}

	journalCmd, err := j.getJournalctlCmd()
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-c", journalCmd)
	DebugLogger.Printf("Running command: %s\n", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	DebugLogger.Printf("Logs read from journalctl, took %s, size %d\n", time.Since(startedAt), len(output))
	j.processLogs(&output)
	return nil
}

// createBatchFile creates a batch file with the logs, ready to be sent to sumologic HTTP source.
// The file represent a POST request body to the endpoint, compressed with zstd.
// Ref: https://help.sumologic.com/docs/send-data/hosted-collectors/http-source/logs-metrics/upload-logs/
func (j *JournalReader) createBatchFile(data *[]byte) error {
	startedAt := time.Now()

	j.counter++
	filename := path.Join(j.workingDir, fmt.Sprintf("%s%d.zst.jsumo", batchFilenamePrefix, j.counter))

	DebugLogger.Printf("Creating batch file %s...\n", filename)
	defer func() {
		DebugLogger.Printf("Batch file created %s, took %s\n", filename, time.Since(startedAt))
	}()

	// Create a buffer to hold the compressed data
	var compressedData bytes.Buffer

	// Create a zstd encoder
	encoder, err := zstd.NewWriter(&compressedData)
	if err != nil {
		return err
	}
	defer encoder.Close()

	// Write the data to the encoder
	_, err = encoder.Write(*data)
	if err != nil {
		return err
	}

	// Flush the encoder to ensure all data is written
	err = encoder.Close()
	if err != nil {
		return err
	}

	DebugLogger.Printf("Compression rate: %.2f\n", float64(len(*data))/float64(len(compressedData.Bytes())))

	// Write the compressed data to the file
	err = os.WriteFile(filename, compressedData.Bytes(), 0644)
	if err != nil {
		return err
	}

	// Add the file to the queue
	UploadQueue.AddFile(filename)
	return nil
}

// shouldReadNewLogs returns true if the logs should be read again. Normally it means
// that all batch files have been sent to SumoLogic
func (j *JournalReader) shouldReadNewLogs() bool {
	files, err := os.ReadDir(j.workingDir)
	if err != nil {
		Logger.Printf("Error reading directory %s: %s\n", j.workingDir, err)
		return false
	}
	found := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), batchFilenamePrefix) {
			UploadQueue.AddFile(path.Join(j.workingDir, file.Name()))
			found = true
		}
	}
	return !found

}

func (j *JournalReader) processLogs(logs *[]byte) error {
	startedAt := time.Now()
	DebugLogger.Println("Processing logs...")
	defer func() {
		j.counter = initialCounter
		DebugLogger.Printf("Logs processed, took %s\n", time.Since(startedAt))
	}()
	if len(*logs) == 0 {
		return nil
	}
	logsStr := string(*logs)
	logsSlice := strings.Split(logsStr, "\n")
	// Last line is the new line, and the second last line is the cursor
	cursorValue := strings.TrimPrefix(logsSlice[len(logsSlice)-2], "-- cursor: ")
	buffer := bytes.Buffer{}
	for _, line := range logsSlice[:len(logsSlice)-2] {
		buffer.WriteString(line + "\n")
		if buffer.Len() > batchSize {
			newBatch := buffer.Bytes()
			err := j.createBatchFile(&newBatch)
			if err != nil {
				return err
			}
			buffer.Reset()
		}
	}
	if buffer.Len() > 0 {
		newBatch := buffer.Bytes()
		err := j.createBatchFile(&newBatch)
		if err != nil {
			return err
		}
	}

	// Write the cursor to the cursor file
	cursorFile := path.Join(j.workingDir, cursorFilename)
	err := j.writeCursorFile(cursorFile, cursorValue)
	if err != nil {
		return err
	}
	return nil
}

// NewJournalReader creates a new Journal instance
func NewJournalReader() (*JournalReader, error) {
	// Create working directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := path.Join(homeDir, workingDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &JournalReader{
		startedAt:     time.Now(),
		workingDir:    dir,
		counter:       initialCounter,
		readyToUpload: make(chan string),
	}, nil
}
