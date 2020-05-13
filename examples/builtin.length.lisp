(print "'()" (length '()))
(print "'(1 2 3 4 5)" (length '(1 2 3 4 5)))


(def make-lazy ()
	 (set
	   iter
	   \(if (< _1 10)
	     (+ _1 1)
		 '()))
     (gen iter 0))
(print "(lazy 10)" (length (make-lazy)))
