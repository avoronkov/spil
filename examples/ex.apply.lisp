(def ints () '(1 2 3 4))
(print (apply + (ints)))

(set ints2 (gen (lambda (if (> _1 1) (list (- _1 1)) '())) 5))
(print (apply + ints2))
