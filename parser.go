package main

import "strings"

func IsSupportedRule(rule string) bool {
	parsed := Parse(rule)
	if _, ok := RuleMap[parsed[len(parsed)-1]]; ok {
		return true
	}
	return false
}

func Parse(rule string) []string {
	return strings.Split(rule, ":")
}
