package easymedium

func maxProfit(prices []int) int {
	if len(prices) == 0 {
		return 0
	}
	minprice := prices[0]
	maxProfit := 0
	for i := 1; i < len(prices); i++ {
		if prices[i] < minprice {
			minprice = prices[i]
		}
		profit := prices[i] - minprice
		if profit > maxProfit {
			maxProfit = profit
		}
	}
	return maxProfit
}
