package Matrix

func numIslands(grid [][]byte) int {
	if len(grid) == 0 {
		return 0
	}

	rows := len(grid)
	cols := len(grid[0])

	var dfs func(int, int)

	dfs = func(r, c int) {
		if r < 0 || c < 0 || r >= rows || c >= cols ||
			grid[r][c] == '0' {
			return
		}

		grid[r][c] = '0'

		dfs(r+1, c)
		dfs(r-1, c)
		dfs(r, c+1)
		dfs(r, c-1)
	}

	islands := 0

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if grid[r][c] == '1' {
				islands++
				dfs(r, c)
			}
		}
	}

	return islands
}
