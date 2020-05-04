; Tail-call optimization test

(def long-sum (n acc)
	 (if (= n 0)
	   acc
	   (long-sum (- n 1) (+ acc 1))))

(print (long-sum 1000000 0))

; vim: ft=lisp
