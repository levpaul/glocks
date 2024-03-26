### GLOCKS - A Go implementation of [Lox](https://craftinginterpreters.com/the-lox-language.html)

This is an attempt to implement the Lox language, defined by Robert Nystrom, in his book [Crafting Interpreters](https://craftinginterpreters.com). It's current a work in progress and no guarantees of correctness are claimed.

#### Running Glocks

To compile the code, run `go build cmd/glocks.go` - and it will produce a binary named `glocks` in the root of the repo. You can use it like so:

`$ glocks`

This will start a [REPL](https://en.wikipedia.org/wiki/Read%E1%80%93eval%E2%80%93print_loop) where you can run arbitrary lines of Lox at will.


`$ glocks run FILE_NAME`

You can pass in a Lox script, <FILE_NAME>, and glocks will interpret and execute it.


#### Developing Glocks

The entire source of Glocks is in this repo and should be somewhat straight forward to follow, from the book.

Notable differences between Glocks and the Java Lox implementation include:
 - There's no boilerplate code generator for AST classes, because it's Go and there's a whole lot less cruft needed for struct definitions. You can find the AST Nodes defined in `parser/nodes.go`.
 - No use of generics in visitor implementation. With duck typing in Go, there wasn't any need for generics, even with Go native support for them


#### Testing

There are unit and acceptance tests throughout the codebase. Run `go test ./...` to run the selection. Running individual tests in debug-mode through Delve or your IDE can be a very useful way to dig into issues or understand the interpreter in practice.


#### Development Activity

I circle back to this project only sporadically, feel free to open an issue if you have any questions. Hopefully I'll finish the implementation some time in 2024.