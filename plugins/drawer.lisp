(use drawer/drawer)

(def drawer.new (name:str width:int height:int) :drawer
	 (drawer.new.native name width height) :drawer)

(def draw.point (d:drawer x:int y:int r:int g:int b:int)
	 (draw.point.native d x y r g b))

(def draw.line (d:drawer x0:int y0:int x1:int y1:int r:int g:int b:int)
	 (draw.line.native d x0 y0 x1 y1 r g b))
