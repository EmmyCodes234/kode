;; TypeScript/TSX tree-sitter query — extract imports, function defs, class defs
;; Used by the multi-language verification gates (syntax, imports, calls)

;; Import declarations
(import_statement
  source: (string) @import.source) @import

;; Named imports  
(import_clause
  (named_imports
    (import_specifier
      name: (identifier) @import.name)))

;; Default imports
(import_clause
  (identifier) @import.default)

;; Dynamic imports
(call_expression
  function: (import) 
  arguments: (arguments (string) @import.dynamic))

;; Require calls
(call_expression
  function: (identifier) @_require
  arguments: (arguments (string) @import.require)
  (#eq? @_require "require"))

;; Function declarations
(function_declaration
  name: (identifier) @function.name) @function.def

;; Arrow functions assigned to variables
(lexical_declaration
  (variable_declarator
    name: (identifier) @function.name
    value: (arrow_function))) @function.def

;; Class declarations
(class_declaration
  name: (type_identifier) @class.name) @class.def

;; Method definitions
(method_definition
  name: (property_identifier) @method.name) @method.def

;; Export declarations
(export_statement
  declaration: (_) @export.declaration) @export

;; Type/Interface declarations
(interface_declaration
  name: (type_identifier) @type.name) @type.def

(type_alias_declaration
  name: (type_identifier) @type.name) @type.def

;; Call expressions (for call checking)
(call_expression
  function: (identifier) @call.name) @call

(call_expression
  function: (member_expression
    object: (identifier) @call.object
    property: (property_identifier) @call.method)) @call.member
