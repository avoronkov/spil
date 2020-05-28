#!/usr/bin/env spil

(def expect-int (i) (< i 10))

(set l '(1 2 3 4 5))
(print (type (head l)))

(print (expect-int (do (head l) :any)))

(set x (head l) :any)
(print (expect-int x))

(def get-int() :any 13)

(print (type (get-int)))
(print (expect-int (get-int)))

(def get-int2() :any 25 :any)

(print (type (get-int2)))
(print (expect-int (get-int2)))

(def thead (l:list[a]) :a (head l) :a)

(print "thead:" (type (thead l)))
(print (expect-int (thead l)))
