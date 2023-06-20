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
	Langs map[string]Lang `yaml:"lang"`
	Theme string          `yaml:"theme"`
}

var DefaultConfig = Config { Langs:
	map[string]Lang{
		"go":         { Lsp: "gopls", TabWidth: 4 },
		//"python":     { Lsp: "pyright-langserver --stdio", Comment: "#", TabWidth: 4 },
		"python":     { Lsp: "pylsp", Comment: "#", TabWidth: 4 },
		"typescript": { Lsp: "typescript-language-server --stdio" },
		"javascript": { Lsp: "typescript-language-server --stdio" },
		"html":       { Lsp: "vscode-html-language-server --stdio" },
		"vue":        { Lsp: "vscode-html-language-server --stdio" },
		"rust":       { Lsp: "rust-analyzer", TabWidth: 4 },
		"c":          { Lsp: "clangd" },
		"c++":        { Lsp: "clangd" },
		"java":       { Lsp: "jdtls", TabWidth: 4 },
		"swift":      { Lsp: "xcrun sourcekit-lsp" },
		"haskell":    { Lsp: "haskell-language-server-wrapper --lsp", Comment: "--" },
		"zig":        { Lsp: "zls", TabWidth: 4 },
		"yaml":       { Comment: "#", TabWidth: 4 },
	},
}

var DefaultLangConfig = Lang{ Name: "", Lsp: "", Comment: "//", TabWidth: 2 }

func GetConfig() Config {

	// override default config
	for langName, langConf := range DefaultConfig.Langs {
		//set default tab width and comment if not specified
		if langConf.TabWidth == 0 { langConf.TabWidth = 2 }
		if langConf.Comment == "" { langConf.Comment = "//" }
		DefaultConfig.Langs[langName] = langConf
	}

	DefaultConfig.Theme = "edgo-light"

	conffilename, exists := os.LookupEnv("EDGO_CONFIG")
	if !exists { conffilename = "config.yaml" }

	data, err := os.ReadFile(conffilename)
	if err != nil { return DefaultConfig }

	var yamlConfig Config
	err = yaml.Unmarshal(data, &yamlConfig)
	if err != nil { return DefaultConfig }

	// read yaml config and override
	for langName, langConf := range yamlConfig.Langs {
		//set default tab width and comment if not specified
		if langConf.TabWidth == 0 { langConf.TabWidth = 2 }
		if langConf.Comment == "" { langConf.Comment = "//" }
		DefaultConfig.Langs[langName] = langConf
	}

	if yamlConfig.Theme != "" { DefaultConfig.Theme = yamlConfig.Theme }

	return DefaultConfig
}