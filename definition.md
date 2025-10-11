# Language definition
type Animal = Dog | Cat | Bird

### Data types
- int, i8, i16, i32, i64
- uint, u8, u16, u32, u64
- float, f32, f64
- bool
- string
- array<Type>
- map<KeyableType, ValueType>

Identifiers can be anything starting with a letter or underscore followed by letters, digits, or underscores.

### Code example
```
  include "path to file or module folder"

  ReturnType name(arg1: Type, arg2: Type) {
    Type verboseVar = value
    variable := value

    // conditional statements always return a value
    result = if condition {
      valueIfTrue
    } else {
      valueIfFalse
    }
    
    // theres no while, for has superporwers
    for item in collection {
      // do something with item
    }

    for i := 0; i < 10; i = i + 1 {
      // do something with i
    }

    // no switch, use match instead, match is super powerful, match when always matches first case
    match expression {
      when value1 {
        // do something for value1
      }
      when value2 {
        // do something for value2
      }
      default {
        // default case
      }
    }

    a = (2, 3, "hello")
    match a {
      when (x, 3, z) {
        // x = 2, z = "hello"
        // do something with x, y, z
      }
    }

  }
```
