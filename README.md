# SPIL - SimPle LIsp implementation written in Go

![Go](https://github.com/avoronkov/spil/workflows/Go/badge.svg)


## Installation

```
$ go get gitlab.com/avoronkov/spil
$ spil
(print "hello world!")
^D
hello world!
```

## Language overview

Well, it's a kind of Lisp, so you write you code with the contructions like that:
```
(print (- (* 2 3) 1))
(print (+ 1 2 3 4 5))
(print "this is true:" 'T "and this is false:" 'F)
(print "this is a raw list:" '(1 2 3 print hello))
```

Also it's (almost) pure functional language, so you have no mutable variables and no loops.
(Actually `print` is the only statement with side effects).

### Comments

Lines started with `;` or `#` are comments.

### Data types

SPIL (like most of other Lisps) has atoms and lists as basic data type.
Atoms include the following:

- Integers (0 1 25 1235 -128 ...)

- Booleans ('T 'F)

- Strings ("hello world!" "foo" "bar" ...)

- Identifiers (foo bar func if ...)

Lists include:

- unquoted s-expressions (`(foo 1 2 3)` ... )

- and quoted (raw) ones (`'(this is "not" evaluated)`)

### Everything is an expression

Every statement in SPIL is an expression i.e. every statement is r-value which can be returned from function or assigned to "variable".

### Basic functions

SPIL has the following builtin functions implemented in Go:

- `print` - prints values of expressions on stdout.

- Arithmetic operations: `+`, `-`, `*`, `/`, `mod`, `<`, `>`, `<=`, `>=`.

- Equality operator: `=`

- Functions to work with lists: `head`, `tail`, `append`, `list`, `empty`.

### User-defined functions

You may define you own function with keyword `def` (of `func`):
```
; (def <function-name> <function-parameters> <body-statement1> ...)
(def plus-one (n) (+ n 1))

(print (plus-one 3))
; 4
```

Return value of function is a return value of last expression in function.

Note that `def` defined so-called "pure" function i.e. its return value can depend only on its arguments.

Function may have multiple definitions with different set of arguments:
```
(def factorial (0) 1)
(def factorial (n) (* n (factorial (- n 1))))
```

### Control flows

SPIL has conditional operator `if` which has the following syntax:
```
(if  some-condition  return-value-if-true return-value-if-false)
```

Note that `if` is also an expression i.e. it has a return value.

### Recursion

SPIL has no loops. Instead it uses recursion as in example above:
```
(def factorial (0) 1)
(def factorial (n) (* n (factorial (- n 1))))
```

Note that such recursion is not very effective because it consumes call-stack.
That's why it's better to use tail-call recursion like that:
```
(def factorial (n) 1)
(def factorial (0 result) result)
(def factorial (n result) (factorial (- n 1) (* result n)))
```
SPIL has Tail Call Optimization so the result will be returned from `(factorial 0 result)` directly to the caller.

If you are not familiar with recursion and tail calls you may read a great book for functional programming beginners [Learn you some Erlang for great good](https://learnyousomeerlang.com/).

### Passing functions as arguments to other functions

You may pass function as an argument by its name:
```
(func plus-one (n) (+ n 1))

(func apply-func-to-ten (fn) (fn 10))

(print (apply-func-to-ten plus-one))
; 11
```

### Lambdas

You can define lambda-functions with `lamda` keyword.
```
(func apply-func-to-ten (fn) (fn 10))

(print (apply-func-to-ten (lambda (+ _1 1))))
; 11
```
Lambdas are very similar to regular functions but they have some difference:

- Lambda can grab values of variable from the context where lambda is defined:
```
(func apply-func-to-ten (fn) (fn 10))

(set n 5)

(print (apply-func-to-ten (lambda (+ _1 n))))
; 15
```

- Lambdas are designed to be small so they use short syntax of accessing arguments:
`_1 _2 _3 ...` for accessing positional arguments and `__args` for accessing whole list of arguments.

### Lazy lists

You can use keyword `gen` to define finite or infinite lazy lists.
For example lazy-list of positive integers can be defined like this:
```
(def inc (n) (+ n 1))

; infinite lazy list of integers: (1 2 3 4 ...)
(set ints (gen inc 0))
```

`gen` has the following syntax:
```
(gen <iterator-function> <initial-state>)
```
When somebody asks for `head` of lazy lists then `iterator-function` is called with value of previous state.
Iterator should return one of the following:

- Empty list `'()` to indicate that list has ended.

- List with one element `(list value)` which will be returned as next element (head) in lazy-list and will be passed to the next call of iterator.

- List of two elements `(list value new-state)`. `value` will be returned by `head`, `new-state` will be passed to the next call of iterator.

For example the infinite list of Fibonacci numbers:
```
(def next-fib (prev)
	(set a (head prev))
	(set b (head (tail prev)))
	(list b (list b (+ a b))))
(set fibs (gen next-fib '(1 1)))

(print (take 10 fibs))
; '(1 2 3 5 8 13 21 34 55 89)
```

### Using modules

You can `use` other modules in your program:
```
(use "some-module.lisp")

(function-from-some-module ...)
```

### Big math
You can use big integers instead of int64 in calculations by adding `(use bigmath)` statement and the beginning of the main module.

### Memoization

You can tell the interpreter to remember function results by defining function with `def'` (or `func'`) keyword.
As a result if such function is called with the same set of arguments twice then its result will be calculated only once.
Second time it will return the stored result.

```
(def' x2 (n) (print "evaluating x2" n) (* n 2))

(print (x2 5))
(print (x2 6))
(print (x2 5))
; evaluating x2 5
; 10
; evaluating x2 6
; 12
; 10
```

### Work with files

You can work with files as lazy-strings (?).
Well, it means that you can open file and iterate over its content with `head` and `tail` methods.
It may seems kinda low-lever so I've implemented functions `lines` and `words` to split string into lines of words
and these functions are also lazy.

```
(set' file (open "somefile.txt"))

(print (map words (lines files))
```

Note that operator `set'` is used instead of simple `set`.
It means that file will be automatically closed when interpreter leaves the current function scope.

(Writing into files is not implemented yet.)

## Types

You can specify types of your function parameters and function's return value.
```
(def contains (value:any '()) :bool 'F)
(def contains (value:any lst:list) :bool
	(if (= (head lst) value)
		'T
		(contains value (tail lst))))

(print (contains 4 '(1 3 5 8)))
```

The following builtin type are available: `:int`, `:str`, `:bool`, `:list`, `:any`.

## Static type checking

SPIL checks the correctness of types usage in "compile time", i.e. before actual execution of the the program.
You can specify option "--check" (or "-c") for syntax and type checking of the program.
E.g. when you misplace the arguments in previous example (`(print (contains '(1 3 5 8) 4))`) you will get the following error:
```
$ spil -c example.lisp
__main__: contains: no matching function implementation found for [{:list {S': {Int64: 1} {Int64: 3} {Int64: 5} {Int64: 8}}} {:int {Int64: 5}}]
```

## Type casting

Sometimes you need to cast expressions types. E.g. in the following example:
```
(def ascending? (l:list) :bool
     (if (<= (length l) 1)
       'T
       (if (> (first l) (second l))
         'F
         (ascending? (tail l)))))

(print (ascending? '(1 2 3 5 8)))
```
you will get the error:
```
ascending?: >: Expected all integer arguments, found {:any <nil>} at position 0
```
because `nth` returns `:any` but `>` expects `:int`.
So you can fix it with casting first and second elemets to `:int`:
```
(def ascending? (l:list) :bool
     (if (<= (length l) 1)
       'T
       (if (> (do (first l) :int) (do (second l) :int))
         'F
         (ascending? (tail l)))))
```
I may look strange but actually it's rather simple. SPIL has the following forms of types casting:
```
; convert result of function to :int
(def get-int () :int (function-returning-any) :int)

; variable var has type :int now
(set var (function-returning-any) :int)

; convert return of do-block to :int
(do (function-returning-any) :int)
```

## User defined types

You may define your own type with `deftype` statement:
```
; (deftype new-type parent-type)
(deftype :my-type :any)
```
It may be helpful in some scenarios, i.e. if we want to implement simple "type-safe" set:
```
(deftype :set :list)

(def set-new () :set '() :set)
(def set-add (elem:any s:set) :set
	(if (contains elem s)
	  s
	  (do (append s elem) :set)))

;; This will cause typecheck error:
(set-add '(1 2 2 4 5) 6)

;; This is OK
(set s1 (set-add (set-new) 1))
(set s2 (set-add s1 2))
(set s3 (set-add s2 2))
(set s4 (set-add s3 3))

(print s4 (length s4))
```
Note that you cannot use :list variable where :set is required, but you can pass :set anywhere where its parent type (:list) is accepted.

## Examples

You can find some examples of code [here](https://gitlab.com/avoronkov/spil/-/tree/master/examples)

## TODO

- [+] do-statement support

- [+] multiple function definition with pattern matching

- pass command line arguments to the command

- [+] lazy lists

- [+] apply

- [+] anonymous functions (?)

- [+] function "list"

- restricted type casting and strict mode.

- "length" and "nth" optimization for static listst.

- Separate pragma parsing and loading std-lib first.

- Functions overloading for user defined types

- "error" and "catch" functions for runtime errors

- Forbidden matching (:delete or something)

- Type of variable is vanished when placed into list.
