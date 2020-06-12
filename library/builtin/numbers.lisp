; int operations
(def + (args:args[int]) :int (apply int.plus args) :int)
(def - (args:args[int]) :int (apply int.minus args) :int)
(def * (args:args[int]) :int (apply int.mult args) :int)
(def / (args:args[int]) :int (apply int.div args) :int)

; float operations
(def + (args:args[float]) :float (apply float.plus args) :float)
(def - (args:args[float]) :float (apply float.minus args) :float)
(def * (args:args[float]) :float (apply float.mult args) :float)
(def / (args:args[float]) :float (apply float.div args) :float)
