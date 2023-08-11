package main

import (
	"fmt"
	"os"
	"regexp"
)

const (
	tokenRegex = "([a-zA-Z0-9:/\\-.]+)"
)

func LexFileAtPath(path string) ([]string, error) {
	if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		s, _ := LexString(string(data))
		fmt.Println("Lexed String: ", s)
		return LexString(string(data))
	}
}

func LexString(str string) ([]string, error) {
	r := regexp.MustCompile(tokenRegex)
	return r.FindAllString(str, -1), nil
}
