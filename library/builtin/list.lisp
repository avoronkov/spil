(contract :a)

(def head (l:list[a]) :a (native.head l) :a)
 
(def tail (l:list[a]) :list[a] (native.tail l) :list[a])

; (def append (l:list[a] x:a) :list[a] (native.append l x) :list[a])

; (def append (l:list[a] x:b) :list[any] (native.append l x) :list[any])

; (def flist args :list args)
