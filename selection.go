package main

type Selection struct {
	ssx int          // selection start x
	ssy int          // selection start y
	sex int          // selection end x
	sey int          // selection end y
	isSelected bool  // true if selection is active
}

func (this *Selection) cleanSelection() {
	this.isSelected = false
	this.ssx, this.ssy, this.sex, this.sey = -1, -1, -1, -1
}

func (this *Selection) isUnderSelection(x, y int) bool {
	// Check if there is an active selection
	if this.ssx == -1 || this.ssy == -1  || this.sex == -1 || this.sey == -1 { return false }

	var startx, starty = this.ssx, this.ssy
	var endx, endy = this.sex, this.sey

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

func (this *Selection) getSelectedIndices(content [][]rune) [][]int {
	var selectedIndices = [][]int{}

	// check for empty selection
	if Equal(this.ssx, this.ssy, this.sex, this.sey) {
		return selectedIndices
	}

	// getting selection start point
	var startx, starty = this.ssx, this.ssy
	var endx, endy = this.sex, this.sey

	// swap points if selection is inversed
	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx
		starty, endy = endy, starty
	}

	var inside = false
	// iterate over content, starting from selection start point until out ouf selection
	for j := starty; j < len(content); j++ {
		for i := 0; i < len(content[j]); i++ {
			if this.isUnderSelection(i, j) {
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

func (this *Selection) getSelectionString(content [][]rune) string {
	var ret = []rune {}
	var in = false

	// check for empty selection
	if Equal(this.ssx, this.ssy, this.sex, this.sey) { return "" }

	// getting selection start point
	var startx, starty = this.ssx, this.ssy
	var endx, endy = this.sex, this.sey

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


func (this *Selection) getSelectedLines(content [][]rune)  []int {
	var lineNumbers = make(Set)
	var in = false

	// check for empty selection
	if Equal(this.ssx, this.ssy, this.sex, this.sey) { return lineNumbers.GetKeys() }

	// getting selection start point
	var startx, starty = this.ssx, this.ssy
	var endx, endy = this.sex, this.sey

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
