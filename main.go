package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var jobIds = make(map[string]int)
var jobIdsMtx sync.Mutex

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("*.html")),
	}

	e.Renderer = renderer

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
				<progress id="myProgressBar" value="0" max="100"></progress>
					
			</body>
   <script>
    const progressBar = document.getElementById("myProgressBar");
        const evtSource = new EventSource("/progressbar/:uuid");

        evtSource.onmessage = (event) => {
            const value = parseInt(event.data);
            progressBar.value = value;
        };
    </script>
  
   

			</html>
		`)
	})

	e.POST("/convert", func(c echo.Context) error {
		folderPath := c.FormValue("folderPath")
		outputFolderPath := c.FormValue("outputFolderPath")
		//delteCheckBox := c.FormValue("deleteInputFile")

		if err := os.MkdirAll(outputFolderPath, 0755); err != nil {
			fmt.Println("Error creating output folder:", err)
			return err
		}

		files, err := filepath.Glob(filepath.Join(folderPath, "*"))
		if err != nil {
			fmt.Println("Error reading folder:", err)
			return err
		}
		var UUID string
		for _, file := range files {

			fmt.Println(file)
			UUID = spawnJob(file, outputFolderPath)

		}

		return c.Redirect(http.StatusFound, "/progressbar/"+UUID)
	})
	// uuid above is conntect to the uuid below it gets it from the web address
	////	e.GET("/progressbar/:uuid", func(c echo.Context) error {

	//		uuid := c.Param("uuid")

	//	progress := getJobProgress(uuid)
	//		fmt.Println(progress)

	//		fmt.Fprintf(c.Response(), "data: %d\n\n", progress)
	//		return c.String(http.StatusOK, fmt.Sprintf("Received ID: %s progress %d", uuid, progress))
	//	})
	e.GET("/progressbar/:uuid", func(c echo.Context) error {

		uuid := c.Param("uuid")

		progress := getJobProgress(uuid)
		fmt.Println(progress)
		urlForUUID := "/api/jobs/"
		m := map[string]string{"url": urlForUUID + uuid}

		return c.Render(http.StatusOK, "ProgressUUIDTemple.html", m)

	})

	e.GET("/api/jobs/:uuid", func(c echo.Context) error {

		return c.String(http.StatusOK, fmt.Sprintf("send uuid List : %s ", 0))
	})

	e.GET("/api/jobs/:uuid", func(c echo.Context) error {
		// Set appropriate headers for streaming (e.g., Content-Type, Cache-Control)
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		uuid := c.Param("uuid")
		// Continuously send chat messages to the client
		for {
			// Get new chat messages from somewhere (e.g., a channel)
			i := getJobProgress(uuid)
			message := strconv.Itoa(i)

			// Send the message to the client
			_, err := c.Response().Write([]byte("data: " + message + "\n\n"))
			if err != nil {
				// Handle any errors (e.g., client disconnect)
				break
			}
		}

		return nil
	})

	e.Start(":8080")
}

func spawnJob(file string, outputFolder string) string {
	jobIdNumber := genrateUUID()

	go func() {
		setJobProgress(jobIdNumber, 10)
		fmt.Println(jobIdNumber)
		convertMedia(file, outputFolder)
		time.Sleep(9000000)
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

func convertMedia(file string, outputFolderPath string) string {
	filename := ""

	if strings.HasSuffix(file, ".mp4") || strings.HasSuffix(file, ".mkv") ||
		strings.HasSuffix(file, ".avi") || strings.HasSuffix(file, ".mov") ||
		strings.HasSuffix(file, ".wmv") || strings.HasSuffix(file, ".mpeg") ||
		strings.HasSuffix(file, ".flv") || strings.HasSuffix(file, ".3gp") {
		// Construct the output filename (replace the extension with .mp4)
		outputFile := filepath.Join(outputFolderPath, strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))+".mp4")

		cmd := exec.Command("ffmpeg-6.1.1-essentials_build/bin/ffmpeg.exe", "-i", file, "-c:v", "libx264", "-preset", "medium", "-crf", "23", "-y", outputFile)

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
