package common

type Vec2 struct {
	x int
	y int
}

func NewVec2(x, y int) Vec2 {
	return Vec2{x: x, y: y}
}
func (v *Vec2) Equal(to Vec2) bool {
	return v.x == to.x && v.y == to.y
}

func (v *Vec2) IsZero() bool {
	return v.x == 0 && v.y == 0
}

func (v *Vec2) Copy() Vec2 {
	return NewVec2(v.x, v.y)
}

func (v *Vec2) Set(x, y int) {
	v.x = x
	v.y = y
}

func (v *Vec2) MultConst(a int) {
	v.x *= a
	v.y *= a
}

func (v *Vec2) Add(a Vec2) {
	v.x += a.x
	v.y += a.y
}
func (v *Vec2) AddConst(a int) {
	v.x += a
	v.y += a
}

func (v *Vec2) Get() (int, int) {
	return v.x, v.y
}
