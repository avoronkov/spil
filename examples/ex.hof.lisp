; High order function example

(def hof (fn) (fn 1))

(def plus1 (n) (+ n 1))

(print (hof plus1))

; vim: ft=lisp
