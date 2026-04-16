// You can edit this code!
// Click here and start typing.
package main

import "fmt"

var N = 3
var tree []int = make([]int, 4*N)

func build(ar []int, index int, start int, end int) {
	if start == end {
		tree[index] = ar[start]
		return
	}

	mid := start + (end-start)/2

	build(ar, 2*index+1, start, mid)
	build(ar, 2*index+2, mid+1, end)
	tree[index] = tree[2*index+1] + tree[2*index+2]
}

func query(index int, start int, end int, qs int, qe int) int {

	if qe < start || qs > end {
		return 0
	}

	if qs <= start && qe >= end {
		return tree[index]
	}

	mid := start + (end-start)/2

	return query(2*index+1, start, mid, qs, qe) + query(2*index+2, mid+1, end, qs, qe)

}

func pointUpdate(index int, start int, end int, pos int, val int) {

	if start == end {
		tree[start] = val
		return
	}

	mid := start + (end-start)/2

	if pos < mid {
		pointUpdate(2*index+1, start, mid, pos, val)
	} else {
		pointUpdate(2*index+2, mid+1, end, pos, val)
	}
	tree[index] = tree[2*index+1] + tree[2*index+2]
}

func main() {
	ar := make([]int, 3)

	ar[0] = 3
	ar[1] = 1
	ar[2] = 5

	build(ar, 0, 0, len(ar)-1)
	fmt.Println(tree)
	fmt.Println(query(0, 0, len(ar)-1, 1, 1))

}
