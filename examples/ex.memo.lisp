(def inc (n) (+ n 1))
(def dec (n) (- n 1))

(def what (1) (print "what" 1) 1)
(def what (n) (print "what" n) (inc (what (dec n))))

(def' what' (1) (print "what'" 1) 1)
(def' what' (n) (print "what'" n) (inc (what' (dec n))))

(def map (fn lst) (map fn lst '()))
(def map (fn '() acc) acc)
(def map (fn lst acc) (map fn (tail lst) (append acc (fn (head lst)))))

(print (map what  '(1 2 3 4 5)))
(print (map what' '(1 2 3 4 5)))
