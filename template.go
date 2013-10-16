package peony

import (
	tmpl "html/template"
	"os"
	"path/filepath"
	"strings"
)

type TemplateLoader struct {
	template *tmpl.Template //
	basePath string
}

func NewTemplateLoader(path string) (*TemplateLoader, error) {
	if path == "." {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	tl := &TemplateLoader{basePath: path}
	return tl, nil
}

func (t *TemplateLoader) load() error {
	err := filepath.Walk(t.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := strings.Replace(path[len(t.basePath)+1:], string(os.PathSeparator), ".", -1)
		println(name)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
