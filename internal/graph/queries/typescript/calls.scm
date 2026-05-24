(call_expression
  function: [
    (identifier) @call.local_func
    (member_expression
      object: (identifier) @call.receiver_or_pkg
      property: (property_identifier) @call.method
    )
  ]
) @call.site
