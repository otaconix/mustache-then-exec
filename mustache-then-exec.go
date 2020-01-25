package main

import (
	"fmt"
	arg "github.com/alexflint/go-arg"
	"github.com/cbroglie/mustache"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

type template struct {
	source      string
	regex       *regexp.Regexp
	replacement string
}

type args struct {
	AllowMissing  bool     `arg:"-a,--allow-missing" help:"whether to allow missing variables (default: false)"`
	Templates     []string `arg:"-t,--template,separate" placeholder:"TEMPLATE" help:"path to a template to be rendered"`
	GlobTemplates []string `arg:"-g,--glob-template,separate" placeholder:"GLOB" help:"glob for templates to be rendered"`
	Binary        string   `arg:"positional,required" placeholder:"BINARY" help:"the binary to run after rendering the templates"`
	Args          []string `arg:"positional" placeholder:"ARG" help:"arguments to the binary to run after rendering the templates"`
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func failErr(err error) {
	fail("%s", err)
}

func environmentAsMap() map[string]string {
	envMap := make(map[string]string)

	for _, v := range os.Environ() {
		kv := strings.SplitN(v, "=", 2)
		envMap[kv[0]] = kv[1]
	}

	return envMap
}

func renderTemplate(template template, environment map[string]string) error {
	outputFileName := template.source
	if template.regex != nil {
		outputFileName = template.regex.ReplaceAllString(template.source, template.replacement)
	}
	fmt.Printf("Rendering template: %s; output: %s\n", template.source, outputFileName)

	renderedTemplate, err := mustache.RenderFile(template.source, environment)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(template.source)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputFileName, []byte(renderedTemplate), fileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func parseArgs() args {
	var args args
	arg.MustParse(&args)

	mustache.AllowMissingVariables = args.AllowMissing

	return args
}

func parseTemplate(templateFileName string) template {
	currentString := []rune{}
	splits := []string{}
	isEscaped := false

	for _, c := range templateFileName {
		if c == '\\' {
			isEscaped = true
			currentString = append(currentString[:], c)
		} else if c == ':' {
			if isEscaped {
				currentString = append(currentString[0:len(currentString)-1], c)
			} else {
				splits = append(splits, string(currentString))
				currentString = []rune{}
			}
			isEscaped = false
		} else {
			currentString = append(currentString, c)
			isEscaped = false
		}
	}

	splits = append(splits, string(currentString))

	if len(splits) == 1 {
		return template{source: splits[0]}
	} else if len(splits) != 3 {
		fail("Template '%s' is invalid (have you escaped colons properly?)", templateFileName)
	}

	regex, err := regexp.Compile(splits[1])
	if err != nil {
		failErr(err)
	}

	return template{
		source:      splits[0],
		regex:       regex,
		replacement: splits[2],
	}
}

func main() {
	args := parseArgs()

	environment := environmentAsMap()

	templates := []template{}
	for _, templateGlob := range args.GlobTemplates {
		parsedTemplateGlob := parseTemplate(templateGlob)
		matchingTemplateFileNames, err := filepath.Glob(parsedTemplateGlob.source)
		if err != nil {
			failErr(err)
		}

		for _, matchingTemplateFileName := range matchingTemplateFileNames {
			templates = append(templates, template{
				source:      matchingTemplateFileName,
				regex:       parsedTemplateGlob.regex,
				replacement: parsedTemplateGlob.replacement,
			})
		}
	}

	for _, template := range args.Templates {
		templates = append(templates, parseTemplate(template))
	}

	for _, template := range templates {
		err := renderTemplate(template, environment)
		if err != nil {
			failErr(err)
		}
	}

	argv := append([]string{args.Binary}, args.Args...)
	err := syscall.Exec(args.Binary, argv, os.Environ())
	if err != nil {
		fail("Error running %s: %s", argv, err)
	}
}
