(use std)

; map over simple may
(set x2 (lambda (* _1 2)))
(print (map x2 '(1 2 3 4 5)))

; map over lazy may
(def below-6 (5) '())
(def below-6 (n) (list (+ n 1)))
(print (map x2 (gen below-6 0)))

; map over endless list
(def next-int (n) (list (+ n 1)))
(print (head (map x2 (gen next-int 0))))
