package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var jobIds = make(map[string]int)
var jobIdsMtx sync.Mutex

func main() {
	e := echo.New()

	e.GET("/progressbar", func(c echo.Context) error {

		progressBarUpdate(c)

		return c.HTML(http.StatusOK, "progressbar.html")
	})

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Video Conversion</title>
			</head>
			<body>
				<h1>Video Conversion</h1>
				<form action="/convert" method="post">
					<label for="folderPath">Folder Path:</label>
					<input type="text" id="folderPath" name="folderPath" required><br>
					<label for="outputFolderPath">Output Folder Path:</label>
					<input type="text" id="outputFolderPath" name="outputFolderPath" required><br>
					<input type="checkbox" id="deleteInputFile" name="deleteInputFile">
					<label for="deleteInputFile">Delete input file after conversion</label><br>
					<input type="submit" value="Convert">
				</form>				
			</body>
   <script>
  
    </script>

			</html>
		`)
	})

	e.POST("/convert", func(c echo.Context) error {
		//folderPath := c.FormValue("folderPath")
		outputFolderPath := c.FormValue("outputFolderPath")
		//delteCheckBox := c.FormValue("deleteInputFile")

		if err := os.MkdirAll(outputFolderPath, 0755); err != nil {
			fmt.Println("Error creating output folder:", err)
			return err
		}

		//	files, err := filepath.Glob(filepath.Join(folderPath, "*"))
		//if err != nil {
		//		fmt.Println("Error reading folder:", err)
		//		return err
		//	}
		jobNumber := spawnJob()
		return c.Redirect(http.StatusFound, "/progressbar/"+jobNumber)
	})
	// uuid above is conntect to the uuid below it gets it from the web address
	e.GET("/progressbar/:uuid", func(c echo.Context) error {

		uuid := c.Param("uuid")

		progress := getJobProgress(uuid)
		fmt.Println(progress)
		return c.String(http.StatusOK, fmt.Sprintf("Received ID: %s", uuid))
	})

	e.Start(":8080")
}

func spawnJob() string {
	jobIdNumber := genrateUUID()

	go func() {
		setJobProgress(jobIdNumber, 0)
		fmt.Println(jobIdNumber)
		time.Sleep(10000)
		setJobProgress(jobIdNumber, 100)
	}()

	return jobIdNumber
}

func setJobProgress(jobUUID string, progress int) {

	jobIdsMtx.Lock()
	jobIds[jobUUID] = progress
	jobIdsMtx.Unlock()

}

func getJobProgress(jobUUID string) int {

	jobIdsMtx.Lock()
	progress := jobIds[jobUUID]
	jobIdsMtx.Unlock()

	return progress
}

func convertMedia(file string, outputFolderPath string, delteCheckBox string) string {
	filename := ""

	if strings.HasSuffix(file, ".mp4") || strings.HasSuffix(file, ".mkv") ||
		strings.HasSuffix(file, ".avi") || strings.HasSuffix(file, ".mov") ||
		strings.HasSuffix(file, ".wmv") || strings.HasSuffix(file, ".mpeg") ||
		strings.HasSuffix(file, ".flv") || strings.HasSuffix(file, ".3gp") {
		// Construct the output filename (replace the extension with .mp4)
		outputFile := filepath.Join(outputFolderPath, strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))+".mp4")

		cmd := exec.Command("ffmpeg-2024-02-04-git-7375a6ca7b-full_build/bin/ffmpeg.exe", "-i", file, "-c:v", "libx264", "-preset", "medium", "-crf", "23", "-y", outputFile)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error converting %s: %v\n", file, err)
			return "failed to convert" + filename
		} else {
			fmt.Printf("Converted %s to %s\n", file, outputFile)
			filename = file

			return filename
		}
	}
	return "file format not supported please enter a reqest to have this format supported"
}

func genrateUUID() string {

	uuidWithHyphen := uuid.New()
	fmt.Println(uuidWithHyphen)

	// Convert UUID to a string without hyphens
	uuidString := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return uuidString
}

func progressBarUpdate(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	for i := 0; i <= 100; i++ {
		fmt.Fprintf(c.Response(), "data: %d\n\n", i)
		c.Response().Flush()
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
