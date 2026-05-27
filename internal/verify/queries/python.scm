;; Python tree-sitter query — extract imports, function defs, class defs
;; Used by the multi-language verification gates (syntax, imports, calls)

;; Import statements
(import_statement
  name: (dotted_name) @import.name) @import

;; From imports
(import_from_statement
  module_name: (dotted_name) @import.source) @import.from

(import_from_statement
  module_name: (relative_import) @import.relative) @import.from.relative

;; Imported names
(import_from_statement
  name: (dotted_name) @import.name)

;; Function definitions
(function_definition
  name: (identifier) @function.name) @function.def

;; Async function definitions  
(function_definition
  name: (identifier) @function.name) @function.async

;; Class definitions
(class_definition
  name: (identifier) @class.name) @class.def

;; Method definitions (inside class)
(class_definition
  body: (block
    (function_definition
      name: (identifier) @method.name))) @method.def

;; Decorated definitions
(decorated_definition
  definition: (function_definition
    name: (identifier) @decorated.function.name)) @decorated.def

(decorated_definition
  definition: (class_definition
    name: (identifier) @decorated.class.name)) @decorated.class

;; Call expressions (for call checking)
(call
  function: (identifier) @call.name) @call

(call
  function: (attribute
    object: (identifier) @call.object
    attribute: (identifier) @call.method)) @call.member

;; Global assignments (module-level variables)
(expression_statement
  (assignment
    left: (identifier) @assignment.name)) @assignment
