package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Lang struct {
	Name     string `yaml:"name"`
	Lsp      string `yaml:"lsp"`
	Comment  string `yaml:"comment"`
	TabWidth int    `yaml:"tabwidth"`
}

type Config struct {
	Langs map[string]Lang
}

var DefaultConfig = Config { Langs:
	map[string]Lang{
		"go": { Name: "go", Lsp: "gopls", TabWidth: 4 },
		"python": { Name: "python", Lsp: "pylsp", Comment: "#", TabWidth: 4},
	},
}

var UnknownLang = Lang{ Name: "", Lsp: "", Comment: "//", TabWidth: 2 }

func GetConfig() (Config) {
	conffilename, exists := os.LookupEnv("EDGO_CONFIG")
	if !exists { conffilename = "config.yaml" }

	data, err := os.ReadFile(conffilename)
	if err != nil { return DefaultConfig }

	var languages Config
	err = yaml.Unmarshal(data, &languages)
	if err != nil { return DefaultConfig }

	// Set default TabWidth and comment if not specified
	for langName, lang := range languages.Langs {
		if lang.TabWidth == 0 { lang.TabWidth = 2; languages.Langs[langName] = lang }
		if lang.Comment == "" { lang.Comment = "//"; languages.Langs[langName] = lang }
		DefaultConfig.Langs[lang.Name] = lang
	}

	return DefaultConfig
}
