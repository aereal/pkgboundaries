package a

import (
	_ "encoding/base64" // want `"encoding/base64" cannot be imported by App`
	_ "encoding/json"   // want `"encoding/json" cannot be imported by App`
	_ "errors"          // OK
	_ "fmt"             // want `"fmt" cannot be imported by App`
)
