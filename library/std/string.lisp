;; split string into words
(def next-string (sep:str str:str) :list (next-string sep str ""))
(def next-string (sep:str '() "") :list '())
(def next-string (sep:str '() acc:str) :list (list acc ""))
(def next-string (sep:str str:str acc:str) :list
	 (set is-sep (sep (head str)) :bool)
	 (if is-sep
	   (if (= acc "")
		 (next-string sep (do (tail str) :str) acc)
		 (list acc (tail str)))
	   (next-string sep (do (tail str) :str) (do (append acc (head str)) :str))))


;; split string into lazy list of words
(def words (str:str) :list  (gen  \(next-string space _1) str))
(def words' (str:str) :list (gen' \(next-string space _1) str))

;; split string into lazy list of lines
(def lines (str:str) :list  (gen  \(next-string eol _1) str))
(def lines' (str:str) :list (gen' \(next-string eol _1) str))
