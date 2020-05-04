# Function with variable number of arguments

( func count-args-aux (lst ) 
		(if (= lst '() )
		  0
		  (+ 1 (count-args-aux (tail lst) ) ) ) )

( func print-count-args args (print (count-args-aux args ) ) )

(print-count-args 1 2 3 )
(print-count-args 1 )
(print-count-args )

# vim: ft=lisp
