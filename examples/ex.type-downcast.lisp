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
