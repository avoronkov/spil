(def inc (n) (+ n 1))
(def dec (n) (- n 1))

(def what (1) (print "what" 1) 1)
(def what (n) (print "what" n) (inc (what (dec n))))

(def' what' (1) (print "what'" 1) 1)
(def' what' (n) (print "what'" n) (inc (what' (dec n))))

(def lmap (fn lst) (lmap fn lst '()))
(def lmap (fn '() acc) acc)
(def lmap (fn lst acc) (lmap fn (tail lst) (append acc (fn (head lst)))))

(print (lmap what  '(1 2 3 4 5)))
(print (lmap what' '(1 2 3 4 5)))
