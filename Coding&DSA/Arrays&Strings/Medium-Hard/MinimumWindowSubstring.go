package mediumhard

func minWindow(s string, t string) string {
	need := make(map[byte]int) // target string freq
	// getting target string freq
	for i := 0; i < len(t); i++ {
		need[t[i]]++
	}
	needCount := len(need)
	if needCount == 0 {
		return ""
	}
	// how much got by window
	have := 0
	window := make(map[byte]int)
	minlen := len(s) + 1
	left := 0
	start := 0 // to store the min starting index of the window
	for right := 0; right < len(s); right++ {
		window[s[right]]++
		if need[s[right]] > 0 && need[s[right]] == window[s[right]] {
			have++
		}
		// for a valid window, try reducing it until it is valid window to get minimum
		for have == needCount {
			if right-left+1 < minlen {
				minlen = right - left + 1
				start = left
			}
			window[s[left]]--
			if need[s[left]] > 0 && window[s[left]] < need[s[left]] {
				have--
			}
			left++
		}
	}
	// if no valid result found
	if minlen == len(s)+1 {
		return ""
	}
	return s[start : start+minlen]
}
