package linkedlists

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func detectCycle(head *ListNode) *ListNode {
	fastNode := head
	slowNode := head
	for fastNode != nil && fastNode.Next != nil {
		slowNode = slowNode.Next
		fastNode = fastNode.Next.Next
		if fastNode == slowNode {
			break
		}
	}
	// No cycle
	if fastNode == nil || fastNode.Next == nil {
		return nil
	}

	// *Find cycle start*
	p1 := head
	p2 := slowNode
	for p1 != p2 {
		p1 = p1.Next
		p2 = p2.Next
	}
	return p1
}
