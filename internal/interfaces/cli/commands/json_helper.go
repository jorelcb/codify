package commands

import (
	"encoding/json"
	"fmt"
	"os"
)

// encodeJSON imprime v como JSON con indentación de 2 espacios. Usado por
// los comandos que ofrecen --json (check, reset-state, etc).
func encodeJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json encode failed: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
