package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	e := echo.New()
	e.GET("/progressbar", progressBarUpdate)

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
        const evtSource = new EventSource("/progressbar");

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
		delteCheckBox := c.FormValue("deleteInputFile")

		filename := ""
		if err := os.MkdirAll(outputFolderPath, 0755); err != nil {
			fmt.Println("Error creating output folder:", err)
			return err
		}

		files, err := filepath.Glob(filepath.Join(folderPath, "*"))
		if err != nil {
			fmt.Println("Error reading folder:", err)
			return err
		}

		for _, file := range files {

			filename = convertMedia(file, outputFolderPath, delteCheckBox)
		}

		return c.String(http.StatusOK, "Video conversion completed!"+filename)
	})

	e.Start(":8080")
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
