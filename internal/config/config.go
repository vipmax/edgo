package config

import (
	"gopkg.in/yaml.v3"
	"os"
)


type Lang struct {
	Name     string `yaml:"name,omitempty"`
	Lsp      string `yaml:"lsp,omitempty"`
	Comment  string `yaml:"comment,omitempty"`
	TabWidth int    `yaml:"tabwidth,omitempty"`
	Cmd      string `yaml:"cmd,omitempty"`
	CmdArgs  string `yaml:"cmdargs,omitempty"`
}


type Config struct {
	Langs map[string]Lang `yaml:"langs"`
	Theme string          `yaml:"theme"`
}

var DefaultConfig = Config { Langs:
	map[string]Lang{
		"go":         { Lsp: "gopls", TabWidth: 4, Cmd: "go", CmdArgs: "run" },
		"python":     { Lsp: "pylsp", Comment: "#", TabWidth: 4, Cmd: "python3" },
		"typescript": { Lsp: "typescript-language-server --stdio", Cmd: "tsx" },
		"javascript": { Lsp: "typescript-language-server --stdio", Cmd: "tsx" },
		"html":       { Lsp: "vscode-html-language-server --stdio" },
		"vue":        { Lsp: "vscode-html-language-server --stdio" },
		"rust":       { Lsp: "rust-analyzer", TabWidth: 4},
		"c":          { Lsp: "clangd" },
		"c++":        { Lsp: "clangd" },
		"d":          { Lsp: "serve-d", Cmd: "dmd", CmdArgs: "-run"},
		"java":       { Lsp: "jdtls", TabWidth: 4, Cmd: "java" },
		"swift":      { Lsp: "xcrun sourcekit-lsp", Cmd: "swift" },
		"haskell":    { Lsp: "haskell-language-server-wrapper --lsp", Comment: "--" },
		"zig":        { Lsp: "zls", TabWidth: 4, Cmd: "zig", CmdArgs: "run" },
		"lua":        { Lsp: "lua-language-server", Cmd: "lua" },
		"yaml":       { Comment: "#", TabWidth: 4 },
		"ocaml":      { Lsp: "ocamllsp", },
		"nim":        { Lsp: "nimlangserver", },
		"bash":       { Lsp: "bash-language-server start", Cmd: "bash", Comment: "#", TabWidth: 2 },
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

	DefaultConfig.Theme = "edgo"

	conffilename, exists := os.LookupEnv("EDGO_CONF")
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
