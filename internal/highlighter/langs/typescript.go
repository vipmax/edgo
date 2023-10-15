package langs

type Typescript struct {

}

func (this *Typescript) Query() string {
	return `
[
  "async"
  "debugger"
  "delete"
  "extends"
  "from"
  "get"
  "new"
  "set"
  "target"
  "typeof"
  "instanceof"
  "void"
  "with"
] @keyword

[
  "of"
  "as"
  "in"
] @keyword.operator

[
  "function"
] @keyword.function

[
  "class"
  "let"
  "var"
] @keyword.storage.type

[
  "const"
  "static"
] @keyword.storage.modifier

[
  "default"
  "yield"
  "finally"
  "do"
  "await"
] @keyword.control

[
  "if"
  "else"
  "switch"
  "case"
  "while"
] @keyword.control.conditional


[
  (true)
  (false)
] @keyword


[
  "for"
] @keyword.control.repeat

[
  "import"
  "export"
] @keyword.control.import 

[
  "return"
  "break"
  "continue"
] @keyword.control.return

[
  "throw"
  "try"
  "catch"
] @keyword.control.exception

; Function and method definitions
;--------------------------------

(function
  name: (identifier) @function)
(function_declaration
  name: (identifier) @function)
(method_definition
  name: (property_identifier) @function.method)

(pair
  key: (property_identifier) @function.method
  value: [(function) (arrow_function)])

(assignment_expression
  left: (member_expression
    property: (property_identifier) @function.method)
  right: [(function) (arrow_function)])

(variable_declarator
  name: (identifier) @function
  value: [(function) (arrow_function)])

(assignment_expression
  left: (identifier) @function
  right: [(function) (arrow_function)])


; Function and method calls
;--------------------------

(call_expression
  function: (identifier) @function)

(call_expression
  function: (member_expression
    property: (property_identifier) @function.method)) 

[
  (string)
  (template_string)
] @string


[
  (this)
  (super)
] @variable.builtin

; Types

(type_identifier) @type
(predefined_type) @type.builtin

; ({ p }: { p: t })
(required_parameter
  (object_pattern
    (shorthand_property_identifier_pattern) @variable.parameter))

(shorthand_property_identifier) @property 

; { i }
(object_pattern
  (shorthand_property_identifier_pattern) @variable)

(variable_declarator
  name: (object_pattern
    (shorthand_property_identifier_pattern))) @variable

[
  (null)
  (undefined)
] @constant.builtin

`
}

