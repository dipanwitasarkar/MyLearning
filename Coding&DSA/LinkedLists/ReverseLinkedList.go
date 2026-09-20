package linkedlists

// Definition for singly-linked list.
type ListNode struct {
	Val  int
	Next *ListNode
}

func reverseList(head *ListNode) *ListNode {
	if head == nil {
		return head
	}
	/*	curr := head
		prev := head
		next := curr.Next

		for curr.Next != nil {
			curr.Next = next.Next
			next.Next = prev
			prev = next
			next = curr.Next
		}
		return prev
	*/
	var prev *ListNode
	curr := head
	for curr != nil {
		next := curr.Next
		curr.Next = prev
		prev = curr
		curr = next
	}
	return prev
}
