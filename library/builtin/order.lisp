(def < (a:int b:int) :bool (native.int.less a b))
(def > (a:int b:int) :bool (native.int.less b a))
(def <= (a:int b:int) :bool (not (native.int.less b a)))
(def >= (a:int b:int) :bool (not (native.int.less a b)))

(def < (a:str b:str) :bool (native.str.less a b))
(def > (a:str b:str) :bool (native.str.less b a))
(def <= (a:str b:str) :bool (not (native.str.less b a)))
(def >= (a:str b:str) :bool (not (native.str.less a b)))
