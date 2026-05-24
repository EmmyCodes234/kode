(call
  function: [
    (identifier) @call.local_func
    (attribute
      object: (identifier) @call.receiver_or_pkg
      attribute: (identifier) @call.method
    )
  ]
) @call.site
