package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jonoans/mongo-gen/utils"
	"gopkg.in/yaml.v2"
)

type ModelsConfig struct {
	PackageName  string   `yaml:"packageName,omitempty"`
	PackagePath  string   `yaml:"packagePath,omitempty"`
	IgnoredFiles []string `yaml:"ignoredFiles,omitempty"`
	PackageRoot  string
	ModuleRoot   string
}

var modReg = regexp.MustCompile(`module\s+(.*)`)

func getModPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	modPath, err := utils.RecursiveSearchDirForFile(dir, "go.mod")
	if err != nil {
		log.Fatal(err)
	}

	if modPath == "" {
		log.Fatal("Could not find go.mod file in current directory tree!")
	}

	modFileContents, err := ioutil.ReadFile(modPath)
	modFileString := string(modFileContents)
	modMatches := modReg.FindStringSubmatch(modFileString)
	if len(modMatches) == 0 {
		return ""
	}

	return modMatches[1]
}

func (mc *ModelsConfig) IsValid() bool {
	if mc.PackageName == "" {
		log.Println("Model package name not specified")
		return false
	}

	if mc.PackagePath == "" {
		mc.PackagePath = mc.PackageName
		if !utils.DirExists(mc.PackagePath) {
			log.Println("Model package path not specified")
			return false
		}
	} else if !utils.DirExists(mc.PackagePath) {
		log.Println("Model package path does not exist")
		return false
	}
	mc.PackagePath = utils.CleanPathInput(mc.PackagePath)

	mc.ModuleRoot = getModPath()
	mc.PackageRoot = filepath.Join(mc.ModuleRoot, mc.PackageName)
	mc.PackageRoot = filepath.ToSlash(mc.PackageRoot)
	return true
}

type OutputConfig struct {
	PackageName  string   `yaml:"packageName,omitempty"`
	PackagePath  string   `yaml:"packagePath,omitempty"`
	IgnoredFiles []string `yaml:"ignoredFiles,omitempty"`
	FileSuffix   string
}

func (oc *OutputConfig) IsValid() bool {
	if oc.PackageName == "" {
		log.Println("Output package name not specified")
		return false
	}

	if oc.PackagePath == "" {
		log.Println("Output package path not specified")
		return false
	}

	return true
}

type ConfigFile struct {
	Filename string       `yaml:"-"`
	Models   ModelsConfig `yaml:"models,omitempty"`
	Output   OutputConfig `yaml:"output,omitempty"`
}

func (c *ConfigFile) IsValid() bool {
	return c.Models.IsValid() && c.Output.IsValid()
}

var cfg *ConfigFile

func ParseConfig(filename string) *ConfigFile {
	if cfg != nil {
		return cfg
	}

	filename, err := utils.AbsFilePath(filename)
	if err != nil {
		log.Fatal(err)
	}

	cfgFilename := findConfigFile(filename)
	fileContents, err := ioutil.ReadFile(cfgFilename)
	if err != nil {
		log.Fatal(err)
	}

	cfg = &ConfigFile{}
	err = yaml.Unmarshal(fileContents, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if !cfg.IsValid() {
		log.Fatal("Invalid config file")
	}

	cfg.Filename = cfgFilename
	return cfg
}

const defaultConfigFilename = "orm.yml"

func findConfigFile(filename string) string {
	if filename != "" {
		if utils.FileExists(filename) {
			return filename
		}
	}

	if !utils.FileExists(defaultConfigFilename) {
		log.Fatal("Config file not found")
	}

	filename, _ = utils.AbsFilePath(defaultConfigFilename)
	return filename
}
