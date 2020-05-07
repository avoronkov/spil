;; standard lisp functions

;; lazy filter
(def filter (pred lst) :list
	 (set
	   filt
	   (lambda
		 (if (empty _1)
		   '()
		   (if (pred (head _1))
			 (list (head _1) (tail _1))
			 (self (tail _1))))))
	 (gen filt lst))

(def filter' (pred lst) :list
	 (set
	   filt
	   (lambda
		 (if (empty _1)
		   '()
		   (if (pred (head _1))
			 (list (head _1) (tail _1))
			 (self (tail _1))))))
	 (gen' filt lst))


;; lazy map
(def map (fn lst) :list
	 (set
	   iter
	   (lambda
		 (if (empty _1)
		   '()
		   (list (fn (head _1)) (tail _1)))))
	 (gen iter lst))

(def map' (fn lst) :list
	 (set
	   iter
	   (lambda
		 (if (empty _1)
		   '()
		   (list (fn (head _1)) (tail _1)))))
	 (gen' iter lst))

;; take first n values from list
(def take (n lst) :list
	 (set
	   iter
	   (lambda
		 (do
		   (set cn (head _1))
		   (set cl (head (tail _1)))
		   (if (or (= cn 0) (empty cl))
			 '()
			 (list (head cl) (list (- cn 1) (tail cl)))))))
	 (gen iter (list n lst)))

; (def take (n lst) (take n lst '()))
; (def take (0 lst acc) acc)
; (def take (n lst acc) (take (- n 1) (tail lst) (append acc (head lst))))


;; take elements from list while condition is true
(def take-while (pred lst) :list
	 (set
	   iter
	   (lambda
		 (if (or (empty _1) (not (pred (head _1))))
		   '()
		   (list (head _1) (tail _1)))))
	 (gen iter lst))


;; drop first n values from list
(def drop (0 lst) lst)
(def drop (n lst) (drop (- n 1) (tail lst)))


;; take nth element from list.
;; Elements numeration is started from 1 (!).
(def nth (n lst) (head (drop (- n 1) lst)))


(def first  (lst) (nth 1 lst))
(def second (lst) (nth 2 lst))
(def third  (lst) (nth 3 lst))


;; reduce
(def reduce (fn '() acc) acc)
(def reduce (fn lst acc) (reduce fn (tail lst) (fn (head lst) acc)))


;; lazy concat
(def concat lists
	 ; return one element at the time
	 (set iter
		  (lambda
			(set h (first _1))
			(set t (second _1))
			(if (empty h)
			  (if (empty t)
				'()
				(self (list (head t) (tail t))))
			  (list (head h) (list (tail h) t)))))
	 (gen iter (list (head lists) (tail lists))))


;; length
(def length (lst) (length lst 0))
(def length ('() acc) acc)
(def length (lst acc) (length (tail lst) (+ acc 1)))
