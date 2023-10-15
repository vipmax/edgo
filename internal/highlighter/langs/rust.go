package langs

type Rust struct {

}

func (this *Rust) Query() string {
	return `
; Identifier conventions

(identifier) @variable




; Other identifiers

(type_identifier) @type

(primitive_type) @type.builtin

(field_identifier) @field

(shorthand_field_initializer (identifier) @field)

(mod_item
  name: (identifier) @namespace)

(self) @variable.builtin

(loop_label ["'" (identifier)] @label)

; Function definitions

(function_item (identifier) @function)

(function_signature_item (identifier) @function)

(parameter (identifier) @parameter)

(closure_parameters (_) @parameter)

; Function calls

(call_expression
  function: (identifier) @function.call)

(call_expression
  function: (scoped_identifier
              (identifier) @function.call .))

(call_expression
  function: (field_expression
    field: (field_identifier) @function.call))

(generic_function
  function: (identifier) @function.call)

(generic_function
  function: (scoped_identifier
    name: (identifier) @function.call))

(generic_function
  function: (field_expression
    field: (field_identifier) @function.call))

; Macro definitions

"$" @function.macro

(metavariable) @function.macro

(macro_definition "macro_rules!" @function.macro)


; Literals

[
  (line_comment)
  (block_comment)
] @comment @spell

(boolean_literal) @boolean

(integer_literal) @number

(float_literal) @float

[
  (raw_string_literal)
  (string_literal)
] @string

(escape_sequence) @string.escape

(char_literal) @character

; Keywords

[
  "use"
  "mod"
] @keyword @include

(use_as_clause "as" @include)

[
  "default"
  "enum"
  "impl"
  "let"
  "move"
  "pub"
  "struct"
  "trait"
  "type"
  "union"
  "unsafe"
  "where"
] @keyword

[
  "async"
  "await"
] @keyword.coroutine

[
  "ref"
 (mutable_specifier)
] @keyword @type.qualifier

[
  "const"
  "static"
  "dyn"
  "extern"
] @keyword @storageclass

(lifetime ["'" (identifier)] @storageclass.lifetime)

"fn" @keyword.function

[
  "return"
  "yield"
] @keyword.return

(type_cast_expression "as" @keyword.operator)

(qualified_type "as" @keyword.operator)

(use_list (self) @namespace)

(scoped_use_list (self) @namespace)

(scoped_identifier [(crate) (super) (self)] @namespace)

(visibility_modifier [(crate) (super) (self)] @namespace)

[
  "if"
  "else"
  "match"
] @keyword @conditional 

[
  "break"
  "continue"
  "in"
  "loop"
  "while"
] @keyword @repeat

"for" @keyword
(for_expression "for" @repeat)
`
}

