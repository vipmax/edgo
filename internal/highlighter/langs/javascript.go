package langs

type Javascript struct {

}

func (this *Javascript) Query() string {
query :=`
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


`
	return query
}
