(contract :t)

(def min (x:t y:t) :t (if (< x y) x y))

(print "min(3, 5) = " (min 3 5))
