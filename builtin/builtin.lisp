;; standard lisp functions

;; lazy filter
(def filter (pred lst)
	 (set 
	   filt
	   (lambda
		 (if (empty _1)
		   '()
		   (if (pred (head _1))
			 (list (head _1) (tail _1))
			 (self (tail _1))))))
	 (gen filt lst))

;; lazy map
(def map (fn lst)
	 (set
	   iter
	   (lambda
		 (if (empty _1)
		   '()
		   (list (fn (head _1)) (tail _1)))))
	 (gen iter lst))

;; take first n values from list
(def take (n lst)
	 (set
	   iter
	   (lambda
		 (do
		   (set cn (head _1))
		   (set cl (head (tail _1)))
		   (if (or (= cn 0) (empty lst))
			 '()
			 (list (head cl) (list (- cn 1) (tail cl)))))))
	 (gen iter (list n lst)))
