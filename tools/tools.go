//go:build tools

package tools

import (
	_ "github.com/matryer/moq"
	_ "golang.org/x/tools/cmd/stringer"
)

//go:generate go build -v -o ../bin/stringer golang.org/x/tools/cmd/stringer
//go:generate go build -v -o ../bin/moq github.com/matryer/moq
