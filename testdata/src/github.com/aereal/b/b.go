package b

import (
	_ "errors" // OK
	_ "fmt"    // want `"fmt" cannot be imported by App`
)
