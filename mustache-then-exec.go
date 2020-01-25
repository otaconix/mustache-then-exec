package main

import (
	"fmt"
	arg "github.com/alexflint/go-arg"
	"github.com/cbroglie/mustache"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

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

func renderTemplate(template string, environment map[string]string) error {
	fmt.Printf("Filling template: %s\n", template)

	renderedTemplate, err := mustache.RenderFile(template, environment)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(template, []byte(renderedTemplate), 0)
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

func main() {
	args := parseArgs()

	environment := environmentAsMap()

	matchedFiles := []string{}
	for _, templateGlob := range args.GlobTemplates {
		templates, err := filepath.Glob(templateGlob)
		if err != nil {
			failErr(err)
		}

		matchedFiles = append(matchedFiles, templates...)
	}

	for _, template := range append(args.Templates, matchedFiles...) {
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
