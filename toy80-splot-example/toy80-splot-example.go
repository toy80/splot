package main

import (
	"fmt"
	"math"

	"github.com/toy80/splot"
)

type (
	Vec3 = [3]float32
	Vec4 = [4]float32
)

func sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func normalize3(v Vec3) Vec3 {
	a2 := v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
	if a2 <= math.SmallestNonzeroFloat32*math.SmallestNonzeroFloat32 {
		return Vec3{1, 0, 0}
	}
	a := sqrt(a2)
	return Vec3{v[0] / a, v[1] / a, v[2] / a}
}

func dot(a, b Vec3) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func cross(a, b Vec3) Vec3 {
	return Vec3{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

func length(v Vec3) float32 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

func normalize4(v Vec4) Vec4 {
	a := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])))
	return Vec4{v[0] / a, v[1] / a, v[2] / a, v[3] / a}
}

func rotate(q Vec4, v Vec3) Vec3 {
	a := Vec3{q[0], q[1], q[2]}
	t := cross(a, v)
	t[0], t[1], t[2] = t[0]*2, t[1]*2, t[2]*2
	s := Vec3{t[0] * q[3], t[1] * q[3], t[2] * q[3]}
	u := cross(a, t)
	return Vec3{v[0] + s[0] + u[0], v[1] + s[1] + u[1], v[2] + s[2] + u[2]}
}

func quaterion(axis Vec3, angle float32) Vec4 {
	a := normalize3(axis)
	sinα, cosα := math.Sincos(float64(angle) * 0.5)
	return Vec4{a[0] * float32(sinα), a[1] * float32(sinα), a[2] * float32(sinα), float32(cosα)}
}

func main() {

	// prepare the quaternion i*x + j*y + k*z + w
	q := normalize4(Vec4{-2, -1, 3, -0.5}) //
	v0 := Vec3{1, 0, 0}                    // initial vector, must be (1,0,0) currently

	origin := Vec3{}
	v0 = normalize3(v0)

	var plot splot.Plot
	plot.Title(fmt.Sprintf("Quaternion %+fi %+fj %+fk %+f", q[0], q[1], q[2], q[3]))

	plot.Width(3).Point(origin).Color("black").Text("O") // origin point

	plot.Width(1).FilledHead()
	plot.Line(Vec3{-1.2, 0, 0}, Vec3{1.2, 0, 0}).Color("#FF80C0")           // X axis
	plot.Line(Vec3{0, -1.2, 0}, Vec3{0, 1.2, 0}).Color("#C0FF80").Text("Y") // Y axis
	plot.Line(Vec3{0, 0, -1.2}, Vec3{0, 0, 1.2}).Color("#80C0FF").Text("Z") // Z axis

	plot.Break().StdColor(0).NoHead()
	//plot.Circle(origin, Vec3{1, 0, 0}, 1) // YZ plane
	//plot.Circle(origin, Vec3{0, 1, 0}, 1) // ZX plane
	plot.Circle(origin, Vec3{0, 0, 1}, 1) // XY plane

	axis := normalize3(Vec3{q[0], q[1], q[2]})
	plot.Line(Vec3{-axis[0], -axis[1], -axis[2]}, axis).Color("purple").Text("axis").FilledHead().Width(2)

	plot.Line(origin, v0).Color("black").Text("v0").FilledHead().Width(2)
	v1 := rotate(q, v0)
	plot.Line(origin, v1).Color("black").Text("v1").FilledHead().Width(2)

	maxAngle := 2 * math.Acos(float64(q[3]))
	sign := float32(1)
	if maxAngle > math.Pi {
		maxAngle -= 2 * math.Pi // to (-π, +π)
		sign = -1
	}
	if true {
		// geneate the trajectory with 3d geometry
		d := dot(v0, axis)
		o1 := Vec3{axis[0] * d, axis[1] * d, axis[2] * d}
		radius1 := float32(math.Sqrt(float64(1 - d*d)))
		plot.Break().Color("red").NoHead().Width(2).Arc(o1, axis, radius1, v0, 0, float32(maxAngle)).FilledHead()
		plot.Break().Color("pink").NoHead().Width(1).Arc(o1, axis, radius1, v0, float32(maxAngle), sign*math.Pi*2)
	} else {
		// geneate the trajectory by construct searial of quaternions
		const deltaAngle = 0.1
		n := int((maxAngle) / deltaAngle)
		if n < 1 {
			n = 1
		}
		da := maxAngle / float64(n)
		for i := 0; i <= n; i++ {
			q1 := quaterion(Vec3{q[0], q[1], q[2]}, float32(da)*float32(i))
			v := rotate(q1, v0)
			if i == 0 {
				plot.MoveTo(v).Color("red").NoHead()
			} else {
				plot.LineTo(v)
			}
		}
	}

	plot.WriteFile("quaternion.plt") // the file can be open with a gnuplot viewer
}
