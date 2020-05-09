(def fn (x x) "same")
(def fn (x y) "differ")

(def fnt (x:int x:int) "same")
(def fnt (x:int y:int) "differ")

(print "differ any arg:" (fn 3 4))
(print "same any arg:" (fn 4 4))

(print "differ int arg:" (fnt 3 4))
(print "same int arg:"   (fnt 4 4))

(def fstr (x:str x:str) "same")
(def fstr (x:str y:str) "differ")

(print "differ str arg:" (fstr "3" "4"))
(print "same str arg:" (fstr "4" "4"))

(def fbool (x:bool x:bool) "same")
(def fbool (x:bool y:bool) "differ")

(print "differ bool arg:" (fbool 'T 'F))
(print "same bool arg:" (fbool 'T 'F))

(def return-int () :int (do 13))

(func xmap (fn:func l:list) :list (xmap fn l '()))
(func xmap (fn:func '() acc:list) :list acc)
(func xmap (fn:func l:list acc:list) :list
	  (xmap fn (tail l) (append acc (fn (head l)))))


(def x2 (n) (* n 2))
(print "xmap(func):" (xmap x2 '(1 2 3 5 8)))
(print "xmap(lambda):" (xmap \(+ _1 10) '(1 2 3 5 8)))
