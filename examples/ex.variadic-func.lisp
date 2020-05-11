(def sum () :int 0)
(def sum args :int (+ (do (head args) :int) (apply sum (tail args))))

(print (sum 1 2 3 4))
