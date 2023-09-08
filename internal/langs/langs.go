package langs

import (
	. "edgo/internal/langs/bash"
	. "edgo/internal/langs/golang"
	. "edgo/internal/langs/html"
	. "edgo/internal/langs/javascript"
	. "edgo/internal/langs/python"
	. "edgo/internal/langs/rust"
	. "edgo/internal/langs/typescript"
	. "edgo/internal/langs/yaml"
)

type Language interface {
	Query() string
}

func MatchQueryLang(lang string) string {
	var languages = map[string]Language{
		"javascript": &Javascript{},
		"typescript": &Typescript{},
		"python":     &Python{},
		"rust":       &Rust{},
		"go":         &Go{},
		"html":       &Html{},
		"yaml":       &Yaml{},
		"bash":       &Bash{},
	}

	if l, exists := languages[lang]; exists {
		return l.Query()
	}
	return languages["javascript"].Query()
}