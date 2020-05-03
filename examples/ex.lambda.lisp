(func plus-n (n) (lambda (+ n _1)))

(set plus-four (plus-n 4))

(print (plus-four 3))

(set ab (lambda (* _1 _2)))
(print "3 * 4 =" (ab 3 4))
