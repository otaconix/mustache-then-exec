package main

import (
	"flag"
	"fmt"
	"github.com/cbroglie/mustache"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

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

func splitArgs(args []string) ([]string, []string) {
	for i, v := range args {
		if v == "--" {
			return args[:i], args[i+1:]
		}
	}

	return args[:], []string{}
}

func parseArgs() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  %s [--allow-missing] [TEMPLATE...] -- BINARY [ARGUMENTS...]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(),
			"\n  Replaces TEMPLATE files (which are Go globs) with their contents\n"+
				"  rendered as a Mustache template, where the environment variables are\n"+
				"  passed as data to those tempales.\n\n")
		flag.PrintDefaults()
	}
	allowMissing := flag.Bool("allow-missing", false, "Whether to allow missing variables (default: false).")

	flag.Parse()

	mustache.AllowMissingVariables = *allowMissing
}

func main() {
	parseArgs()

	templateGlobs, execArgs := splitArgs(flag.Args())

	if len(execArgs) == 0 {
		fail("No binary to execute provided")
	}

	environment := environmentAsMap()

	for _, templateGlob := range templateGlobs {
		templates, err := filepath.Glob(templateGlob)
		if err != nil {
			failErr(err)
		}
		for _, template := range templates {
			err := renderTemplate(template, environment)
			if err != nil {
				failErr(err)
			}
		}
	}

	if len(execArgs) > 0 {
		err := syscall.Exec(execArgs[0], execArgs, os.Environ())
		if err != nil {
			fail("Error running %s: %s", execArgs[0], err)
		}
	}
}
