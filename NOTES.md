## Replace `VarRecord`
- Every variable should be assigned a `CellTemplate` during the **transform**
  stage. This includes imports and builtin functions.


## Imports and Builtins
- Import statements must be at the top-level of the script to allow for
  predictable dependency analysis.
- Imports and builtins are technically added as variables local to the top-level
  of scope. This will allow for easier renaming and manipulation of the
  non-local values.


## Operator Type Handling
- Any builtin operator will have an associated 2d lookup-table where one axis is
  the left operator and the other axis is the right operator. If the operators
  and types are supported, the intersection of both axes will contain a function
  for handling the operation and generating the appropriate bytecode
  instructions.
