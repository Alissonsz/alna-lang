; Control flow statements
(if_statement "if" @keyword)
(if_statement "else" @keyword)
(while_statement "while" @keyword)
(for_statement "for" @keyword)
(return_statement "return" @keyword)
(break_statement) @keyword
(continue_statement) @keyword

; Types
(type) @type

; Functions
(function_declaration
  name: (identifier) @function)

(function_call
  function: (identifier) @function.call)

; Variables
(variable_declaration
  name: (identifier) @variable.builtin)

(parameter
  name: (identifier) @variable.parameter)

(assignment
  left: (identifier) @variable)

; Operators
(binary_expression
  operator: _ @operator)

"=" @operator
"!" @operator

; Literals
(number) @number
(boolean) @boolean

; Punctuation
"(" @punctuation.bracket
")" @punctuation.bracket
"{" @punctuation.bracket
"}" @punctuation.bracket
"," @punctuation.delimiter
";" @punctuation.delimiter

; Comments
(comment) @comment

; General identifiers (fallback)
(identifier) @variable
