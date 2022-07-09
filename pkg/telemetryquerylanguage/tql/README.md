# Telemetry Query Language

The Telemetry Query Language is a query language for transforming open telemetry data based on the [OpenTelemetry Collector Processing Exploration](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/processing.md).

This package reads in TQL queries and converts them to invokable conditions and functions based on the TQL's grammar.

The TQL is signal agnostic; it is not aware of the type of telemetry on which it will operate.  Instead, the conditions and functions returned by the package must be passed a TransformContext, which provide access to the signal's telemetry. 

## Grammar

The TQL grammar includes Invocations, Values and Conditions.

### Invocations

Invocations represent a function call. Invocations are made up of 2 parts

- a string identifier. The string identifier must start with a letter or an underscore (`_`).
- zero or more Values (comma separated) surrounded by parentheses (`()`).

**The TQL does not define any functions implementations.** Users must supply a map between string identifiers and the actual function implementation.  The TQL will use this map and reflection to generate Invocations, that can then be invoked by the user.

Example Invocations
- `drop()`
- `set(field, 1)`

### Values

Values are the things that get passed to an Invocation or used in a Condition. Values can be either a Path, a Literal, or an Invocation.  

Invocations as Values allows calling functions as parameters to other functions. See [Invocations](#invocations) for details on Invocation syntax.

#### Paths

A Path Value is a reference to a telemetry field.  Paths are made up of string identifiers, dots (`.`), and square brackets combined with a string key (`["key"]`).  **The interpretation of a Path is NOT implemented by the TQL.**  Instead, the user must provide a `PathExpressionParser` that the TQL can use to interpret paths.  As a result, how the Path parts are used is up to the user.  However, it is recommended, that the parts be used like so:

- Identifiers are used to map to a telemetry field.  
- Dots (`.`) are used to separate nested fields.
- Square brackets and keys (`["key"]`) are used to access maps or slices.

Example Paths
- `name`
- `resource.name`
- `resource.attributes["key"]`

#### Literals

Literals are literal interpretations of the Value into a Go value.  Accepted literals are:

- Strings. Strings are represented as literals by surrounding the string in double quotes (`""`).
- Ints.  Ints are represented by any digit, optionally prepended by plus (`+`) or minus (`-`). Internally the TQL represents all ints as `int64`
- Floats.  Floats are represented by digits separated by a dot (`.`), optionally prepended by plus (`+`) or minus (`-`). The leading digit is optional. Internally the TQL represents all Floats as `float64.
- Bools.  Bools are represented by the exact strings `true` and `false`.
- Nil.  Nil is represented by the exact string `nil`. 
- Byte slices.  Byte slices are represented via a hex string prefaced with `0x`

Example Literals
- `"a string"`
- `1`, `-1`
- `1.5`, `-.5`
- `true`, `false`
- `nil`,
- `0x0001`

### Conditions

Conditions allow a decision to be made about whether an Invocation should be called. The TQL does not force a condition to be used, it only allows the opportunity for the condition to be invoked before invoking the associated Invocation.  Conditions allways return true or false.

Conditions are made up of a left Value, an operator, and a right Value. See [Values](#values) for details on what a Value can be.

Operators determine how the two Values are compared.  The valid operators are:

- Equal (`==`). Equal (`==`) checks if the left and right Values are equal, using Go's `==` operator.
- Not Equal (`!=`).  Not Equal (`!=`) checks if the left and right Values are not equal, using Go's `!=` operator.

## Examples

These examples contain a SQL-like declarative language.  Applied statements interact with only one signal, but statements can be declared across multiple signals.  Functions used in examples are indicative of what could be useful, but are not implemented by the TQL itself.

### Remove a forbidden attribute

```
traces:
  delete(attributes["http.request.header.authorization"])
metrics:
  delete(attributes["http.request.header.authorization"])
logs:
  delete(attributes["http.request.header.authorization"])
```

### Remove all attributes except for some

```
traces:
  keep_keys(attributes, "http.method", "http.status_code")
metrics:
  keep_keys(attributes, "http.method", "http.status_code")
logs:
  keep_keys(attributes, "http.method", "http.status_code")
```

### Reduce cardinality of an attribute

```
traces:
  replace_match(attributes["http.target"], "/user/*/list/*", "/user/{userId}/list/{listId}")
```

### Reduce cardinality of a span name

```
traces:
  replace_match(name, "GET /user/*/list/*", "GET /user/{userId}/list/{listId}")
``` 

### Reduce cardinality of any matching attribute

```
traces:
  replace_all_matches(attributes, "/user/*/list/*", "/user/{userId}/list/{listId}")
``` 

### Decrease the size of the telemetry payload

```
traces:
  delete(resource.attributes["process.command_line"])
metrics:
  delete(resource.attributes["process.command_line"])
logs:
  delete(resource.attributes["process.command_line"])
```

### Drop specific telemetry

```
metrics:
  drop() where attributes["http.target"] = "/health"
```

### Attach information from resource into telemetry

```
metrics:
  set(attributes["k8s_pod"], resource.attributes["k8s.pod.name"])
```

### Group spans by trace ID

```
traces:
  group_by(trace_id, 2m)
```


### Update a spans ID

```
logs:
  set(span_id, SpanID(0x0000000000000000))
traces:
  set(span_id, SpanID(0x0000000000000000))
```

### Create utilization metric from base metrics.

```
metrics:
  create_gauge("pod.cpu.utilized", read_gauge("pod.cpu.usage") / read_gauge("node.cpu.limit")
```