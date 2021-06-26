// Package splot provide a simple method to convert lines and points to gnuplot.
package splot

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
)

var colorTable = []string{
	"#808080",
	"#000000",
	"#C20A0A",
	"#C2780A",
	"#9DC20A",
	"#2FC20A",
	"#0AC253",
	"#0AC2C2",
	"#0A53C2",
	"#2F0AC2",
	"#9D0AC2",
	"#C20A78",
}

func StdColor(idx int) string {
	if idx < 0 {
		idx = len(colorTable) - idx
	}
	return colorTable[idx%len(colorTable)]
}

func abs(x float32) float32 {
	return float32(math.Abs(float64(x)))
}

// Vec3 is 3D vector
type Vec3 = [3]float32

func normalize(v Vec3) Vec3 {
	a2 := v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
	if a2 <= math.SmallestNonzeroFloat32*math.SmallestNonzeroFloat32 {
		return Vec3{1, 0, 0}
	}
	a := float32(math.Sqrt(float64(a2)))
	return Vec3{v[0] / a, v[1] / a, v[2] / a}
}

func dot(v, b Vec3) float32 { return v[0]*b[0] + v[1]*b[1] + v[2]*b[2] }

func cross(v, b Vec3) Vec3 {
	return Vec3{v[1]*b[2] - v[2]*b[1], v[2]*b[0] - v[0]*b[2], v[0]*b[1] - v[1]*b[0]}
}

type mat3 = [3][3]float32

func mat3MulVec3(m mat3, a [3]float32) (b [3]float32) {
	b[0] = a[0]*m[0][0] + a[1]*m[1][0] + a[2]*m[2][0]
	b[1] = a[0]*m[0][1] + a[1]*m[1][1] + a[2]*m[2][1]
	b[2] = a[0]*m[0][2] + a[1]*m[1][2] + a[2]*m[2][2]
	return
}

type style struct {
	color     string
	width     int    // 1, 2, 3 ...
	lineAttr  string // "filled head" "nohead" etc.
	pointAttr string //
	isPoint   bool   // 考虑到有的线段长度为零

	key string
	cv  int
}

type primitive struct {
	p0   Vec3   // 端点坐标
	p1   Vec3   // 矢量方向
	text string // 标注
	style
}

func (p *style) prepareStyleKey() {
	if p.color == "" {
		p.color = "black"
	}
	var c, a string
	if p.isPoint {
		if p.width <= 0 {
			p.width = 3
		}
		a = p.pointAttr
	} else {
		if p.width <= 0 {
			p.width = 1
		}
		if p.lineAttr == "" {
			p.lineAttr = "nohead"
		}
		a = p.lineAttr
	}
	c = p.color

	// isPoint  width  color | attr
	n := 1 + 2 + len(c) + 1 + len(a)
	buf := make([]byte, n)
	if p.isPoint {
		buf[0] = 'P'
	} else {
		buf[0] = 'L'
	}
	buf[1] = byte('0' + p.width/10%10)
	buf[2] = byte('0' + p.width%10)
	copy(buf[3:3+len(c)], c)
	buf[3+len(c)] = '|'
	copy(buf[4+len(c):], a)
	p.key = string(buf)
}

func (p *primitive) dir() Vec3 {
	if p.isPoint {
		return Vec3{}
	}
	return Vec3{p.p1[0] - p.p0[0], p.p1[1] - p.p0[1], p.p1[2] - p.p0[2]}
}

func (p *primitive) mid() Vec3 {
	if p.isPoint {
		return p.p0
	}
	return Vec3{
		(p.p1[0] + p.p0[0]) * 0.5,
		(p.p1[1] + p.p0[1]) * 0.5,
		(p.p1[2] + p.p0[2]) * 0.5,
	}
}

type Plot struct {
	title string
	prims []primitive
	dummy primitive
	start bool
}

func (p *Plot) last() *primitive {
	if len(p.prims) == 0 || p.start {
		return &p.dummy
	}
	return &p.prims[len(p.prims)-1]
}

func (p *Plot) break_() *primitive {
	if p.start || len(p.prims) == 0 {
		return &p.dummy
	}
	last := p.last()
	p.start = true
	p.dummy = *last
	return &p.dummy
}

func (p *Plot) new() *primitive {
	last := p.last()
	p.start = false
	p.prims = append(p.prims, *last)
	cur := &p.prims[len(p.prims)-1]
	cur.text = ""
	cur.p0, cur.p1, cur.isPoint = last.p1, Vec3{}, true
	return cur
}

func (p *Plot) Title(s string) *Plot {
	p.title = s
	return p
}

func (p *Plot) Break() *Plot {
	p.break_()
	return p
}

func (p *Plot) CurPos() Vec3 {
	last := p.last()
	if last.isPoint {
		return last.p0
	}
	return last.p1
}

func (p *Plot) MoveTo(pt Vec3) *Plot {
	last := p.break_()
	if last.isPoint {
		last.p0 = pt
	} else {
		last.p1 = pt
	}
	return p
}

func (p *Plot) Point(pt Vec3) *Plot {
	v := p.new()
	v.p0, v.p1, v.isPoint = pt, pt, true
	return p
}

func (p *Plot) Line(pt0, pt1 Vec3) *Plot {
	v := p.new()
	v.p0, v.p1, v.isPoint = pt0, pt1, false
	return p
}

func (p *Plot) LineTo(pt Vec3) *Plot {
	return p.Line(p.CurPos(), pt)
}

func (p *Plot) Vector(pt0, dir Vec3) *Plot {
	v := p.new()
	pt1 := Vec3{pt0[0] + dir[0], pt0[1] + dir[1], pt0[2] + dir[2]}
	v.p0, v.p1, v.isPoint = pt0, pt1, false
	return p
}

func (p *Plot) MakeCircle(center, normal Vec3, radius float32, color string) {
	normal = normalize(normal)
	var tangent, bitangent Vec3
	tangent = Vec3{1, 0, 0}
	if abs(dot(normal, tangent)) < 0.9 {
		bitangent = normalize(cross(normal, tangent))
		tangent = normalize(cross(bitangent, normal))
	} else {
		bitangent = Vec3{0, 1, 0}
		tangent = normalize(cross(bitangent, normal))
		bitangent = normalize(cross(normal, tangent))
	}
	rot := mat3{tangent, bitangent, normal}

	const N = 50
	delta := float32(2 * math.Pi / N)
	a := float32(0)
	var pts []Vec3
	for i := 0; i < N; i++ {
		sinA, cosA := math.Sincos(float64(a))
		v := Vec3{radius * float32(cosA), radius * float32(sinA), 0}
		v = mat3MulVec3(rot, v)
		pt := Vec3{center[0] + v[0], center[1] + v[1], center[2] + v[2]}
		pts = append(pts, pt)
		a += delta
	}
	for i, j := N-1, 0; j < N; i, j = j, j+1 {
		p.Line(pts[i], pts[j])
		if j == 0 {
			p.NoHead().Color(color)
		}
	}
}

func (p *Plot) LineDir(pt0, pt1 Vec3) *Plot {
	return p.Vector(pt0, pt1)
}

func (p *Plot) Attr(attr string) *Plot {
	last := p.last()
	if last.isPoint {
		last.pointAttr = attr
	} else {
		last.lineAttr = attr
	}
	return p
}

// FilledHead ...
func (p *Plot) FilledHead() *Plot {
	p.last().lineAttr = "filled head"
	return p
}

// NoHead ..
func (p *Plot) NoHead() *Plot {
	p.last().lineAttr = "nohead"
	return p
}

func (p *Plot) Text(text string) *Plot {
	p.last().text = text
	return p
}

func (p *Plot) Textf(format string, v ...interface{}) *Plot {
	p.last().text = fmt.Sprintf(format, v...)
	return p
}

// Color 设置当前图元的颜色. 例如 "#RRGGBB", "red", "green", "blue"
func (p *Plot) Color(cr string) *Plot {
	p.last().color = cr
	return p
}

// StdColor 设置当前图元的颜色
func (p *Plot) StdColor(idx int) *Plot {
	p.last().color = StdColor(idx)
	return p
}

func (p *Plot) Width(w int) *Plot {
	if w < 1 {
		w = 1
	}
	if w > 99 {
		w = 00
	}
	p.last().width = w
	return p
}

func (p *Plot) foreachPrim(s style, do func(*primitive)) {
	for i := range p.prims {
		if p.prims[i].key == s.key {
			do(&p.prims[i])
		}
	}
}

func (p *Plot) WriteFile(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}

	defer func() {
		err1 := f.Close()
		if err == nil {
			err = err1
		}
	}()

	err = p.Out(f)
	return
}

func (p *Plot) Out(f io.Writer) (err error) {

	// write common properties
	if p.title != "" {
		fmt.Fprintf(f, "set title %q\n", p.title)
	}
	fmt.Fprintln(f, `set view equal xyz`)
	fmt.Fprintln(f, `unset key`)
	if len(p.prims) == 0 {
		return
	}

	// style grouping
	var styleList []style
	{
		tmpStyleMap := make(map[string]style)
		for i := range p.prims {
			p.prims[i].prepareStyleKey()
			key := p.prims[i].key
			if _, ok := tmpStyleMap[key]; ok {
				continue
			}
			tmpStyleMap[key] = p.prims[i].style
			styleList = append(styleList, p.prims[i].style)
		}
	}
	// draw lines before points
	sort.Slice(styleList, func(i, j int) bool {
		return styleList[i].key < styleList[j].key
	})

	// generate palette
	colorMap := make(map[string]int)
	colorValue := 0
	for i, style := range styleList {
		if v, ok := colorMap[style.color]; ok {
			styleList[i].cv = v
			continue
		}
		styleList[i].cv = colorValue
		colorMap[style.color] = colorValue
		colorValue++
	}
	type pal struct {
		color string
		value int
	}
	colorList := make([]pal, len(colorMap))
	for c, v := range colorMap {
		colorList[v] = pal{color: c, value: v}
	}

	// sort palette by value
	sort.Slice(colorList, func(i, j int) bool {
		return colorList[i].value < colorList[j].value
	})

	// write palete
	fmt.Fprintf(f, "set palette model RGB maxcolors %d\n", len(colorList))
	fmt.Fprint(f, `set palette defined (`)
	for i, c := range colorList {
		if i > 0 {
			fmt.Fprint(f, `, `)
		}
		fmt.Fprintf(f, "%d %q", c.value, c.color)
	}
	fmt.Fprintln(f, `)`)
	fmt.Fprintln(f, `# sets the range of palette values`)
	fmt.Fprintf(f, "set cbrange [-0.5:%.1f]\n", float32(len(colorList))-0.5)
	fmt.Fprintln(f)

	// separator use in data table
	sep := " "
	for i := range p.prims {
		p.prims[i].text = strings.ReplaceAll(p.prims[i].text, sep, "-")
	}

	// there are actually multiple plots:
	//   plot_1, label_1, plot_2, label_2, plot_3, label_3 ...

	first := true
	for _, style := range styleList {
		if first {
			first = false
			fmt.Fprint(f, `splot "-" `)
		} else {
			fmt.Fprint(f, " \\\n  , \"\" ")
		}
		if style.isPoint {
			// draw point
			fmt.Fprintf(f, `using 1:2:3:4 with points %s pointsize %d palette`, style.pointAttr, style.width)
		} else {
			// draw vector/line
			fmt.Fprintf(f, ` using 1:2:3:4:5:6:7 with vectors %s linewidth %d palette`, style.lineAttr, style.width)

		}
	}
	// draw label text
	fmt.Fprintf(f, ` , "" using 1:2:3:4 with labels`)

	fmt.Fprintln(f) // separate between gnuplot command an data tables

	for _, style := range styleList {
		if style.isPoint {
			// point data
			p.foreachPrim(style, func(v *primitive) {
				fmt.Fprint(f, v.p0[0], sep, v.p0[1], sep, v.p0[2], sep, style.cv)
				fmt.Fprintln(f)
			})

			fmt.Fprintln(f, "e") // separate between data tables

		} else {
			// vector/line data
			p.foreachPrim(style, func(v *primitive) {
				d := v.dir()
				fmt.Fprint(f, v.p0[0], sep, v.p0[1], sep, v.p0[2], sep, d[0], sep, d[1], sep, d[2], sep, style.cv)
				fmt.Fprintln(f)
			})

			fmt.Fprintln(f, "e") // separate between data tables
		}
	}
	// point label data
	for _, v := range p.prims {
		if v.text != "" {
			pt := v.mid()
			fmt.Fprint(f, pt[0], sep, pt[1], sep, pt[2], sep, v.text)
			fmt.Fprintln(f)
		}
	}
	fmt.Fprintln(f, "e")

	// allow ineractive op
	fmt.Fprintln(f, "pause mouse keypress")

	return
}
