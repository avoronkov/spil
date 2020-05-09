(func expect-int (n:int) :str "ok")

(func get-int () :any 14)

(print 13 (expect-int 13))

; This should fail
; (print 14 (expect-int (get-int)))

(set n (get-int) :int)
(print "(set n (get-int) :int)" (expect-int n))

(print "(do (get-int) :int)" (expect-int (do (get-int) :int)))

(def return-int () :int (get-int) :int)

(print "(def return-int () :int (get-int) :int)" (expect-int (return-int)))
