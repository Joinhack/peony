package peony

import (
	"errors"
	tmpl "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	NotDirError     = errors.New("Is not a directory")
	NoTemplateFound = errors.New("no template found")
)

type TemplateLoader struct {
	template *tmpl.Template //
	basePath string
}

func (t *TemplateLoader) skipDir(n string) bool {
	return strings.HasPrefix(n, ".")
}

func (t *TemplateLoader) skipFile(n string) bool {
	return strings.HasPrefix(n, ".")
}

func templateName(path string) string {
	if os.PathSeparator == '\\' {
		return strings.Replace(path, string(os.PathSeparator), ".", -1)
	}
	return path
}

func NewTemplateLoader(base string) (*TemplateLoader, error) {
	basePath, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}
	var bPathFileInfo os.FileInfo
	if bPathFileInfo, err = os.Stat(basePath); err != nil {
		return nil, err
	}
	if !bPathFileInfo.IsDir() {
		return nil, NotDirError
	}
	tl := &TemplateLoader{basePath: basePath}
	return tl, nil
}

func (t *TemplateLoader) load() error {
	basename := templateName(t.basePath)
	if t.template == nil {
		t.template = tmpl.New(basename)
	}
	template := t.template
	err := filepath.Walk(t.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if t.skipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if t.skipFile(info.Name()) {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		tmplName := templateName(path[len(t.basePath)+1:])
		_, err = template.New(tmplName).Parse(string(data))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *TemplateLoader) Lookup(path string) (*tmpl.Template, error) {
	template := t.template.Lookup(templateName(path))
	var err = NoTemplateFound
	if template != nil {
		err = nil
	}
	return template, err
}
