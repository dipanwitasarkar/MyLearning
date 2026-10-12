package Matrix

func spiralOrder(matrix [][]int) []int {
	// Maintain 4 boundaries top, bottom, left, right. traverse all elements all directoins exactly once
	// o(mn) and shrink the boundaries.
	// 1. edge cases
	if len(matrix) == 0 {
		return []int{}
	}
	top, bottom := 0, len(matrix)-1
	left, right := 0, len(matrix[0])-1
	result := make([]int, 0, len(matrix)*len(matrix[0]))
	for top <= bottom && left <= right {
		// left -> right
		for col := left; col <= right; col++ {
			result = append(result, matrix[top][col])
		}
		top++
		// top -> bottom
		for row := top; row <= bottom; row++ {
			result = append(result, matrix[row][right])
		}
		right--
		// right -> left
		if top <= bottom {
			for col := right; col >= left; col-- {
				result = append(result, matrix[bottom][col])
			}
		}
		bottom--
		// bottom -> top
		if left <= right {
			for row := bottom; row >= top; row-- {
				result = append(result, matrix[row][left])
			}
		}
		left++
	}
	return result
}
