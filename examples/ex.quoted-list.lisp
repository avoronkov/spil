# Quoted list example

(print '(1 2 3 ) )

(set lst '((1 2 3) (4 5 6) (7 8 9)))
(print lst)

(set first-item (head lst) :list)
(print (head (tail first-item)))
