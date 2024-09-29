// stravauploader project main.go
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	strava "github.com/strava/go.strava"
)

var accessToken string
var client *strava.Client
var uploadService *strava.UploadsService
var athleteService *strava.CurrentAthleteService

func init() {
	flag.StringVar(&accessToken, "token", "", "user access_token from Strava")
	flag.Parse()

	if accessToken == "" {
		log.Println("\nPlease provide an access_token, one can be found at https://www.strava.com/settings/api")
		flag.Usage()
		os.Exit(1)
	}

	client = strava.NewClient(accessToken)
	uploadService = strava.NewUploadsService(client)
	athleteService = strava.NewCurrentAthleteService(client)
}

func main() {
	activitiesPath := "path/to/tcx/dir"
	activityFiles := getActivityFiles(activitiesPath)
	for _, file := range activityFiles {
		uploadData(activitiesPath, file)
	}

}

func getActivityFiles(path string) []os.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("Activity files cannot be found", err)
	}
	var fileInfos []os.FileInfo
	for _, fileInfo := range files {
		if filepath.Ext(fileInfo.Name()) == ".tcx" {
			fileInfos = append(fileInfos, fileInfo)
		}
	}
	return fileInfos
}

func uploadData(basePath string, file os.FileInfo) {
	log.Printf("Uploading file %s \n", file.Name())

	filePath := filepath.Join(basePath, file.Name())
	reader, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to read file "+filePath, err)
	}
	upload, err := uploadService.
		Create(strava.FileDataTypes.TCX, file.Name(), reader).
		Private().
		Do()
	if err != nil {
		if e, ok := err.(strava.Error); ok && e.Message == "Authorization Error" {
			log.Printf("Make sure your token has 'write' permissions. You'll need implement the oauth process to get one")
		}

		log.Fatal("Error sending file ", err)
	}

	log.Printf("Upload Complete...")
	jsonForDisplay, _ := json.Marshal(upload)
	log.Printf(string(jsonForDisplay))

	log.Printf("Waiting a 5 seconds so the upload will finish (might not)")
	time.Sleep(5 * time.Second)

	uploadSummary, err := uploadService.Get(upload.Id).Do()
	jsonForDisplay, _ = json.Marshal(uploadSummary)
	log.Printf(string(jsonForDisplay))

	log.Printf("Your new activity is id %d", uploadSummary.ActivityId)
	log.Printf("You can view it at http://www.strava.com/activities/%d", uploadSummary.ActivityId)
}
