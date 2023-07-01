package selection

import . "edgo/internal/utils"
type Selection struct {
	Ssx        int  // selection Start x
	Ssy        int  // selection Start y
	Sex        int  // selection end x
	Sey        int  // selection end y
	IsSelected bool // true if selection is active
}

func (this *Selection) CleanSelection() {
	this.IsSelected = false
	this.Ssx, this.Ssy, this.Sex, this.Sey = -1, -1, -1, -1
}


func (this *Selection) IsSelectionNonEmpty() bool {
	if this.Ssx == -1 || this.Ssy == -1  || this.Sex == -1 || this.Sey == -1 { return false }
	if Equal(this.Ssx, this.Ssy, this.Sex, this.Sey) {
		return false
	}

	return true
}

func (this *Selection) IsUnderSelection(x, y int) bool {
	// Check if there is an active selection
	if this.Ssx == -1 || this.Ssy == -1  || this.Sex == -1 || this.Sey == -1 { return false }

	var startx, starty = this.Ssx, this.Ssy
	var endx, endy = this.Sex, this.Sey

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx
		starty, endy = endy, starty
	}

	return GreaterEqual(x, y, startx, starty) && LessThan(x, y, endx, endy)
}

func GreaterThan(x, y, x1, y1 int) bool {
	if y > y1 {
		return true
	}
	return y == y1 && x > x1
}

func LessThan(x, y, x1, y1 int) bool {
	if y < y1 {
		return true
	}
	return y == y1 && x < x1
}

func GreaterEqual(x, y, x1, y1 int) bool {
	if y > y1 {
		return true
	}
	if y == y1 && x >= x1 {
		return true
	}
	return false
}

func Equal(x, y, x1, y1 int) bool {
	return x == x1 && y == y1
}

func (this *Selection) GetSelectedIndices(content [][]rune) [][]int {
	var selectedIndices = [][]int{}

	// check for empty selection
	if Equal(this.Ssx, this.Ssy, this.Sex, this.Sey) {
		return selectedIndices
	}

	// getting selection Start point
	var startx, starty = this.Ssx, this.Ssy
	var endx, endy = this.Sex, this.Sey

	// swap points if selection is inversed
	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx
		starty, endy = endy, starty
	}

	var inside = false
	// iterate over Content, starting from selection Start point until out ouf selection
	for j := starty; j < len(content); j++ {
		for i := 0; i < len(content[j]); i++ {
			if this.IsUnderSelection(i, j) {
				selectedIndices = append(selectedIndices, []int{i, j})
				inside = true
			} else  {
				if inside == true { // first time when out ouf selection
					return selectedIndices
				}
			}
		}
	}
	return selectedIndices
}

func (this *Selection) GetSelectionString(content [][]rune) string {
	var ret = []rune {}
	var in = false

	// check for empty selection
	if Equal(this.Ssx, this.Ssy, this.Sex, this.Sey) { return "" }

	// getting selection Start point
	var startx, starty = this.Ssx, this.Ssy
	var endx, endy = this.Sex, this.Sey

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx // swap  points if selection inverse
		starty, endy = endy, starty
	}

	for j := starty; j < len(content); j++ {
		row := content[j]
		for i, char := range row {
			// if inside selection
			if GreaterEqual(i, j, startx, starty) && LessThan(i, j, endx, endy) {
				ret = append(ret, char)
				in = true
			} else {
				in = false
				// only one selection area can be, early return
				if len(ret) > 0 {
					// remove the last newline if present
					if len(ret) > 0 && ret[len(ret)-1] == '\n' { ret = ret[:len(ret)-1] }
					return string(ret)
				}
			}
		}
		if in && LessThan(0, j, endx, endy) {
			ret = append(ret, '\n')
		}
	}

	if len(ret) > 0 && ret[len(ret)-1] == '\n' { ret = ret[:len(ret)-1] }
	return string(ret)
}


func (this *Selection) GetSelectedLines(content [][]rune)  []int {
	var lineNumbers = make(Set)
	var in = false

	// check for empty selection
	if Equal(this.Ssx, this.Ssy, this.Sex, this.Sey) { return lineNumbers.GetKeys() }

	// getting selection Start point
	var startx, starty = this.Ssx, this.Ssy
	var endx, endy = this.Sex, this.Sey

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx // swap  points if selection inverse
		starty, endy = endy, starty
	}

	for j := starty; j < len(content); j++ {
		row := content[j]
		for i, _ := range row {
			// if inside selection
			if GreaterEqual(i, j, startx, starty) && LessThan(i, j, endx, endy) {
				lineNumbers.Add(j)
				in = true
			} else {
				in = false
				// only one selection area can be, early return
				if len(lineNumbers) > 0 {
					return lineNumbers.GetKeys()
				}
			}
		}
		if in && LessThan(0, j, endx, endy) {
			lineNumbers.Add(j)
		}
	}
	return lineNumbers.GetKeys()
}
