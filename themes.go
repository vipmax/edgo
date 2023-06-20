package main

import "github.com/alecthomas/chroma"
import "github.com/alecthomas/chroma/styles"

var EdgoDark = styles.Register(chroma.MustNewStyle("edgo", chroma.StyleEntries{
	chroma.Comment: "#a8a8a8",
	chroma.Keyword: "#FF69B4",
	chroma.KeywordNamespace: "#FF69B4",
	chroma.String: "#90EE90",
	chroma.LiteralStringDouble: "#90EE90",
	chroma.Literal: "#90EE90",
	chroma.StringChar: "#90EE90",
	chroma.KeywordType: "#7FFFD4",
	chroma.KeywordDeclaration: "#7FFFD4",
	chroma.KeywordReserved: "#7FFFD4",
	chroma.NameTag: "#7FFFD4",
	chroma.NameFunction: "#7FFFD4",
	chroma.NumberInteger: "#00BFFF",
	chroma.NameBuiltinPseudo: "#FF69B4",
	chroma.NameFunctionMagic: "#7FFFD4",
}))

var EdgoLight = styles.Register(chroma.MustNewStyle("edgo-light", chroma.StyleEntries{
	chroma.Comment: "#a8a8a8",
	chroma.Keyword: "#FF69B4",
	chroma.KeywordNamespace: "#FF69B4",
	chroma.String: "#65aa70",
	chroma.LiteralStringDouble: "#65aa70",
	chroma.Literal: "#65aa70",
	chroma.StringChar: "#65aa70",
	chroma.KeywordType: "#60CCC0",
	chroma.KeywordDeclaration: "#60CCC0",
	chroma.KeywordReserved: "#60CCC0",
	chroma.NameTag: "#60CCC0",
	chroma.NameFunction: "#60CCC0",
	chroma.NumberInteger: "#00BFFF",
	chroma.NameFunctionMagic: "#60CCC0",
	chroma.NameBuiltinPseudo: "#FF69B4",
}))


var Darcula = styles.Register(chroma.MustNewStyle("darcula", chroma.StyleEntries{
	chroma.Comment: "#707070",
	//chroma.NameConstant: "#7A9EC2",
	chroma.Keyword: "#CC8242",
	chroma.KeywordNamespace: "#CC8242",
	chroma.String: "#6A8759",
	chroma.LiteralStringDouble: "#6A8759",
	chroma.Literal: "#6A8759",
	chroma.StringChar: "#6A8759",
	chroma.KeywordType: "#CC8242",
	chroma.KeywordDeclaration: "#CC8242",
	chroma.KeywordReserved: "#CC8242",
	//chroma.NameTag: "#FFC66D",
	//chroma.NameFunction: "#FFC66D",
	chroma.NumberInteger: "#7A9EC2",
	//chroma.NameFunction: "#FFC66D",
	//chroma.NameFunction: "#AD9E7E",
}))


var IdeaLight = styles.Register(chroma.MustNewStyle("idea-light", chroma.StyleEntries{
	chroma.Comment: "#707070",
	chroma.KeywordConstant: "#1232AC",
	chroma.Keyword: "#1132AC",
	chroma.KeywordNamespace: "#1232AC",
	chroma.String: "#65aa70",
	chroma.LiteralStringDouble: "#65aa70",
	chroma.Literal: "#65aa70",
	chroma.StringChar: "#65aa70",
	chroma.KeywordType: "#8B588A",
	chroma.KeywordDeclaration: "#1232AC",
	chroma.KeywordReserved: "#8B588A",
	chroma.NameTag: "#FFC66D",
	chroma.NumberInteger: "#284FE2",
	chroma.NameFunction: "#286077",
	chroma.NameFunctionMagic: "#A320AC",
	chroma.NameBuiltinPseudo: "#8B588A",
}))



