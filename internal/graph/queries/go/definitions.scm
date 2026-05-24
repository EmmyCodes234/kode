(function_declaration
  name: (identifier) @func.name
  parameters: (parameter_list) @func.params
  result: [(parameter_list) (type_identifier)]? @func.return
) @func.def

(method_declaration
  receiver: (parameter_list) @method.receiver
  name: (field_identifier) @method.name
  parameters: (parameter_list) @method.params
  result: [(parameter_list) (type_identifier)]? @method.return
) @method.def
