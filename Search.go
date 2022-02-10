package speed

import (
	"fmt"
)

func BinarySearch(record [][]string, item string) (exists bool, index int) {

	item = formatPostcode(item)
	var middle int
	LENGTH := len(record)

	fmt.Println(LENGTH)
	if LENGTH%2 == 0 {
		fmt.Println("It is of odd length")
	}

	middle = LENGTH / 2
	end := 0
	area := ""

	for {
		area = formatPostcode(record[middle][0])
		fmt.Println("Middle: ", middle)
		fmt.Println("Item: ", area)

		if middle == 0 || end == 20 || middle == LENGTH {
			fmt.Println("Reached zero/5")
			break
		}
		if area == item {
			return true, middle
		} else if area < item {
			fmt.Print(" Splitting left")
			end++
			middle = middle / 2
		} else {
			fmt.Print(" Splitting right")
			end++
			middle = (LENGTH + middle) / 2
		}

	}

	return false, -1
}

func LinearSearch(record [][]string, element string) (index int) {
	for i := 0; i < len(record); i++ {
		if record[i][0] == element {
			return i
		}
	}
	return -1
}

// func findMiddle(x, y int) (middle int) {
// 	return (x + y) / 2
// }
