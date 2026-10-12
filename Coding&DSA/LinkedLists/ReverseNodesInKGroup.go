package linkedlists

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func reverseKGroup(head *ListNode, k int) *ListNode {
	dummy := &ListNode{Next: head}
	groupPrev := dummy
	for {
		// check if k nodes exist
		kth := getKthNode(groupPrev, k)
		if kth == nil {
			break
		}
		groupNext := kth.Next

		prev := groupNext
		curr := groupPrev.Next
		for curr != groupNext {
			temp := curr.Next
			curr.Next = prev
			prev = curr
			curr = temp
		}
		temp := groupPrev.Next
		groupPrev.Next = kth
		groupPrev = temp
	}
	return dummy.Next
}

func getKthNode(curr *ListNode, k int) *ListNode {
	for curr != nil && k > 0 {
		curr = curr.Next
		k--
	}
	return curr
}
