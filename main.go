package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/urfave/cli/v2"
)

type TugboatConfig struct {
	GlobPattern []string `json:"globPattern,omitempty"`
	Ignore      []string `json:"ignore,omitempty"`
}

func getConfig() *TugboatConfig {
	file, err := os.ReadFile("./tugboat.config.json")
	if err != nil {
		log.Fatal(err)
	}
	var result TugboatConfig
	json.Unmarshal(file, &result)
	return &result
}

func getAllProjectDirectories(tugboatConfig *TugboatConfig) []string {
	projectDirs := []string{"."}
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		for _, ignorePath := range tugboatConfig.Ignore {
			if strings.Contains(path, ignorePath) {
				return nil
			}
		}
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

func FilePathMatchesGlobPattern(tugboatConfig *TugboatConfig, filePath string) bool {
	matches := true
	for _, globPattern := range tugboatConfig.GlobPattern {
		g := glob.MustCompile(globPattern)
		if !g.Match(filePath) {
			matches = false
			break
		}
	}
	return matches
}

func resetProjectDirectories(watcher *fsnotify.Watcher, projectDirs *[]string, tugboatConfig *TugboatConfig) error {
	for _, path := range *projectDirs {
		watcher.Remove(path)
	}
	updatedProjectDirectories := getAllProjectDirectories(tugboatConfig)
	for _, path := range getAllProjectDirectories(tugboatConfig) {
		watcher.Add(path)
	}
	*projectDirs = updatedProjectDirectories
	return nil
}

func eventLoop(watcher *fsnotify.Watcher, tugboatConfig *TugboatConfig, projectDirs *[]string, ctx *cli.Context) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// log.Println("event: ", event)
			if event.Has(fsnotify.Write) {
				if FilePathMatchesGlobPattern(tugboatConfig, event.Name) {
					output, timeDiff := Generate(tugboatConfig, projectDirs, ctx)
					os.WriteFile(ctx.String("output"), []byte(output), 0644)
					log.Printf("Modified watched file '%s' - Updated output in %d ms", event.Name, timeDiff)
				}
			}
			if event.Has(fsnotify.Create) {
				// A new directory may have been added. Should remove all current
				// watched directories and get all new project directories
				resetProjectDirectories(watcher, projectDirs, tugboatConfig)
				fmt.Println("Detected filesystem change, watching directories: ", *projectDirs)
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

	go eventLoop(watcher, tugboatConfig, &projectDirs, ctx)

	projectDirs = getAllProjectDirectories(tugboatConfig)
	fmt.Println("Watching directories: ", projectDirs)
	output, timeDiff := Generate(tugboatConfig, &projectDirs, ctx)
	os.WriteFile(ctx.String("output"), []byte(output), 0644)
	log.Printf("Generated initial output in %d ms\n", timeDiff)
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
					&cli.StringFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Value:   "",
					},
				},
			},
			{
				Name:  "init",
				Usage: "creates the default config file",
				Action: func(ctx *cli.Context) error {
					defaultPath := "./tugboat.config.json"
					_, err := os.Stat(defaultPath)
					if err == nil {
						fmt.Println("config file already exists")
					} else if os.IsNotExist(err) {
						defaultConfig := []byte("{\n\t\"globPattern\": [\"*.html\"],\n\t\"ignore\": [\"node_modules\"]\n}")
						err := os.WriteFile(defaultPath, defaultConfig, 0644)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println("Created default config file: ")
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
