package easymedium

func twoSum(nums []int, target int) []int {
	// for i:=0;i<len(nums);i++{
	//     for j:=1;j<len(nums);j++{
	//         if i != j && nums[i] + nums[j] == target{
	//             return []int{i,j}
	//         }
	//     }
	// }
	// return []int{}
	var seenMap = make(map[int]int)
	for i, num := range nums {
		complement := target - num
		if idx, found := seenMap[complement]; found {
			return []int{idx, i}
		}
		seenMap[num] = i
	}
	return nil
}
