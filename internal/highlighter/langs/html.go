package langs

type Html struct {

}

func (this *Html) Query() string {
	return `

(doctype) @tag
(tag_name) @tag
(erroneous_end_tag_name) @error
(comment) @comment
(attribute_name) @tag.attribute
(attribute
  (quoted_attribute_value) @string)

(attribute
  (quoted_attribute_value) @string)

(text) @string


((script_element
  (raw_text) @injection.content.javascript)
 (#set! injection.language "javascript"))

((style_element
  (raw_text) @injection.content.css)
 (#set! injection.language "css"))
`
}
