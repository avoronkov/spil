(use std)

; Lazy read from file

(set' file (open "examples/testdata.numbers.txt"))

(print file)

(def line->numbers (l:str) :list
	 (map int (words l)))    

(set nums (map line->numbers (lines file)))

(print "numbers:" nums)
