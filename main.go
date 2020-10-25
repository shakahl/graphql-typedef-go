//go:generate statik -f -dest=internal -p statikdata -src=assets/templates/ -include=*.gotmpl,*.graphql

package main

import (
	"github.com/shakahl/gqlassist/cmd"
)

func main() {
	cmd.Execute()
}
