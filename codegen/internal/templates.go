package internal

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jonoans/mongo-gen/utils"
)

var templateFiles = map[string]*template.Template{}

func ReadAllTemplateFiles() {
	myPath := utils.GetMyPath()
	matches, err := filepath.Glob(filepath.Join(myPath, "*.gotmpl"))
	if err != nil {
		log.Fatalf("Error reading template files: %s", err)
	}

	for _, match := range matches {
		file, err := os.Open(match)
		if err != nil {
			file.Close()
			log.Fatalf("Error reading template files: %s", err)
		}

		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			file.Close()
			log.Fatalf("Error reading template files: %s", err)
		}

		base := utils.FilenameWoExt(match)
		templateFiles[base] = template.Must(template.New(base).Parse(string(fileContents)))
		file.Close()
	}
}

func GetTemplate(name string) *template.Template {
	return templateFiles[name]
}
