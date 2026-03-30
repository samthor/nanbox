Go package to pack data into the NaN bits of `float64`.

This is a trick used by various JS engines to avoid memory waste.

None of the methods here will ever generate the "sentinel" NaN that Go internally gives you when you call `math.NaN()`.
NaN is not comparable, ... except via using `math.Float64bits()` to coerce to a `uint64`: this value will never be equal.

Why?
I don't know.
Maybe you have a reason.

### Usage

```go
import (
  "math"

  "github.com/samthor/nanbox"
)

func foo() {
  x := nanbox.PackInt32(int32(1234))
  if !math.IsNaN(x) {
    panic("must be NaN")
  }

  y := nanbox.UnpackInt32(x)
  if y != 1234 {
    panic("must be 1234")
  }
}
```
