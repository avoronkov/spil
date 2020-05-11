;; simple set implementation with type-checking

(def equal (a:any a:any) :bool 'T)
(def equal (a:any b:any) :bool 'F)


(def concat' (l:list '()) :list l)
(def concat' (l:list r:list) :list
	 (concat' (append l (head r)) (tail r)))

(deftype :set :list)

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

(def my-head (l:list) :any (head l))

(print (head s1))
(print "set:" s4)

(deftype :positive :int)
(set xp (do 13 :positive))

(print (type xp))
(print (+ xp 12))
(print (mod xp 5))

(deftype :mystring :str)
(set ms (do " " :mystring))
(print (space ms))
(print (append ms "OK"))
