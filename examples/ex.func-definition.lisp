; Polymorphic function definition

(def fact (0 acc) acc)
(def fact (n acc) (fact (- n 1) (* n acc)))
(def fact (n) (fact n 1))

(print (fact 5))
(print (fact 6))

; vim: ft=lisp
