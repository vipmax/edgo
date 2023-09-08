package python


type Python struct {

}

func (this *Python) Query() string {
	return `

;(identifier) @identifier

[
  "def"
  "lambda"
] @keyword.function

[
  "as"
  "assert"
  "await"
  "from"
  "pass"
  "with"
] @keyword.control

[
  "if"
  "elif"
  "else"
  "match"
  "case"
] @keyword.control.conditional

[
  "while"
  "for"
  "break"
  "continue"
] @keyword.control.repeat

[
  "return"
  "yield"
] @keyword.control.return
(yield "from" @keyword.control.return)

[
  (true)
  (false)
] @constant.builtin.boolean

[
  "raise"
  "try"
  "except"
  "finally"
] @keyword.control.except
(raise_statement "from" @keyword.control.except)
"import" @keyword.control.import

(for_statement "in" @keyword.control)
(for_in_clause "in" @keyword.control)

[
  "and"
  "async"
  "class"
  "exec"
  "global"
  "nonlocal"
  ;"print"
] @keyword

[
  "and"
  "or"
  "in"
  "not"
  "del"
  "is"
] @keyword.operator

(none) @constant
(integer) @constant
(float) @constant
(comment) @comment
;(string) @string


; Function calls

(call
  function: (identifier) @function.call)

(call
  function: (attribute
              attribute: (identifier) @method.call))

; Function definitions

(function_definition
 name: (identifier) @function)

(type (identifier) @type)
(type
 (subscript
   (identifier) @type)) ; type subscript: Tuple[int]

((call
function: (identifier) @_isinstance
arguments: (argument_list
  (_)
  (identifier) @type))
(#eq? @_isinstance "isinstance"))

;; Class definitions

(class_definition name: (identifier) @type)

(class_definition
  body: (block
          (function_definition
            name: (identifier) @method)))
`
	return `

;(identifier) @identifier

[
  "def"
  "lambda"
] @keyword.function

[
  "as"
  "assert"
  "await"
  "from"
  "pass"
  "with"
] @keyword.control

[
  "if"
  "elif"
  "else"
  "match"
  "case"
] @keyword.control.conditional

[
  "while"
  "for"
  "break"
  "continue"
] @keyword.control.repeat

[
  "return"
  "yield"
] @keyword.control.return
(yield "from" @keyword.control.return)

[
  "raise"
  "try"
  "except"
  "finally"
] @keyword.control.except
(raise_statement "from" @keyword.control.except)
"import" @keyword.control.import

(for_statement "in" @keyword.control)
(for_in_clause "in" @keyword.control)

[
  "and"
  "async"
  "class"
  "exec"
  "global"
  "nonlocal"
 ; "print"
] @keyword
[
  "and"
  "or"
  "in"
  "not"
  "del"
  "is"
] @keyword.operator

(call
  function: (identifier) @function)

(function_definition
  name: (identifier) @constructor
 (#match? @constructor "^(__new__|__init__)$"))

(integer) @constant
(float) @constant
(comment) @comment
(string) @string

;; Class definitions

(class_definition name: (identifier) @type)

(class_definition
  body: (block
          (function_definition
            name: (identifier) @method)))

(class_definition
  superclasses: (argument_list
                  (identifier) @type))

((class_definition
  body: (block
          (expression_statement
            (assignment
              left: (identifier) @field))))
 (#lua-match? @field "^%l.*$"))

((class_definition
  body: (block
          (expression_statement
            (assignment
              left: (_
                     (identifier) @field)))))
 (#lua-match? @field "^%l.*$"))

((class_definition
  (block
    (function_definition
      name: (identifier) @constructor)))
 (#any-of? @constructor "__new__" "__init__"))

; Function calls

(call
  function: (identifier) @function.call)

(call
  function: (attribute
              attribute: (identifier) @method.call))

((call
   function: (identifier) @constructor)
 (#lua-match? @constructor "^%u"))

((call
  function: (attribute
              attribute: (identifier) @constructor))
 (#lua-match? @constructor "^%u"))

[
  (true)
  (false)
] @constant.builtin.boolean


;; Function definitions

(function_definition
  name: (identifier) @function)

(type (identifier) @type)
(type
  (subscript
    (identifier) @type)) ; type subscript: Tuple[int]

((call
  function: (identifier) @_isinstance
  arguments: (argument_list
    (_)
    (identifier) @type))
 (#eq? @_isinstance "isinstance"))


;; Class definitions

(class_definition name: (identifier) @type)

(class_definition
 body: (block
         (function_definition
           name: (identifier) @method)))

(class_definition
 superclasses: (argument_list
                 (identifier) @type))

((class_definition
 body: (block
         (expression_statement
           (assignment
             left: (identifier) @field))))
(#lua-match? @field "^%l.*$"))
((class_definition
 body: (block
         (expression_statement
           (assignment
             left: (_
                    (identifier) @field)))))
(#lua-match? @field "^%l.*$"))

((class_definition
 (block
   (function_definition
     name: (identifier) @constructor)))
(#any-of? @constructor "__new__" "__init__"))

; Function calls

(call
 function: (identifier) @function.call)

(call
 function: (attribute
             attribute: (identifier) @method.call))

((call
  function: (identifier) @constructor)
(#lua-match? @constructor "^%u"))

((call
 function: (attribute
             attribute: (identifier) @constructor))
(#lua-match? @constructor "^%u"))



;; Function definitions

(function_definition
 name: (identifier) @function)

(type (identifier) @type)
(type
 (subscript
   (identifier) @type)) ; type subscript: Tuple[int]

((call
function: (identifier) @_isinstance
arguments: (argument_list
  (_)
  (identifier) @type))
(#eq? @_isinstance "isinstance"))
`

//
//
//	return `
//;; Class definitions
//
//(class_definition name: (identifier) @type)
//
//(class_definition
// body: (block
//         (function_definition
//           name: (identifier) @method)))
//
//(class_definition
// superclasses: (argument_list
//                 (identifier) @type))
//
//((class_definition
// body: (block
//         (expression_statement
//           (assignment
//             left: (identifier) @field))))
//(#lua-match? @field "^%l.*$"))
//((class_definition
// body: (block
//         (expression_statement
//           (assignment
//             left: (_
//                    (identifier) @field)))))
//(#lua-match? @field "^%l.*$"))
//
//((class_definition
// (block
//   (function_definition
//     name: (identifier) @constructor)))
//(#any-of? @constructor "__new__" "__init__"))
//
//; Function calls
//
//(call
// function: (identifier) @function.call)
//
//(call
// function: (attribute
//             attribute: (identifier) @method.call))
//
//((call
//  function: (identifier) @constructor)
//(#lua-match? @constructor "^%u"))
//
//((call
// function: (attribute
//             attribute: (identifier) @constructor))
//(#lua-match? @constructor "^%u"))
//
//
//
//;; Function definitions
//
//(function_definition
// name: (identifier) @function)
//
//(type (identifier) @type)
//(type
// (subscript
//   (identifier) @type)) ; type subscript: Tuple[int]
//
//((call
// function: (identifier) @_isinstance
// arguments: (argument_list
//   (_)
//   (identifier) @type))
//(#eq? @_isinstance "isinstance"))
//
//`
}

