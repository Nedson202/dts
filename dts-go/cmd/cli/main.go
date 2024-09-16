package main

import (
	"fmt"
	"os"

	"github.com/nedson202/dts-go/cmd/cli/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
