; Factorial function example

( func fact ( n ) ( if ( = n 0 )  1  ( * n ( fact (- n 1 ) ) )  ) )

(print ( fact 6 ) ) 

; vim: ft=lisp
