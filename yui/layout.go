package yui

import (
	"math"
	"slices"
	"strings"
	"unicode/utf8"
	"yam/y3d"
)

type axis int

var (
	flat    []*UIElement = make([]*UIElement, 0, 1000)
	revFlat []*UIElement = make([]*UIElement, 0, 1000)
	s1      []*UIElement = make([]*UIElement, 0, 1000)
)

const (
	axisX axis = iota
	axisY
)

const eps = 0.01

var MeasureText = func(s string, font *Font) y3d.Vec2 {
	if font != nil {
		return y3d.Vec2{}
	}
	return y3d.Vec2{X: float32(utf8.RuneCountInString(s)) * float32(font.width), Y: float32(font.height)}
}

func (a axis) other() axis { return 1 - a }

func get(v y3d.Vec2, a axis) float32 {
	if a == axisX {
		return v.X
	}
	return v.Y
}

func set(v *y3d.Vec2, a axis, f float32) {
	if a == axisX {
		v.X = f
	} else {
		v.Y = f
	}
}

func (ui *UIElement) mainAxis() axis {
	if ui.Flag&UI_LAYOUT_TOP_TO_BOTTOM != 0 {
		return axisY
	}
	return axisX
}

func (ui *UIElement) padStart(a axis) float32 {
	if a == axisX {
		return ui.Padding.X
	}
	return ui.Padding.Z
}

func (ui *UIElement) pad(a axis) float32 {
	if a == axisX {
		return ui.Padding.X + ui.Padding.Y
	}
	return ui.Padding.Z + ui.Padding.W
}

func (ui *UIElement) gap() float32 {
	if len(ui.Children) < 2 {
		return 0
	}
	return float32((len(ui.Children) - 1) * ui.ChildGap)
}

func growFlag(a axis) uint32 {
	if a == axisX {
		return UI_GROW_X
	}
	return UI_GROW_Y
}

func (ui *UIElement) grows(a axis) bool { return ui.Flag&growFlag(a) != 0 }
func (ui *UIElement) resizable(a axis) bool {
	return ui.Flag&(UI_FIT|growFlag(a)) != 0 || (a == axisX && ui.Text != "")
}

func (ui *UIElement) max(a axis) float32 {
	if m := get(ui.MaxSize, a); m > 0 {
		return m
	}
	return math.MaxFloat32
}

func clamp(v, lo, hi float32) float32 { return min(max(v, lo), hi) }

func reset(ui *UIElement) {
	ui.Size = ui.BaseSize
	ui.minSize = ui.MinSize
	ui.Lines = nil
	if ui.Text != "" {
		full := MeasureText(ui.Text, ui.Font)
		ui.Size.X = max(ui.Size.X, full.X+ui.pad(axisX))
		ui.Size.Y = max(ui.Size.Y, full.Y+ui.pad(axisY))
		var longest float32
		for w := range strings.FieldsSeq(ui.Text) {
			longest = max(longest, MeasureText(w, ui.Font).X)
		}
		ui.minSize.X = max(ui.minSize.X, longest+ui.pad(axisX))
	}

	if !ui.resizable(axisX) {
		set(&ui.minSize, axisX, max(get(ui.minSize, axisX), get(ui.Size, axisX)))
	}
	if !ui.resizable(axisY) {
		set(&ui.minSize, axisY, max(get(ui.minSize, axisY), get(ui.Size, axisY)))
	}
}

func flatten(root *UIElement) {
	if root == nil {
		return
	}
	flat = flat[:0]
	revFlat = revFlat[:0]
	s1 = s1[:0]

	s1 = append(s1, root)
	for len(s1) > 0 {
		v := s1[len(s1)-1]
		s1 = s1[:len(s1)-1]
		flat = append(flat, v)
		s1 = append(s1, v.Children...)
	}
	revFlat = append(revFlat, flat...)
	slices.Reverse(revFlat)
}

func bfs(root *UIElement, apply func(ui *UIElement)) {
	if root == nil {
		return
	}
	for _, i := range flat {
		apply(i)
	}
}

func dfspo(root *UIElement, apply func(ui *UIElement)) {
	if root == nil {
		return
	}
	for _, i := range revFlat {
		apply(i)
	}
}

func fitSize(ui *UIElement, a axis) {
	if ui == nil || ui.Text != "" || ui.Flag&(UI_FIT|growFlag(a)) == 0 {
		return
	}
	var content, minContent float32
	if a == ui.mainAxis() {
		for _, c := range ui.Children {
			if c.Flag&UI_FLOATING != 0 {
				continue
			}
			content += get(c.Size, a)
			minContent += get(c.minSize, a)
		}
		content += ui.gap()
		minContent += ui.gap()
	} else {
		for _, c := range ui.Children {
			if c.Flag&UI_FLOATING != 0 {
				continue
			}
			content = max(content, get(c.Size, a))
			minContent = max(minContent, get(c.minSize, a))
		}
	}
	p := ui.pad(a)
	lo := max(get(ui.minSize, a), minContent+p)
	set(&ui.minSize, a, lo)
	set(&ui.Size, a, clamp(max(get(ui.Size, a), content+p), lo, ui.max(a)))
}
func FitSizeX(ui *UIElement) { fitSize(ui, axisX) }
func FitSizeY(ui *UIElement) { fitSize(ui, axisY) }

func grow(children []*UIElement, a axis, remaining float32) {
	var cands []*UIElement
	for _, c := range children {
		if c.Flag&UI_FLOATING != 0 {
			continue
		}
		if c.grows(a) && get(c.Size, a) < c.max(a) {
			cands = append(cands, c)
		}
	}
	for remaining > eps && len(cands) > 0 {
		smallest, second := float32(math.MaxFloat32), float32(math.MaxFloat32)
		for _, c := range cands {
			s := get(c.Size, a)
			if s < smallest {
				second, smallest = smallest, s
			} else if s > smallest && s < second {
				second = s
			}
		}
		count := 0
		for _, c := range cands {
			if get(c.Size, a)-smallest < eps {
				count++
			}
		}
		target := smallest + remaining/float32(count)
		if second-smallest <= remaining/float32(count) {
			target = second
		}
		kept := cands[:0:0]
		for _, c := range cands {
			s := get(c.Size, a)
			if s-smallest < eps {
				ns := min(target, c.max(a))
				remaining -= ns - s
				set(&c.Size, a, ns)
			}
			if get(c.Size, a) < c.max(a)-eps {
				kept = append(kept, c)
			}
		}
		cands = kept
	}
}

func shrink(children []*UIElement, a axis, excess float32) {
	var cands []*UIElement
	for _, c := range children {
		if c.resizable(a) && get(c.Size, a) > get(c.minSize, a)+eps {
			cands = append(cands, c)
		}
	}
	for excess > eps && len(cands) > 0 {
		largest, second := float32(-math.MaxFloat32), float32(-math.MaxFloat32)
		for _, c := range cands {
			s := get(c.Size, a)
			if s > largest {
				second, largest = largest, s
			} else if s < largest && s > second {
				second = s
			}
		}
		count := 0
		for _, c := range cands {
			if largest-get(c.Size, a) < eps {
				count++
			}
		}
		target := largest - excess/float32(count)
		if largest-second <= excess/float32(count) {
			target = second
		}
		kept := cands[:0:0]
		for _, c := range cands {
			s := get(c.Size, a)
			if largest-s < eps {
				ns := max(target, get(c.minSize, a))
				excess -= s - ns
				set(&c.Size, a, ns)
			}
			if get(c.Size, a) > get(c.minSize, a)+eps {
				kept = append(kept, c)
			}
		}
		cands = kept
	}
}

func growShrink(ui *UIElement, a axis) {
	if ui == nil || len(ui.Children) == 0 {
		return
	}
	inner := get(ui.Size, a) - ui.pad(a)
	if a != ui.mainAxis() {
		for _, c := range ui.Children {
			switch {
			case c.grows(a):
				set(&c.Size, a, clamp(inner, get(c.minSize, a), c.max(a)))
			case c.resizable(a) && get(c.Size, a) > inner:
				set(&c.Size, a, max(inner, get(c.minSize, a)))
			}
		}
		return
	}

	remaining := inner - ui.gap()
	for _, c := range ui.Children {
		remaining -= get(c.Size, a)
	}
	switch {
	case remaining > eps:
		grow(ui.Children, a, remaining)
	case remaining < -eps:
		shrink(ui.Children, a, -remaining)
	}
}
func GrowShrinkX(ui *UIElement) { growShrink(ui, axisX) }
func GrowShrinkY(ui *UIElement) { growShrink(ui, axisY) }

func WrapText(ui *UIElement) {
	if ui == nil || ui.Text == "" || ui.Font == nil {
		return
	}
	width := ui.Size.X - ui.pad(axisX)
	var lines []string
	for para := range strings.SplitSeq(ui.Text, "\n") {
		line := ""
		for w := range strings.FieldsSeq(para) {
			candidate := w
			if line != "" {
				candidate = line + " " + w
			}
			if line != "" && MeasureText(candidate, ui.Font).X > width+eps {
				lines = append(lines, line)
				line = w
			} else {
				line = candidate
			}
		}
		lines = append(lines, line)
	}
	var h float32
	for _, l := range lines {
		h += MeasureText(l, ui.Font).Y + ui.Font.VerticalSpacing
	}
	ui.Lines = lines
	ui.Size.Y = max(ui.BaseSize.Y, h+ui.pad(axisY))
	ui.minSize.Y = ui.Size.Y
}

func alignOffset(f uint32, a axis, free float32) float32 {
	if a == axisX {
		switch {
		case f&UI_ALIGN_CENTER_X != 0:
			return free / 2
		case f&UI_ALIGN_RIGHT != 0:
			return free
		}
		return 0
	}
	switch {
	case f&UI_ALIGN_CENTER_Y != 0:
		return free / 2
	case f&UI_ALIGN_BOTTOM != 0:
		return free
	}
	return 0
}

func crossFlags(a axis) uint32 {
	if a == axisX {
		return UI_ALIGN_LEFT | UI_ALIGN_CENTER_X | UI_ALIGN_RIGHT
	}
	return UI_ALIGN_TOP | UI_ALIGN_CENTER_Y | UI_ALIGN_BOTTOM
}

func Position(ui *UIElement) {
	if ui == nil || len(ui.Children) == 0 {
		return
	}
	main, cross := ui.mainAxis(), ui.mainAxis().other()

	used := ui.gap()
	for _, c := range ui.Children {
		if c.Flag&UI_FLOATING != 0 {
			continue
		}
		used += get(c.Size, main)
	}
	free := get(ui.Size, main) - ui.pad(main) - used
	offset := ui.padStart(main) + alignOffset(ui.Flag, main, free)
	innerCross := get(ui.Size, cross) - ui.pad(cross)

	for _, c := range ui.Children {
		if c.Flag&UI_FLOATING != 0 {
			continue
		}
		f := ui.Flag
		if c.Flag&crossFlags(cross) != 0 {
			f = c.Flag
		}
		set(&c.Position, main, get(ui.Position, main)+offset)
		set(&c.Position, cross, get(ui.Position, cross)+ui.padStart(cross)+
			alignOffset(f, cross, innerCross-get(c.Size, cross)))
		offset += get(c.Size, main) + float32(ui.ChildGap)
	}
}

// nic baker's clays algorithm
func Layout(root *UIElement) {

	if root == nil {
		return
	}
	//flatten the tree
	flatten(root)

	bfs(root, reset)
	dfspo(root, FitSizeX)
	bfs(root, GrowShrinkX)
	bfs(root, WrapText)
	dfspo(root, FitSizeY)
	bfs(root, GrowShrinkY)
	bfs(root, Position)
}
