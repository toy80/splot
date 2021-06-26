package main

import (
	"fmt"
	"math"

	"github.com/toy80/splot"
)

func normalize3(v [3]float32) [3]float32 {
	a := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	return [3]float32{v[0] / a, v[1] / a, v[2] / a}
}

func cross(a, b [3]float32) [3]float32 {
	return [3]float32{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

func normalize4(v [4]float32) [4]float32 {
	a := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])))
	return [4]float32{v[0] / a, v[1] / a, v[2] / a, v[3] / a}
}

func rotate(q [4]float32, v [3]float32) [3]float32 {
	a := [3]float32{q[0], q[1], q[2]}
	t := cross(a, v)
	t[0], t[1], t[2] = t[0]*2, t[1]*2, t[2]*2
	s := [3]float32{t[0] * q[3], t[1] * q[3], t[2] * q[3]}
	u := cross(a, t)
	return [3]float32{v[0] + s[0] + u[0], v[1] + s[1] + u[1], v[2] + s[2] + u[2]}
}

func quaterion(axis [3]float32, angle float32) [4]float32 {
	a := normalize3(axis)
	sinα, cosα := math.Sincos(float64(angle) * 0.5)
	return [4]float32{a[0] * float32(sinα), a[1] * float32(sinα), a[2] * float32(sinα), float32(cosα)}
}

func main() {

	// prepare the quaternion i*x + j*y + k*z + w
	q := normalize4([4]float32{1, 2, 3, 0.3}) //
	v0 := [3]float32{1, 0, 0}                 // initial vector

	var plot splot.Plot
	plot.Title(fmt.Sprintf("Quaternion %fi + %fj + %fk + %f", q[0], q[1], q[2], q[3]))
	plot.Width(1)
	plot.Line([3]float32{-1.2, 0, 0}, [3]float32{1.2, 0, 0}).Color("#FF80C0").FilledHead() // X axis
	plot.Line([3]float32{0, -1.2, 0}, [3]float32{0, 1.2, 0}).Color("#C0FF80").FilledHead() // Y axis
	plot.Line([3]float32{0, 0, -1.2}, [3]float32{0, 0, 1.2}).Color("#80C0FF").FilledHead() // Z axis
	plot.MakeCircle([3]float32{}, [3]float32{1, 0, 0}, 1, splot.StdColor(0))               // YZ plane
	plot.MakeCircle([3]float32{}, [3]float32{0, 1, 0}, 1, splot.StdColor(0))               // ZX plane
	plot.MakeCircle([3]float32{}, [3]float32{0, 0, 1}, 1, splot.StdColor(0))               // XY plane

	axis := normalize3([3]float32{q[0], q[1], q[2]})
	plot.Line([3]float32{}, axis).Color("purple").Text("axis").NoHead().Width(2)

	plot.Line([3]float32{}, normalize3(v0)).Color("black").Text("v0").FilledHead().Width(2)
	v1 := rotate(q, normalize3(v0))
	plot.Line([3]float32{}, v1).Color("black").Text("v1").FilledHead().Width(2)

	// geneate the trajectory by construct searial of quaternions

	maxAngle := 2 * math.Acos(float64(q[3]))
	const deltaAngle = 0.1
	n := int((maxAngle) / deltaAngle)
	if n < 1 {
		n = 1
	}
	da := maxAngle / float64(n)
	for i := 0; i <= n; i++ {
		q1 := quaterion([3]float32{q[0], q[1], q[2]}, float32(da)*float32(i))
		v := rotate(q1, normalize3(v0))
		if i == 0 {
			plot.MoveTo(v).Color("red").NoHead()
		} else {
			plot.LineTo(v)
		}
	}

	plot.WriteFile("quaternion.plt") // the file can be open with a gnuplot viewer
}
