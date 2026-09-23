package mediumhard

// classic sliding window
func lengthOfLongestSubstring(s string) int {
	seen := make(map[byte]int)
	left := 0
	result := 0
	for right := 0; right < len(s); right++ {
		ch := s[right]
		if idx, found := seen[ch]; found {
			left = max(left, idx+1)
		}
		if right-left+1 > result {
			result = right - left + 1
		}
		seen[ch] = right
	}
	return result
}
