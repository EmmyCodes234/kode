(function_definition
  name: (identifier) @func.name
  parameters: (parameters) @func.params
  return_type: (type)? @func.return
) @func.def

(class_definition
  name: (identifier) @class.name
  body: (block) @class.body
) @class.def
