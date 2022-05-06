This package has a modified copy of the go cover package source code... or rather parts of it.

* [The cover story](https://go.dev/blog/cover) explains how the coverage package works in Go
* [Cover package](https://pkg.go.dev/golang.org/x/tools/cover) details including list of exported variables and functions
* [Cover tool](https://cs.opensource.google/go/go/+/refs/tags/go1.18.1:src/cmd/cover/cover.go) the source-code for the cover tool.

(`edit_copy` can't be accessed from outside the `go cmd` tool because it's an internal module. So I had to copy it too. Commands can use that though, and cover is a tool)

## Example usage
```
$ go run .                                            <-- run
Enter name of file to annotate                        
hello.go                                              <-- file to annotate
//line hello.go:1                                     <-- annotated file contents
package main

import "fmt"

func Hello() {GoCover.Count[0]++;
        fmt.Println("Hello world!")
}

var GoCover = struct {                                <-- struct to keep track of coverage data
        Count     [1]uint32
        Pos       [3 * 1]uint32
        NumStmt   [1]uint16
} {
        Pos: [3 * 1]uint32{
                5, 7, 0x2000e, // [0]
        },
        NumStmt: [1]uint16{
                1, // 0
        },
})
```