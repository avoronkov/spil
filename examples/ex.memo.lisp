(def what (1) (print "what" 1) 1)
(def what (n) (print "what" n) (inc (what (dec n))))

(def' what' (1) (print "what'" 1) 1)
(def' what' (n) (print "what'" n) (inc (what' (dec n))))

(print (map what  '(1 2 3 4 5)))
(print (map what' '(1 2 3 4 5)))
