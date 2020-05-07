;; split string into words
(def next-string (sep str) (next-string sep str ""))
(def next-string (sep '() "") '())
(def next-string (sep '() acc) (list acc ""))
(def next-string (sep str acc)
	 (if (sep (head str))
	   (if (= acc "")
		 (next-string sep (tail str) acc)
		 (list acc (tail str)))
	   (next-string sep (tail str) (append acc (head str)))))


;; split string into lazy list of words
(def words (str)  (gen  \(next-string space _1) str))
(def words' (str) (gen' \(next-string space _1) str))

;; split string into lazy list of lines
(def lines (str)  (gen  \(next-string eol _1) str))
(def lines' (str) (gen' \(next-string eol _1) str))
