;; Rust tree-sitter query — extract use statements, function/impl defs
;; Used by the multi-language verification gates

;; Use declarations
(use_declaration
  argument: (_) @import.path) @import

;; Function definitions
(function_item
  name: (identifier) @function.name) @function.def

;; Struct definitions
(struct_item
  name: (type_identifier) @struct.name) @struct.def

;; Enum definitions
(enum_item
  name: (type_identifier) @enum.name) @enum.def

;; Impl blocks
(impl_item
  type: (type_identifier) @impl.type) @impl.def

;; Trait definitions
(trait_item
  name: (type_identifier) @trait.name) @trait.def

;; Macro invocations (for call checking)
(macro_invocation
  macro: (identifier) @call.macro) @call

;; Function calls
(call_expression
  function: (identifier) @call.name) @call.fn

(call_expression
  function: (scoped_identifier
    path: (identifier) @call.path
    name: (identifier) @call.name)) @call.scoped

;; Method calls
(call_expression
  function: (field_expression
    field: (field_identifier) @call.method)) @call.method_call

;; Mod declarations
(mod_item
  name: (identifier) @mod.name) @mod.def
