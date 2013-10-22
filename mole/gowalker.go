package mole

import (
	//"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type SourceInfo struct {
}

func ProcessSources(roots []string) (*SourceInfo, error) {
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println("error scan app source:", err)
				return nil
			}
			//if is normal file or name is temp skip
			//directory is needed
			if !info.IsDir() || info.Name() == "tmp" {
				return nil
			}

			tok := token.NewFileSet()
			astPkgs, err := parser.ParseDir(tok, path, func(info os.FileInfo) bool {
				name := info.Name()
				return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
			}, 0)

			if err != nil {
				//err is ErrorList
				if errList, ok := err.(scanner.ErrorList); ok {

				}

			}

			//ignore the main package
			delete(astPkgs, "main")
			//ignore the empty package
			if len(astPkgs) == 0 {
				return nil
			}

			// for name, pkg := range astPkgs {

			// }

			return nil
		})
		if err != nil {
			return nil, err
		}

	}
	return nil, nil
}
