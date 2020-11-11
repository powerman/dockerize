package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jwilder/gojq"
)

var errNotADirectory = errors.New("not a directory")

type templateConfig struct {
	noOverwrite bool
	strict      bool
	delims      delimsFlag
	data        struct {
		Env map[string]string
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

func isTrue(s string) bool {
	b, _ := strconv.ParseBool(strings.ToLower(s))
	return b
}

func jsonQuery(jsonObj string, query string) (interface{}, error) {
	parser, err := gojq.NewStringQuery(jsonObj)
	if err != nil {
		return nil, err
	}
	return parser.Query(query)
}

func readFile(fileName string) (string, error) {
	data, err := ioutil.ReadFile(fileName) //nolint:gosec // File inclusion via variable.
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}

func processTemplatePaths(cfg templateConfig, paths []string) error {
	for _, srcdst := range paths {
		var src, dst string
		switch parts := strings.SplitN(srcdst, ":", 2); len(parts) {
		case 1:
			src = parts[0]
		case 2: //nolint:gomnd // TODO Refactor?
			src, dst = parts[0], parts[1]
		}

		fi, err := os.Stat(src)
		if err == nil {
			if fi.IsDir() {
				err = processTemplateDir(cfg, src, dst)
			} else {
				err = processTemplate(cfg, src, dst)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func processTemplate(cfg templateConfig, src, dst string) error {
	option := "missingkey=default"
	if cfg.strict {
		option = "missingkey=error"
	}
	tmpl, err := template.New(filepath.Base(src)).
		Funcs(sprig.TxtFuncMap()).
		Funcs(template.FuncMap{
			"exists":    exists,
			"parseUrl":  url.Parse,
			"isTrue":    isTrue,
			"jsonQuery": jsonQuery,
			"readFile":  readFile,
		}).
		Delims(cfg.delims[0], cfg.delims[1]).
		Option(option).
		ParseFiles(src)
	if err != nil {
		return err
	}

	file := os.Stdout
	if dst != "" {
		file, err = createDestFile(src, dst, cfg.noOverwrite)
		if err != nil {
			return err
		}
		defer warnIfFail(file.Close)
	}

	return tmpl.Execute(file, cfg.data)
}

func processTemplateDir(cfg templateConfig, src, dst string) error {
	if dst != "" {
		err := ensureDestDir(src, dst)
		if err != nil {
			return err
		}
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		nextSrc := filepath.Join(src, entry.Name())
		nextDst := filepath.Join(dst, entry.Name())
		if dst == "" {
			nextDst = ""
		}
		if entry.IsDir() {
			err = processTemplateDir(cfg, nextSrc, nextDst)
		} else {
			err = processTemplate(cfg, nextSrc, nextDst)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func createDestFile(src, dst string, noOverwrite bool) (*os.File, error) {
	like, err := os.Stat(src)
	if err != nil {
		return nil, err
	}
	likeSys, ok := like.Sys().(*syscall.Stat_t)

	openFlags := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if noOverwrite {
		openFlags = os.O_RDWR | os.O_CREATE | os.O_EXCL
	}

	file, err := os.OpenFile(dst, openFlags, like.Mode().Perm()) //nolint:gosec // File inclusion.
	if err != nil {
		return nil, err
	}
	if ok {
		err = file.Chown(int(likeSys.Uid), int(likeSys.Gid))
		if err != nil && !os.IsPermission(err) {
			warnIfFail(file.Close)
			return nil, err
		}
	}
	return file, nil
}

func ensureDestDir(src, dst string) error {
	like, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !like.IsDir() {
		return fmt.Errorf("%w: %s", errNotADirectory, src)
	}
	likeSys, ok := like.Sys().(*syscall.Stat_t)

	fi, err := os.Stat(dst)
	switch {
	case err == nil && fi.IsDir():
		return nil
	case err == nil:
		return fmt.Errorf("%w: %s", errNotADirectory, dst)
	case !os.IsNotExist(err):
		return err
	}

	err = os.Mkdir(dst, like.Mode())
	if err == nil && ok {
		err = os.Chown(dst, int(likeSys.Uid), int(likeSys.Gid))
		if os.IsPermission(err) {
			err = nil
		}
	}
	return err
}
