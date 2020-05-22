(def id (x:a) :a x)

(def expect-int (i:int) :str "ok")

(print (expect-int (id 13)))
