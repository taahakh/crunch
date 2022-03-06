package speed

import "log"

// import (
// 	"encoding/csv"
// 	"log"
// 	"os"
// )

const (
	DIREC_A string = "/Users/taaha/go/"
	DIREC_B string = "/Users/taaha/Documents/GitHub/scraping/go/"
)

// type Ip struct {
// 	ip, header string
// }

// this function NEEDS to be called with defer
func Catch_Panic() {
	if err := recover(); err != nil {
		log.Println("Recovering from panic. Error --> ", err)
	}
}
