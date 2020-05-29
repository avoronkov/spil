(def partition (l:list[a] pivot:a) :list (partition l pivot (do '() :list[a]) (do '() :list[a])))
(def partition ('() pivot:a lo:list[a] hi:list[a]) :list[any] (list lo hi))

(def partition (l:list[a] pivot:a lo:list[a] hi:list[a]) :list[any]
	 (set h (head l))
	 (set t (tail l))
	 (if (< h pivot)
	   (partition t pivot (do (append lo h) :list[a]) hi)
	   (partition t pivot lo (do (append hi h) :list[a]))) :list[any])

(def conc (l1:list[a] '()) :list[a] l1)
(def conc (l1:list[a] l2:list[a]) :list[a]
	 (conc (do (append l1 (head l2)) :list[a]) (tail l2)))

(def len ('()) :int 0)
(def len (l:list[a]) :int (+ 1 (len (tail l))))

(def fst (l:list[a]) :a (head l))

(def snd (l:list[a]) :a (head (tail l)))

(def sort (l:list[a]) :list[a]
	 (if (<= (len l) 1)
	   l
	   (do
		 (set parts (partition (tail l) (head l)))
		 (set lo (fst parts) :list[a])
		 (set hi (snd parts) :list[a])
		 (conc (do (append (sort lo) (head l)) :list[a]) (sort hi)))) :list[a])


(set l '(5 13 2 8 3 1) :list[int])
(print (partition l 4))

(print (sort l))
; (print (sort '(1 2 3 "hello")))
