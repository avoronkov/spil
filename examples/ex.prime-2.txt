; Check if number is prime

(func modulo (n k)
	  (if (< n k)
		n
		(modulo (- n k) k)))

(func prime? (n) (prime? n 2))
(func prime? (n n) 'T)
(func prime? (n i)
	  (if (= (modulo n i) 0)
		'F
		(prime? n (+ i 1))))

(func primes-between (start finish)
	  (primes-between start finish '()))
(func primes-between (f f acc) acc)
(func primes-between (s f acc)
	  (primes-between
		(+ s 1)
		f
		(if (prime? s) (append acc s) acc)))

(print (primes-between 2 30))

; vim: ft=lisp
