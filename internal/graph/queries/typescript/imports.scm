(import_statement
  (import_clause
    (identifier)? @import.default
    (named_imports (import_specifier name: (identifier) @import.name))?
  )
  source: (string) @import.path
) @import.statement
