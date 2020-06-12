(use plugin "drawer")

; RGB color
(deftype :rgb :list[int])

(def rgb (r:int g:int b:int) :rgb
	 (list r g b) :rgb)

(def red   (c:rgb) :int (head c))
(def green (c:rgb) :int (head (tail c)))
(def blue  (c:rgb) :int (head (tail (tail c))))

; Drawer
(def drawer.new (name:str width:int height:int) :drawer
	 (drawer.new.native name width height) :drawer)

; Draw point
(def draw.point (d:drawer x:int y:int r:int g:int b:int)
	 (draw.point.native d x y r g b))

(def draw.point (d:drawer x:int y:int c:rgb)
	 (draw.point.native d x y (red c) (green c) (blue c)))

; Draw line
(def draw.line (d:drawer x0:int y0:int x1:int y1:int r:int g:int b:int)
	 (draw.line.native d x0 y0 x1 y1 r g b))

(def draw.line (d:drawer x0:int y0:int x1:int y1:int c:rgb)
	 (draw.line.native d x0 y0 x1 y1 (red c) (green c) (blue c)))
