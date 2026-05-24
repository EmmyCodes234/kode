(call_expression
  function: [
    (identifier) @call.local_func
    (selector_expression
      operand: (identifier) @call.receiver_or_pkg
      field: (field_identifier) @call.method
    )
  ]
) @call.site
