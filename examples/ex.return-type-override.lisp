; polymophic function with different return types

(def add (x:int y:int) :int (+ x y))
(def add (x:str y:str) :str (append x y) :str)

(def expect-int (a:int) :int a)
(def expect-str (s:str) :str s)

(print (expect-int (add 3 4)))
; 7

(print (expect-str (add "hello" " world")))
; "hello world"
