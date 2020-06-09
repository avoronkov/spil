; typed function with variable number of arguments

(def print-ints (msg:str args:args[int]) :any
	 (if (empty args)
	   '()
	   (do
		 (print msg (head args))
		 (print-ints msg (do (tail args) :args[int])))))



(print-ints "arg:" 1 2 3 5)
