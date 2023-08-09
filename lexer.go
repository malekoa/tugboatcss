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
		return LexString(string(data))
	}
}

func LexString(str string) ([]string, error) {
	r := regexp.MustCompile(tokenRegex)
	return r.FindAllString(str, -1), nil
}
