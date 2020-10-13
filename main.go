package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/richardwilkes/swaggerdoc/swag"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs/embedded"
	"gopkg.in/yaml.v2"
)

const apiDir = "api"

func main() {
	cmdline.AppName = "Swagger Doc"
	cmdline.CopyrightYears = "2019-2020"
	cmdline.CopyrightHolder = "Richard A. Wilkes"

	efs, err := embedded.NewEFSFromEmbeddedZip()
	if err != nil {
		fmt.Printf("embedded file system is not present, please rebuild %s\n", cmdline.AppCmdName)
		atexit.Exit(1)
	}

	cl := cmdline.New(true)
	searchDir := "."
	mainAPIFile := "main.go"
	destDir := "docs"
	baseName := "swagger"
	cl.NewStringOption(&searchDir).SetSingle('s').SetName("search").SetArg("dir").SetUsage("The directory root to search for documentation directives")
	cl.NewStringOption(&mainAPIFile).SetSingle('m').SetName("main").SetArg("file").SetUsage("The Go file to search for the main documentation directives")
	cl.NewStringOption(&destDir).SetSingle('d').SetName("dest").SetArg("dir").SetUsage("The destination directory to write the documentation files to")
	cl.NewStringOption(&baseName).SetSingle('n').SetName("name").SetArg("name").SetUsage("The base name to use for the definition files")
	cl.Parse(os.Args[1:])
	jot.FatalIfErr(generate(efs, searchDir, mainAPIFile, destDir, baseName))
	atexit.Exit(0)
}

func generate(efs *embedded.EFS, searchDir, mainAPIFile, destDir, baseName string) error {
	if err := os.MkdirAll(filepath.Join(destDir, apiDir), 0755); err != nil {
		return errs.Wrap(err)
	}
	parser := swag.NewParser()
	if err := parser.ParseAPI(searchDir, mainAPIFile, 0); err != nil {
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
	fs := efs.PrimaryFileSystem()
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
		data, exists := fs.ContentAsBytes(name)
		if !exists {
			return errs.Newf("unable to locate %s", name)
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
