## Focus: Architecture

Review the diff for structural decisions that will be **costly to fix later**.

**Flag:**
- Business logic placed in transport/handler layers, or I/O mixed into pure computation
- Concrete type dependencies where an interface at the package boundary would decouple components
- Circular or upward dependencies between packages — a lower-level package importing a higher-level one
- God structs or functions that accumulate unrelated responsibilities

**Do not flag:**
- Code that is simple enough that an abstraction would add complexity without benefit
- Small programs or scripts where layering would be over-engineering
