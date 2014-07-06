java2go
=======

##### Convert Java code to something like Go

This has been my free-time hacking project for the last few months.  It's in a pre-alpha state, but it can convert around 75% of the Java code on my laptop into something that kind of looks like Go code, and if you're lucky that converted code may even be compilable.

To use it:

 `go get github.com/dglo/java2go`

then either:

 `go run src/github.com/dglo/java2go/main.go -- path/to/my.java`

or:

 `go build github.com/dglo/java2go`

 `./java2go -- path/to/my.java`

If you specify a directory with `-dir`, it will write the translated file(s) to the corresponding `.go` file.

##### Tweaking the code to translate your project

This version does some minor transformations from idiomatic Java into
somewhat idiomatic Go, but doesn't do anything with channels or goroutines.

If you'd like to add your own Java-to-Go transformation(s), you can check out `java2go/parser/transform.go`.  The `-report` option is helpful in determining what should be transformed.

##### Bugs

The most difficult bug to fix is the Lex/Yacc code which chokes on some legal Java, especially the  '...' operator.
