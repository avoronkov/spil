;; compare integers
(def < (a:int b:int) :bool (int.less a b))
(def > (a:int b:int) :bool (int.less b a))
(def <= (a:int b:int) :bool (not (int.less b a)))
(def >= (a:int b:int) :bool (not (int.less a b)))

;; compare strings
(def < (a:str b:str) :bool (str.less a b))
(def > (a:str b:str) :bool (str.less b a))
(def <= (a:str b:str) :bool (not (str.less b a)))
(def >= (a:str b:str) :bool (not (str.less a b)))

;; compare floats
(def < (a:float b:float) :bool (float.less a b))
(def > (a:float b:float) :bool (float.less b a))
(def <= (a:float b:float) :bool (not (float.less b a)))
(def >= (a:float b:float) :bool (not (float.less a b)))
