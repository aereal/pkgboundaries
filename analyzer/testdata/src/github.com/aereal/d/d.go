package d

import (
	_ "encoding/json" // want `"encoding/json" cannot be imported by App`
	_ "errors"        // want `"errors" cannot be imported by App`
	_ "fmt"           // want `"fmt" cannot be imported by App`
)
