(contract :b)

(def reduce (fn:func[a,b,b] '() acc:b) :b acc)
(def reduce (fn:func[a,b,b] l:list[a] acc:b) :b
	 (reduce fn (tail l) (fn (head l) acc)))

(def add (x:int y:int) :int (+ x y))

(set l '(1 2 3 4 5) :list[int])

(print "reduce add (1 2 3 4 5) 0 =" (reduce add l 0))
