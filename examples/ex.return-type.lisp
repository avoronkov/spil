(def foo () :int 1)
(def bar () :int (foo))
(def baz (x:int) :int x)
(def owl (x:int) :int (set res x) res)

(def len (lst:list) :int (len lst 0))
(def len ('() acc:int) :int acc)
(def len (lst:list acc:int) :int (len (tail lst) (+ acc 1)))

(print (foo))
(print (bar))
(print (baz 13))
(print (owl 25))

(print (len '(1 2 3 4 5)))
