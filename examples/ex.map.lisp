; High order function map implementation 

(def map-aux (fn lst acc)
	 (if (= lst '())
	   acc
	   (map-aux fn (tail lst) (append acc (fn (head lst))))))

(def map (fn lst) (map-aux fn lst '()))

(def x2 (x) (* x 2))

(print (map x2 '(1 2 3 5 8)))

; vim: ft=lisp
