;; storing value into list preserves its type

(deftype :set[a] :list[a])

(set s (list (do '(1 2 3) :set[int])))

(print (type (head s)))
; :set[int]
(print (head s))
; '(1 2 3)
