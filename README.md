java2go
=======

##### Convert Java code to something like Go

This has been my free-time hacking project for the last few months.  It's in
a pre-alpha state, but it can convert around 75% of the Java code on my
laptop into something that kind of looks like Go code, and if you're lucky
that converted code may even be compilable.

This version does some minor transformations from idionmatic Java into
somewhat idiomatic Go, but doesn't do anything with channels or goroutines.
