package main

import (
	"os"
	"regexp"
)

const (
	tokenRegex = "\\b([a-zA-Z0-9:-]+)\\b"
)

func LexFileAtPath(path string) ([]string, error) {
	if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		r := regexp.MustCompile(tokenRegex)
		return r.FindAllString(string(data), -1), nil
	}
}
