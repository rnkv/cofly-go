# cofly-go

Tiny change/patch library for Go values (`any`), designed for applying **changes on the fly**.
The name **Cofly** comes from the phrase **“Changes on the fly”**.

## Installation

```bash
go get github.com/rnkv/cofly-go
```

## Supported value model

The library works with values shaped like JSON:

- `nil`
- `bool`
- `int`, `float64` (with a special equivalence rule, see below)
- `string`
- `map[string]any`
- `[]any`

Anything else is considered **unsupported** (some functions return `false`, others panic; see each function contract below).

## JSON compatibility

Cofly works with a JSON-shaped data model and produces changes that are **JSON-serializable** (maps, arrays, primitives).
This makes it convenient for patching values that come from JSON (or will be sent as JSON).

Notes:

- **Numbers**: JSON has just “number”, but Go often uses `int`/`float64`. Cofly treats `int` and `float64` as equal when numerically equal (`1` equals `1.0`).
- **`Undefined`**: standard JSON has no `undefined`. Cofly introduces its own JSON-compatible extension for deletions: the marker is the string `"\u0000"` (NUL). It is JSON-serializable, but it is **reserved** by the contract (see below).

## Reserved value: `Undefined`

`Undefined` is a special marker:

- In changes for objects (`map[string]any`), it means **delete this key**.

Important: `Undefined` is the string `"\x00"`, so the string value `"\x00"` is **reserved** by the contract.

## Public API

### `Difference(oldValue, newValue any) any`

Computes a change that transforms `oldValue` into `newValue`.

Return value:

- `Undefined` means “no change”
- `nil` means “set to nil”
- otherwise returns either:
  - a primitive (replacement), or
  - a `map[string]any` (object change-map or array splice-map), or
  - a `[]any` (replacement)

#### Primitive example

```go
oldValue := 1
newValue := 2

change := cofly.Difference(oldValue, newValue) // 2
```

JSON view:

```json
{ "old": 1, "new": 2, "change": 2 }
```

#### Object (map) example

```go
oldValue := map[string]any{"a": 1, "b": 2}
newValue := map[string]any{"a": 1, "b": 3, "c": "new"}

change := cofly.Difference(oldValue, newValue).(map[string]any)
// change == map[string]any{"b": 3, "c": "new"}
```

JSON view:

```json
{
  "old": { "a": 1, "b": 2 },
  "new": { "a": 1, "b": 3, "c": "new" },
  "change": { "b": 3, "c": "new" }
}
```

Deletion uses `Undefined`:

```go
oldValue := map[string]any{"a": 1, "b": 2}
newValue := map[string]any{"a": 1}

change := cofly.Difference(oldValue, newValue).(map[string]any)
// change["b"] == cofly.Undefined
```

JSON view (deletion marker is `"\u0000"`):

```json
{
  "old": { "a": 1, "b": 2 },
  "new": { "a": 1 },
  "change": { "b": "\u0000" }
}
```

#### Array example (splice-map format)

For arrays, `Difference` returns a **splice map**: `map[string]any` where keys are spans:

- `"i..j"`: a span inside the array (from `i` to `j`, like half-open `[i, j)`)
- `"i.."`: a span starting at `i` (for insertions before `i`, or append when `i == len(old)`), where `j == i`

The value is always a `[]any` payload.

Splice-map semantics:

- **No key** for some index range means **no change** for that range (this is a change-map, not a full array snapshot).
- **Deletion** is represented by a splice with an **empty payload** (`[]` in JSON / `[]any{}` in Go).
- **Out-of-bounds insertions are “magnetized” to the array**: if a splice inserts far to the left (negative indices) it behaves like a prepend, and if it inserts far to the right (beyond `len(old)`) it behaves like an append (no gaps are created).
- **Splice spans must not overlap**. Overlapping spans make the patch ambiguous; Cofly treats such splice-maps as invalid and will panic when applying them.

Implementation note: array changes are computed using the **Myers diff algorithm** (Myers, 1986).

Examples:

```go
oldArray := []any{"a", "b", "c"}
newArray := []any{"a", "B", "c", "d"}

change := cofly.Difference(oldArray, newArray).(map[string]any)
// change == map[string]any{
//   "1..2": []any{"B"}, // replace element 1
//   "3..":  []any{"d"}, // append at end
// }
```

JSON view:

```json
{
  "old": ["a", "b", "c"],
  "new": ["a", "B", "c", "d"],
  "change": {
    "1..2": ["B"],
    "3..": ["d"]
  }
}
```

Deletion from an array:

```go
oldArray := []any{"a", "b"}
newArray := []any{"a"}

change := cofly.Difference(oldArray, newArray).(map[string]any)
// change == map[string]any{"1..2": []any{}}
```

JSON view:

```json
{
  "old": ["a", "b"],
  "new": ["a"],
  "change": { "1..2": [] }
}
```

#### Numeric equivalence (`int` vs `float64`)

`Difference(1, 1.0)` returns `Undefined` (they are treated as equal), and the same for `Difference(1.0, 1)`.

### `Merge(target, change any, doClean bool) any`

Applies `change` to `target` and returns the resulting value.

General rules:

- If `change == nil`, result is `nil` (set-to-nil).
- If `change` is a primitive, result is that primitive (replacement).
- If `change` is a `map[string]any`:
  - if it is a valid splice-map, it is applied to arrays (`[]any`)
  - otherwise it is treated as an object change (or a full replacement, depending on target type)

You can also **merge a change into another change** (compose patches):

- Object changes (`map[string]any` change-maps) can be merged into existing object changes.
- Array splice-maps can be merged into arrays; merging splice-map into splice-map is not implemented yet.

`doClean` controls deletions in object merge:

- `doClean == true`: keys with `Undefined` are removed from the target map
- `doClean == false`: keys with `Undefined` are kept as-is (the marker is stored)

#### Object merge example

```go
target := map[string]any{"a": 1, "b": 2}
change := map[string]any{"b": 3, "c": "new"}

outputValue := cofly.Merge(target, change, true).(map[string]any)
// outputValue == map[string]any{"a": 1, "b": 3, "c": "new"}
```

JSON view:

```json
{
  "target": { "a": 1, "b": 2 },
  "change": { "b": 3, "c": "new" },
  "out": { "a": 1, "b": 3, "c": "new" }
}
```

Deletion with `doClean`:

```go
target := map[string]any{"a": 1, "b": 2}
change := map[string]any{"b": cofly.Undefined}

outClean := cofly.Merge(cofly.Clone(target), change, true).(map[string]any)
// outClean == map[string]any{"a": 1}

outKeep := cofly.Merge(cofly.Clone(target), change, false).(map[string]any)
// outKeep["b"] == cofly.Undefined
```

#### Array merge (apply splice-map)

```go
target := []any{"a", "b", "c"}
change := map[string]any{
  "1..2": []any{"B"}, // replace
  "3..":  []any{"d"}, // append
}

outputArray := cofly.Merge(target, change, true).([]any)
// outputArray == []any{"a", "B", "c", "d"}
```

JSON view:

```json
{
  "target": ["a", "b", "c"],
  "change": { "1..2": ["B"], "3..": ["d"] },
  "out": ["a", "B", "c", "d"]
}
```
### `Apply(target *any, isSnapshot bool, change *any, doClean bool) bool`

Convenience helper for two modes:

- **Snapshot mode** (`isSnapshot == true`):
  - treats `*change` as the new full value (snapshot)
  - computes a change: `*change = Difference(*target, snapshot)`
  - if the change is `Undefined`, returns `false` and does not change `*target`
  - otherwise sets `*target = snapshot` and returns `true`

- **Patch mode** (`isSnapshot == false`):
  - applies: `*target = Merge(*target, *change, doClean)`
  - returns `true`

Example:

```go
var target any = map[string]any{"a": 1}
var snapshot any = map[string]any{"a": 2}

applied := cofly.Apply(&target, true, &snapshot, true)
// applied == true
// target == map[string]any{"a": 2}
// snapshot now holds the change (map[string]any{"a": 2})
```

### `Equal(a, b any) bool`

Deep equality for supported values.

Notes:

- `int` and `float64` are considered equal when numerically equal (`1` equals `1.0`).
- Unsupported values return `false`.

```go
cofly.Equal(1, 1.0) // true
cofly.Equal([]any{1, "a"}, []any{1.0, "a"}) // true
```

### `Clone(value any) any`

Deep clone for supported values (`map[string]any` / `[]any`).
Panics for unsupported types.

```go
originalValue := map[string]any{"nested": map[string]any{"a": 1}}
clonedValue := cofly.Clone(originalValue).(map[string]any)
_ = clonedValue
```

## Typical workflow

```go
oldValue := map[string]any{"a": 1, "arr": []any{"x", "y"}}
newValue := map[string]any{"a": 2, "arr": []any{"x", "Y", "z"}}

change := cofly.Difference(oldValue, newValue)
patchedValue := cofly.Merge(cofly.Clone(oldValue), change, true)

// patchedValue should now represent newValue (per library semantics).
_ = patchedValue
```
