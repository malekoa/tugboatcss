package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/urfave/cli/v2"
)

type TugboatConfig struct {
	Content []string `json:"content,omitempty"`
}

func getConfig() *TugboatConfig {
	file, err := os.ReadFile("./tugboatcss.config.json")
	if err != nil {
		log.Fatal(err)
	}
	var result TugboatConfig
	json.Unmarshal(file, &result)
	return &result
}

func getAllProjectDirectories() []string {
	projectDirs := []string{"./"}
	err := filepath.WalkDir("./", func(path string, d fs.DirEntry, err error) error {
		if path[0] != 46 && d.IsDir() {
			projectDirs = append(projectDirs, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return projectDirs
}

func filePathMatchesGlobPattern(tugboatConfig *TugboatConfig, filePath string) bool {
	matches := true
	for _, globPattern := range tugboatConfig.Content {
		g := glob.MustCompile(globPattern)
		if !g.Match(filePath) {
			matches = false
			break
		}
	}
	return matches
}

func resetProjectDirectories(watcher *fsnotify.Watcher, projectDirs *[]string) error {
	for _, path := range *projectDirs {
		watcher.Remove(path)
	}
	updatedProjectDirectories := getAllProjectDirectories()
	for _, path := range getAllProjectDirectories() {
		watcher.Add(path)
	}
	*projectDirs = updatedProjectDirectories
	return nil
}

func eventLoop(watcher *fsnotify.Watcher, tugboatConfig *TugboatConfig, projectDirs *[]string) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// log.Println("event: ", event)
			if event.Has(fsnotify.Write) {
				if filePathMatchesGlobPattern(tugboatConfig, event.Name) {
					log.Printf("Modified watched file '%s' - Updating output...", event.Name)
					// TODO: build output using files that match globPattern in projectdirs
				}
			}
			if event.Has(fsnotify.Create) {
				// A new directory may have been added. Should remove all current
				// watched directories and get all new project directories
				resetProjectDirectories(watcher, projectDirs)
				fmt.Println("Detected filesystem change, now watching directories: ", *projectDirs)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error: ", err)
		}
	}
}

func watch(ctx *cli.Context) error {
	var projectDirs []string
	tugboatConfig := getConfig()
	// find all files in project directory that match the glob pattern
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher: ", err)
	}
	defer watcher.Close()

	go eventLoop(watcher, tugboatConfig, &projectDirs)

	projectDirs = getAllProjectDirectories()
	fmt.Println("Watching directories: ", projectDirs)
	for _, path := range projectDirs {
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-make(chan struct{})
	return nil
}

func defaultAction(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		fmt.Println(cli.ShowAppHelp(ctx))
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:   "tugboatcss",
		Usage:  "A CSS utility class generator",
		Action: defaultAction,
		Commands: []*cli.Command{
			{
				Name:   "watch",
				Usage:  "watch project files for changes",
				Action: watch,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "./out.css",
					},
				},
			},
			{
				Name:  "init",
				Usage: "creates the default config file",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
