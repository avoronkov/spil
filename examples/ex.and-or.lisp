(print "or  true " (if (or (< 2 1) (= 3 3)) "ok" "failed"))
(print "or  false" (if (or (< 2 1) (= 3 4) 'F) "failed" "ok"))
(print "and true " (if (and (= 1 1) (> 2 0) 'T) "ok" "failed"))
(print "and false" (if (and (= 1 1) (< 2 0) 'T) "failed" "ok"))
