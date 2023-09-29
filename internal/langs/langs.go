package langs

import (
	. "edgo/internal/langs/bash"
	. "edgo/internal/langs/c"
	. "edgo/internal/langs/cpp"
	. "edgo/internal/langs/golang"
	. "edgo/internal/langs/html"
	. "edgo/internal/langs/javascript"
	. "edgo/internal/langs/python"
	. "edgo/internal/langs/rust"
	. "edgo/internal/langs/java"
	. "edgo/internal/langs/typescript"
	. "edgo/internal/langs/yaml"
)

type Language interface {
	Query() string
}

var languages = map[string]Language{
	"javascript": &Javascript{},
	"typescript": &Typescript{},
	"python":     &Python{},
	"rust":       &Rust{},
	"go":         &Go{},
	"c":          &C{},
	"c++":        &Cpp{},
	"cpp":        &Cpp{},
	"html":       &Html{},
	"yaml":       &Yaml{},
	"java":       &Java{},
	"bash":       &Bash{},
}

func MatchQueryLang(lang string) string {
	if l, exists := languages[lang]; exists {
		return l.Query()
	}
	return languages["javascript"].Query()
}
