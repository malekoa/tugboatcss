package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func getAllFilesMatchingGlobPattern(tugboatConfig *TugboatConfig, projectDirs *[]string) []string {
	var set = make(map[string]bool)
	for _, dir := range *projectDirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if FilePathMatchesGlobPattern(tugboatConfig, d.Name()) {
				path := fmt.Sprintf("%s/%s", dir, d.Name())
				_, err := os.Open(path)
				if err != nil && os.IsNotExist(err) {
					// fmt.Println("asdfadsfafsf")
				} else {
					set[path] = true
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	filesMatchingGlobPattern := make([]string, len(set))
	i := 0
	for k := range set {
		filesMatchingGlobPattern[i] = k
		i += 1
	}
	return filesMatchingGlobPattern
}

func getAllLexemes(filesMatchingGlobPattern []string) []string {
	set := make(map[string]bool)
	for _, path := range filesMatchingGlobPattern {
		lexemes, err := LexFileAtPath(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Found non-existing file: %s\n", err.Error())
			} else {
				log.Fatal(err)
			}
		}
		for _, lexeme := range lexemes {
			set[lexeme] = true
		}
	}
	lexemes := make([]string, len(set))
	i := 0
	for k := range set {
		lexemes[i] = k
		i += 1
	}
	return lexemes
}

func getAllSupportedRules(lexemes []string) []string {
	supportedRules := []string{}
	for _, lexeme := range lexemes {
		if IsSupportedRule(lexeme) {
			supportedRules = append(supportedRules, lexeme)
		}
	}
	return supportedRules
}

func buildBreakpointRule(rule []string) string {
	breakpointWidth := map[string]string{
		"sm":  "640",
		"md":  "768",
		"lg":  "1024",
		"xl":  "1280",
		"2xl": "1536",
	}
	inner := fmt.Sprintf(".%s\\:%s{%s}", rule[0], rule[1], RuleMap[rule[1]])
	return fmt.Sprintf("@media(min-width:%s){%s}\n", breakpointWidth[rule[0]], inner)
}

func buildPseudoClassRule(rule []string) string {
	pseudoClassMap := map[string]string{
		"hover":  "hover",
		"first":  "first-child",
		"last":   "last-child",
		"active": "active",
		"focus":  "focus",
	}
	return fmt.Sprintf(".%s\\:%s:%s{%s}\n", rule[0], rule[1], pseudoClassMap[rule[0]], RuleMap[rule[1]])
}

func buildDarkRule(rule []string) string {
	after, _ := strings.CutPrefix(RuleMap[rule[1]], ".")
	inner := fmt.Sprintf(".%s\\:%s", rule[0], after)
	return fmt.Sprintf("@media(prefers-color-scheme:dark){%s}\n", inner)
}

func buildOutput(supportedRules []string, ctx *cli.Context) string {
	output := "/*! normalize.css v8.0.1 | MIT License | github.com/necolas/normalize.css */button,hr,input{overflow:visible}progress,sub,sup{vertical-align:baseline}[type=checkbox],[type=radio],legend{box-sizing:border-box;padding:0}html{line-height:1.15;-webkit-text-size-adjust:100%}details,main{display:block}h1{font-size:2em;margin:.67em 0}hr{box-sizing:content-box;height:0}code,kbd,pre,samp{font-family:monospace,monospace;font-size:1em}a{background-color:transparent}abbr[title]{border-bottom:none;text-decoration:underline;text-decoration:underline dotted}b,strong{font-weight:bolder}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative}sub{bottom:-.25em}sup{top:-.5em}img{border-style:none}button,input,optgroup,select,textarea{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,select{text-transform:none}[type=button],[type=reset],[type=submit],button{-webkit-appearance:button}[type=button]::-moz-focus-inner,[type=reset]::-moz-focus-inner,[type=submit]::-moz-focus-inner,button::-moz-focus-inner{border-style:none;padding:0}[type=button]:-moz-focusring,[type=reset]:-moz-focusring,[type=submit]:-moz-focusring,button:-moz-focusring{outline:ButtonText dotted 1px}fieldset{padding:.35em .75em .625em}legend{color:inherit;display:table;max-width:100%;white-space:normal}textarea{overflow:auto}[type=number]::-webkit-inner-spin-button,[type=number]::-webkit-outer-spin-button{height:auto}[type=search]{-webkit-appearance:textfield;outline-offset:-2px}[type=search]::-webkit-search-decoration{-webkit-appearance:none}::-webkit-file-upload-button{-webkit-appearance:button;font:inherit}summary{display:list-item}body{margin:0;font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Oxygen,Ubuntu,Cantarell,'Open Sans','Helvetica Neue',sans-serif}[hidden],template{display:none}\n"
	if contents, err := os.ReadFile(ctx.String("input")); err == nil {
		output += string(contents) + "\n"
	} else {
		log.Fatalf("Found no input file '%s'\n", ctx.String("input"))
	}
	sort.Strings(supportedRules)
	for _, rule := range supportedRules {
		if r, ok := RuleMap[rule]; ok {
			output += fmt.Sprintf("%s\n", r)
		} else {
			parsed := Parse(rule)
			if len(parsed) > 2 {
				log.Printf("Skipping unsupported rule '%s'", rule)
			} else if slices.Contains([]string{"sm", "md", "lg", "xl", "2xl"}, parsed[0]) {
				output += buildBreakpointRule(parsed)
			} else if slices.Contains([]string{"hover", "first", "last", "active", "focus"}, parsed[0]) {
				output += buildPseudoClassRule(parsed)
			} else if slices.Contains([]string{"dark"}, parsed[0]) {
				output += buildDarkRule(parsed)
			}
		}
	}
	return output
}

func Generate(tugboatConfig *TugboatConfig, projectDirs *[]string, ctx *cli.Context) (string, int64) {
	startTime := time.Now()
	filesMatchingGlobPattern := getAllFilesMatchingGlobPattern(tugboatConfig, projectDirs)
	fmt.Println("filesMatchingGlobPattern: ", filesMatchingGlobPattern)
	lexemes := getAllLexemes(filesMatchingGlobPattern)
	supportedRules := getAllSupportedRules(lexemes)

	timeDiff := time.Since(startTime).Milliseconds()
	return buildOutput(supportedRules, ctx), timeDiff
}
