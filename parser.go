package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func BuildRuleMap(ruleFilePath string) map[string]string {
	ruleFile, err := os.Open(ruleFilePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error opening file at path %s -", ruleFilePath), err)
	}
	csvReader := csv.NewReader(ruleFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Error reading csv file -", err)
	}
	ruleMap := make(map[string]string)
	for _, record := range records {
		ruleMap[record[0]] = record[1]
	}
	return ruleMap
}

func Parse(rule string) {
	for key, value := range RuleMap {
		fmt.Println(key, value)
	}
}
