package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func UploadFilesToSumo(wd string, url string) error {
	if url == "" {
		return fmt.Errorf("receiver URL is empty")
	}
	// Get list of files in the working directory and check if there are any batch files
	files, err := os.ReadDir(wd)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), batchFilenamePrefix) {
			filename := path.Join(wd, file.Name())
			DebugLogger.Println(green(fmt.Sprintf("Found batch file %s. Uploading to SumoLogic...\n", filename)))
			err = uploadFileToSumoSource(filename, url)
			if err != nil {
				return err
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
