(def id (x:a) :a x)

(def id2 (x:a) :a x :a)

(def id3 (x:a) :a (do x :a))

(def id4 (x:a) :a (set y x) y)

(def id5 (x:a) :a (set y x :a) y)

(def expect-int (i:int) :str "ok")

(print "id:" (expect-int (id 13)))
(print "id2:" (expect-int (id2 14)))
(print "id3:" (expect-int (id3 15)))
(print "id4:" (expect-int (id4 16)))
(print "id5:" (expect-int (id5 17)))
