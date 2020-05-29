(def apply-to-10 (f:func[int,int]) :int (f 10))

(def x2 (x:int) :int (* 2 x))

(print "2 * 10 =" (apply-to-10 x2))
