package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log" //nolint:depguard
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs/embedded"
	"github.com/swaggo/swag"

	"gopkg.in/yaml.v2"
)

const apiDir = "api"

func main() {
	cmdline.AppName = "Swagger Doc"
	cmdline.CopyrightYears = "2019"
	cmdline.CopyrightHolder = "Richard A. Wilkes"
	cmdline.AppVersion = "1.0"
	cl := cmdline.New(true)
	searchDir := "."
	mainAPIFile := "main.go"
	destDir := "docs"
	cl.NewStringOption(&searchDir).SetSingle('s').SetName("search").SetArg("dir").SetUsage("The directory root to search for documentation directives")
	cl.NewStringOption(&mainAPIFile).SetSingle('m').SetName("main").SetArg("file").SetUsage("The Go file to search for the main documentation directives")
	cl.NewStringOption(&destDir).SetSingle('d').SetName("dest").SetArg("dir").SetUsage("The destination directory to write the documentation files to")
	cl.Parse(os.Args[1:])
	jot.FatalIfErr(generate(searchDir, mainAPIFile, destDir))
	atexit.Exit(0)
}

func generate(searchDir, mainAPIFile, destDir string) error {
	if err := os.MkdirAll(filepath.Join(destDir, apiDir), 0755); err != nil {
		return errs.Wrap(err)
	}
	log.SetOutput(ioutil.Discard) // Disable console output from the library that we don't want
	parser := swag.New()
	if err := parser.ParseAPI(searchDir, mainAPIFile); err != nil {
		return errs.Wrap(err)
	}
	jData, err := json.Marshal(parser.GetSwagger())
	if err != nil {
		return errs.Wrap(err)
	}
	if err = ioutil.WriteFile(filepath.Join(destDir, "swagger.json"), jData, 0644); err != nil {
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
	if err = ioutil.WriteFile(filepath.Join(destDir, "swagger.yaml"), yData, 0644); err != nil {
		return errs.Wrap(err)
	}
	efs, efsErr := embedded.NewEFSFromEmbeddedZip()
	if efsErr != nil {
		return errs.Wrap(efsErr)
	}
	fs := efs.PrimaryFileSystem()
	for _, name := range []string{
		"favicon-16x16.png",
		"favicon-32x32.png",
		"index_tmpl.html",
		"oauth2-redirect.html",
		"swagger-ui.css",
		"swagger-ui.js",
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
	} {
		data, exists := fs.ContentAsBytes(name)
		if !exists {
			return errs.Newf("unable to locate %s", name)
		}
		if err = ioutil.WriteFile(filepath.Join(destDir, apiDir, name), data, 0644); err != nil {
			return errs.Wrap(err)
		}
	}
	indexFile, exists := fs.ContentAsBytes("index_tmpl.html")
	if !exists {
		return errs.New("unable to locate index.html")
	}
	if err = ioutil.WriteFile(filepath.Join(destDir, apiDir, "index.html"), bytes.Replace(indexFile, []byte("SPEC"), jData, 1), 0644); err != nil {
		return errs.Wrap(err)
	}
	return nil
}
