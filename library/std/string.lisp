;; split string into words
(def next-string (sep:func str:str) :list (next-string sep str ""))
(def next-string (sep:func '() "") :list '())
(def next-string (sep:func '() acc:str) :list (list acc ""))
(def next-string (sep:func str:str acc:str) :list
	 (set is-sep (sep (head str)) :bool)
	 (set tl (do (tail str) :str))
	 (if is-sep
	   (if (= acc "")
		 (next-string sep tl acc)
		 (list acc tl))
	   (next-string sep tl (do (append acc (head str)) :str))))


;; split string into lazy list of words
(def words (str:str) :list[str]  (gen  \(next-string space _1) str) :list[str])
(def words' (str:str) :list[str] (gen' \(next-string space _1) str) :list[str])

;; split string into lazy list of lines
(def lines (str:str) :list[str]  (gen  \(next-string eol _1) str) :list[str])
(def lines' (str:str) :list[str] (gen' \(next-string eol _1) str) :list[str])
