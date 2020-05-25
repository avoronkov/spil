(def thead (l:list[t]) :t (head l) :t)

(def what-type (i:int) :str "int")
(def what-type (s:str) :str "str")
(def what-type (x:any) :str "any")

(set intlist '(1 2 3 4) :list[int])
(set strlist '("foo" "bar" "baz") :list[str])

(print (what-type (thead intlist)))
(print (what-type (thead strlist)))
(print (what-type (thead '(1 2 3 "4" "foo"))))
