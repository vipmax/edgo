package html

type Html struct {

}

func (this *Html) Query() string {
	return `
(tag_name) @tag
(erroneous_end_tag_name) @error
(comment) @comment
(attribute_name) @tag.attribute
(attribute
  (quoted_attribute_value) @string)

(attribute
  (quoted_attribute_value) @string)
(text) @text @spell
`
}
