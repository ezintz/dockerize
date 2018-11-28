package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/jwilder/gojq"
)

type templateConfig struct {
	noOverwrite bool
	delims      delimsFlag
}

// Context is the type passed into the template renderer
type Context struct{}

// Env is bound to the template rendering Context and returns the
// environment variables passed to the program
func (c *Context) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func parseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatalf("unable to parse url %s: %s", rawurl, err)
	}
	return u
}

func isTrue(s string) bool {
	b, err := strconv.ParseBool(strings.ToLower(s))
	if err == nil {
		return b
	}
	return false
}

func jsonQuery(jsonObj string, query string) (interface{}, error) {
	parser, err := gojq.NewStringQuery(jsonObj)
	if err != nil {
		return "", err
	}
	res, err := parser.Query(query)
	if err != nil {
		return "", err
	}
	return res, nil
}

func generateFile(cfg templateConfig, templatePath, destPath string) bool {
	tmpl := template.New(filepath.Base(templatePath)).Funcs(sprig.TxtFuncMap()).Funcs(template.FuncMap{
		"exists":    exists,
		"parseUrl":  parseURL,
		"isTrue":    isTrue,
		"jsonQuery": jsonQuery,
	})

	if cfg.delims[0] != "" {
		tmpl = tmpl.Delims(cfg.delims[0], cfg.delims[1])
	}
	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		log.Fatalf("unable to parse template: %s", err)
	}

	// Don't overwrite destination file if it exists and no-overwrite flag passed
	_, err = os.Stat(destPath)
	if err == nil && cfg.noOverwrite {
		return false
	}

	dest := os.Stdout
	if destPath != "" {
		dest, err = os.Create(destPath)
		if err != nil {
			log.Fatalf("unable to create %s", err)
		}
		defer dest.Close()
	}

	err = tmpl.ExecuteTemplate(dest, filepath.Base(templatePath), &Context{})
	if err != nil {
		log.Fatalf("template error: %s\n", err)
	}

	if fi, err := os.Stat(destPath); err == nil {
		if err := dest.Chmod(fi.Mode()); err != nil {
			log.Fatalf("unable to chmod temp file: %s\n", err)
		}
		if err := dest.Chown(int(fi.Sys().(*syscall.Stat_t).Uid), int(fi.Sys().(*syscall.Stat_t).Gid)); err != nil {
			log.Fatalf("unable to chown temp file: %s\n", err)
		}
	}

	return true
}

func generateDir(cfg templateConfig, templateDir, destDir string) bool {
	if destDir != "" {
		fiDest, err := os.Stat(destDir)
		if err != nil {
			log.Fatalf("unable to stat %s, error: %s", destDir, err)
		}
		if !fiDest.IsDir() {
			log.Fatalf("if template is a directory, dest must also be a directory (or stdout)")
		}
	}

	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		log.Fatalf("bad directory: %s, error: %s", templateDir, err)
	}

	for _, file := range files {
		switch {
		case file.IsDir():
			nextDestination := filepath.Join(destDir, file.Name())
			if err := os.Mkdir(nextDestination, file.Mode()); err != nil {
				log.Fatalf("failed to create directory: %s, error: %s", nextDestination, err)
			}
			generateDir(cfg, filepath.Join(templateDir, file.Name()), nextDestination)
		case destDir == "":
			generateFile(cfg, filepath.Join(templateDir, file.Name()), "")
		default:
			generateFile(cfg, filepath.Join(templateDir, file.Name()), filepath.Join(destDir, file.Name()))
		}
	}

	return true
}
