package cmd

import (
	"os"
	"path"
	"strings"
)

func UploadFilesToSumo(wd string, url string) error {
	// Get list of files in the working directory and check if there are any batch files
	files, err := os.ReadDir(wd)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), batchFilenamePrefix) {
			DebugLogger.Printf("Found batch file %s. Uploading to SumoLogic...\n", file.Name())
			if url != "" {
				err = uploadFileToSumoSource(file.Name(), url)
				if err != nil {
					return err
				}
			} else {
				DebugLogger.Println("No SumoLogic receiver URL provided. Skipping upload.")
			}
			DebugLogger.Printf("Removing batch file %s\n", file.Name())
			err = os.Remove(path.Join(wd, file.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
