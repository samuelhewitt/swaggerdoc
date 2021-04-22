package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/swaggo/swag"
	"gopkg.in/yaml.v2"
)

const apiDir = "api"

//go:embed dist
var efs embed.FS

func main() {
	cmdline.AppName = "Swagger Doc"
	cmdline.CopyrightYears = "2019-2021"
	cmdline.CopyrightHolder = "Richard A. Wilkes"

	cl := cmdline.New(true)
	searchDir := "."
	mainAPIFile := "main.go"
	destDir := "docs"
	baseName := "swagger"
	maxDependencyDepth := 2
	markdownFileDir := ""
	cl.NewStringOption(&searchDir).SetSingle('s').SetName("search").SetArg("dir").SetUsage("The directory root to search for documentation directives")
	cl.NewStringOption(&mainAPIFile).SetSingle('m').SetName("main").SetArg("file").SetUsage("The Go file to search for the main documentation directives")
	cl.NewStringOption(&destDir).SetSingle('o').SetName("output").SetArg("dir").SetUsage("The destination directory to write the documentation files to")
	cl.NewStringOption(&baseName).SetSingle('n').SetName("name").SetArg("name").SetUsage("The base name to use for the definition files")
	cl.NewIntOption(&maxDependencyDepth).SetSingle('d').SetName("depth").SetUsage("The maximum depth to resolve dependencies; use 0 for unlimited")
	cl.NewStringOption(&markdownFileDir).SetSingle('i').SetName("mdincludes").SetArg("dir").SetUsage("The directory root to search for markdown includes")
	cl.Parse(os.Args[1:])
	jot.FatalIfErr(generate(searchDir, mainAPIFile, destDir, baseName, markdownFileDir, maxDependencyDepth))
	atexit.Exit(0)
}

func generate(searchDir, mainAPIFile, destDir, baseName, markdownFileDir string, maxDependencyDepth int) error {
	if err := os.MkdirAll(filepath.Join(destDir, apiDir), 0755); err != nil {
		return errs.Wrap(err)
	}

	var parser *swag.Parser
	if markdownFileDir != "" {
		parser = swag.New(swag.SetMarkdownFileDirectory(markdownFileDir))
	} else {
		parser = swag.New()
	}

	parser.ParseDependency = true
	parser.ParseInternal = true
	if err := parser.ParseAPI(searchDir, mainAPIFile, maxDependencyDepth); err != nil {
		return errs.Wrap(err)
	}
	jData, err := json.Marshal(parser.GetSwagger())
	if err != nil {
		return errs.Wrap(err)
	}
	if err = ioutil.WriteFile(filepath.Join(destDir, apiDir, baseName+".json"), jData, 0644); err != nil { //nolint:gosec // Yes, I want 0644 permissions
		return errs.Wrap(err)
	}
	// Since the object that parser.GetSwagger() returned has no yaml keys
	// defined, trying to marshal it directly results in absurdly large files
	// that aren't needed. Instead, we'll use the yaml marshaller to first
	// unmarshal the json, then marshal it back to yaml.
	var j interface{}
	if err = yaml.Unmarshal(jData, &j); err != nil {
		return errs.Wrap(err)
	}
	yData, yErr := yaml.Marshal(j)
	if yErr != nil {
		return errs.Wrap(yErr)
	}
	if err = ioutil.WriteFile(filepath.Join(destDir, apiDir, baseName+".yaml"), yData, 0644); err != nil { //nolint:gosec // Yes, I want 0644 permissions
		return errs.Wrap(err)
	}

	var subFS fs.FS
	if subFS, err = fs.Sub(efs, "dist"); err != nil {
		return errs.Wrap(err)
	}

	for _, name := range []string{
		"favicon-16x16.png",
		"favicon-32x32.png",
		"index.html",
		"oauth2-redirect.html",
		"swagger-ui.css",
		"swagger-ui.js",
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
	} {
		data, err := fs.ReadFile(subFS, name)
		if err != nil {
			return errs.Newf("unable to read %s", name)
		}
		if name == "index.html" {
			data = bytes.Replace(data, []byte("./swagger."), []byte("./"+baseName+"."), 2)
			data = bytes.Replace(data, []byte(" Swagger "), []byte(" "+txt.ToCamelCase(baseName)+" "), 2)
		}
		if err = ioutil.WriteFile(filepath.Join(destDir, apiDir, name), data, 0644); err != nil { //nolint:gosec // Yes, I want 0644 permissions
			return errs.Wrap(err)
		}
	}
	return nil
}
