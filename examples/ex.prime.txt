; Check if number is prime

(func modulo (n k)
	  (if (< n k)
		n
		(modulo (- n k) k)))

(func prime-aux (n i)
	  (if (= n i)
		'T
		(if (= (modulo n i) 0)
		  'F
		  (prime-aux n (+ i 1)))))

(func prime? (n) (prime-aux n 2))

(func primes-between (start finish)
	  (if (prime? start)
		(print start)
		'())
	  (if (= start finish)
		'()
		(primes-between (+ start 1) finish)))

(primes-between 2 30)
; vim: ft=lisp
