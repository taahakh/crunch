package speed

import "log"

const (
	DIREC_A string = "/Users/taaha/go/"
	DIREC_B string = "/Users/taaha/Documents/GitHub/scraping/go/"
	DIREC_C string = "/Users/taaha/go/src/github.com/taahakh/speed/data/spd/"
)

// this function NEEDS to be called with defer
func Catch_Panic() {
	if err := recover(); err != nil {
		log.Println("Recovering from panic. Error --> ", err)
	}
}
