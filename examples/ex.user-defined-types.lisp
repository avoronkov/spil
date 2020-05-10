;; simple set implementation with type-checking

(def equal (a:any a:any) :bool 'T)
(def equal (a:any b:any) :bool 'F)


(def concat' (l:list '()) :list l)
(def concat' (l:list r:list) :list
	 (concat' (append l (head r)) (tail r)))

(def set-new () :set '() :set)

(def set-add (s:set e:any) :set (set-add '() (do s :list) e))
(def set-add (left:list '() e:any) :set (append left e) :set)
(def set-add (left:list right:list e:any) :set
	 (set h (head right))
	 (if (equal h e)
	   (do (concat' left right) :set)
	   (set-add (append left h) (tail right) e)))

(set s (set-new))

(set s1 (set-add s 1))
(set s2 (set-add s1 2))
(set s3 (set-add s2 2))
(set s4 (set-add s3 3))


(print (head (do s1 :list)))
(print "set:" s4)
