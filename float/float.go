package float

import (
	"math"
)

type fishFloat struct {
	ID int
	pos
}
type analyser struct {
	pos         //上次位置
	h   float64 //连续下降高度
}
type pos struct{ x, y, z float64 }

var float fishFloat
var ana = &analyser{}

//Distance get the distance from given pos to the float
func Distance(x, z float64) float64 {
	x0 := float.x - x
	z0 := float.z - z
	return math.Sqrt(x0*x0 + z0*z0)
}

//Set set Float id and pos when new float was spawn
func Set(ID int, x, y, z float64) {
	float = fishFloat{
		ID:  ID,
		pos: pos{x: x, y: y, z: z},
	}
	ana.setPos(x, y, z)
}

//IsMine judge the belonging of the float
func IsMine(EID int) bool {
	return EID == float.ID
}

//IsFish update the pos of the float with deltaXYZ and judge if hook
func IsFish(DeltaX, DeltaY, DeltaZ int) bool {
	float.x += float64(DeltaX) / 4096
	float.y += float64(DeltaY) / 4096
	float.z += float64(DeltaZ) / 4096
	//	fmt.Println(float)
	return ana.AddPos()
}

func (a *analyser) setPos(x, y, z float64) {
	a = &analyser{
		pos: pos{x: x, y: y, z: z},
		h:   0,
	}
}

func (a *analyser) AddPos() bool {
	if Distance(a.x, a.z) > 0.6 {
		a.setPos(float.x, float.y, float.z)
	} else if a.y > float.y {
		a.h += a.y - float.y
	} else {
		a.h = 0
	}
	a.pos = float.pos
	if a.h > 0.3 {
		a.h = -1
		return true
	}
	return false
}
