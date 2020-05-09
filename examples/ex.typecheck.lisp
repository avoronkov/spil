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
