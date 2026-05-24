(function_declaration
  name: (identifier) @func.name
  parameters: (formal_parameters) @func.params
  return_type: (type_annotation)? @func.return
) @func.def

(method_definition
  name: (property_identifier) @method.name
  parameters: (formal_parameters) @method.params
  return_type: (type_annotation)? @method.return
) @method.def

(class_declaration
  name: (type_identifier) @class.name
  body: (class_body) @class.body
) @class.def
