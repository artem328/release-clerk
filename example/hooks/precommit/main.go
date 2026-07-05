package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/artem328/release-clerk/hook"
)

func main() {
	var isError bool

	flag.BoolVar(&isError, "error", false, "return an error")
	flag.Parse()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var h hook.Input
	if err := json.Unmarshal(data, &h); err != nil {
		panic(err)
	}

	logs := []hook.Log{{Message: "Log1"}, {Message: "Log2"}, {Message: "Log3", Debug: true}}

	var out hook.Output
	if isError {
		out = hook.NewErrorPrecommitOutput(hook.Error{Message: "Hook failed"}, logs...)
	} else {
		out = hook.NewPrecommitOutput(logs...)
	}

	j, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprint(os.Stdout, string(j))
}
