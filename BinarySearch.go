package speed

import (
	"fmt"
	"strings"
	"unicode"
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

func CleanPostcode(pc string) {
	area := splitPostcode(pc)
	fmt.Println(area)
}

func CleanPostcodes(pc [][]string) *[][]string {

	LENGTH := len(pc)
	// LENGTH := 20
	fmt.Println(LENGTH)
	arr := make([][]string, LENGTH)

	for i := 0; i < LENGTH; i++ {
		arr = append(arr, splitPostcode(pc[i][0]))
	}

	return &arr
}

func SaveCleanPostcodes(filePC, fileSave string) {
	r := ReadPCCSV(filePC)
	c := CleanPostcodes(r)
	saveCleanPostcodes(fileSave, &c)
}

func saveCleanPostcodes(file string, pc **[][]string) {
	WriteCSVpointer(file, pc)
}

func splitPostcode(pc string) []string {
	var a strings.Builder
	var s strings.Builder
	var inbetween int

	a.Grow(4)
	s.Grow(4)

	for i, char := range pc {
		if !unicode.IsSpace(char) {
			a.WriteRune(char)
		} else {
			inbetween = i
			break
		}
	}

	for _, char := range pc[inbetween:] {
		if !unicode.IsSpace(char) {
			s.WriteRune(char)
		}
	}

	return []string{a.String(), s.String()}
}

// formats postcodes without any spaces
func formatPostcode(pc string) (area string) {
	// creating a string builder and allocating size
	var s strings.Builder
	s.Grow(len(pc))

	for _, char := range pc {
		if !unicode.IsSpace(char) {
			s.WriteRune(char)
		}
	}
	return s.String()
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
