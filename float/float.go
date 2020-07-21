package float

import "math"

type fishFloat struct {
	ID int
	x  float64
	y  float64
	z  float64
}

var float fishFloat

//Distance get the distance from given pos to the float
func Distance(x, z float64) float64 {
	x0 := float.x - x
	z0 := float.z - z
	return math.Sqrt(x0*x0 + z0*z0)
}

//Set set Float id and pos when new float was spawn
func Set(ID int, x, y, z float64) {
	float = fishFloat{
		ID: ID,
		x:  x,
		y:  y,
		z:  z,
	}
}

//IsMine judge the belonging of the float
func IsMine(EID int) bool {
	return EID == float.ID
}

//Move update the pos of the float with deltaXYZ
func Move(DeltaX, DeltaY, DeltaZ int) {
	float.x += float64(DeltaX) / 4096
	float.y += float64(DeltaY) / 4096
	float.z += float64(DeltaZ) / 4096
}
