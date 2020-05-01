; lazy map


(def lazy-map (fn lst) (lazy-map fn lst '()))
(def lazy-map (fn '() acc) acc)
(def lazy-map (fn lst acc)
	 (lazy-map fn (tail lst) (append acc (fn (head lst)))))


(def x2 (x) (* x 2))

(def below-5 (5) '())
(def below-5 (n) (append '() (+ n 1)))

(set n (gen below-5 0))

; (print n)
(print (lazy-map x2 '(1 2 3 4 5)))
(print (lazy-map x2 n))
