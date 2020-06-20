;; parse symbolic expressions
(set s "\"hello\" 1 2 3 (4 5 6)")

(set l (parse s))

(print "string:" l (type (head l)))

(set' f (open "examples/testdata.expr.txt"))
(set fl (parse f))

(print "file:" fl (type (head fl)))
