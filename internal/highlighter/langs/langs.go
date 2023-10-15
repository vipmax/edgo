package langs

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
