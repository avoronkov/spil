(def f (0) (print "f:" 0) 0)
(def f (n) (print "f:" n) (f (- n 1)))
(def f () (f 3))

(print (f))
