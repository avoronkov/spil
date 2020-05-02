; Endless lazy lists

(def take (n lst) (take n lst '()))
(def take (0 lst acc) acc)
(def take (n lst acc) (take (- n 1) (tail lst) (append acc (head lst))))

(def lazy-filter (pred lst)
	 (set
	   filt 
	   (lambda
		 (if (pred (head _1))
		   (list (head _1) (tail _1))
		   (self (tail _1)))))
	 (gen filt lst))

(func modulo (n k)
	  (if (< n k)
		n
		(modulo (- n k) k)))

(func prime? (1) 'F)
(func prime? (n) (prime? n 2))
(func prime? (n n) 'T)
(func prime? (n i)
	  (if (= (modulo n i) 0)
		'F
		(prime? n (+ i 1))))

(set integers (gen (lambda (list (+ _1 1))) 0))

(print (take 25 (lazy-filter prime? integers)))
