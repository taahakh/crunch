package speed

import (
	"fmt"
	"strings"
	"unicode"
)

// this postcode works with no space inbetween or anywhere
type Postcode struct {
	Postcode                     string // postcode with space
	Status                       bool   // 0 live 1 terminated
	Usertype                     bool   // 0 small 1 large
	Easting                      int
	Northing                     int
	Positional_quality_indicator int
	Country                      string
	Latitude                     float64
	Longitude                    float64
	Postcode_area                string
	Postcode_district            string
	Postcode_sector              string
	Outcode                      string
	Incode                       string
}

// func (p *Postcode) setPostcode(pc string, pool *map[string]Postcode) {

// }

func CreatePostcodeName(pc string) Postcode {
	p := Postcode{}
	p.Postcode = pc
	return p
}

func CreatePostcodeDefault() Postcode {
	p := Postcode{}
	p.Postcode = "xxx xxx"
	p.Status = false
	p.Usertype = false
	p.Easting = 0
	p.Northing = 0
	p.Positional_quality_indicator = 0
	p.Country = "xxxxxxxxxx"
	p.Latitude = 0.0
	p.Longitude = 0.0
	p.Postcode_area = "xxx"
	p.Postcode_district = "xxx"
	p.Postcode_sector = "xxx"
	p.Outcode = "xxx"
	p.Incode = "xxx"
	return p
}

func CreateMapPostcode() map[string]Postcode {
	return make(map[string]Postcode)
}

func CleanPostcode(pc string) {
	area := splitPostcode(pc)
	fmt.Println(area)
}

func CleanPostcodes(pc [][]string) *[][]string {

	// has to be dynamically created and added onto the array
	// want to return the address of the array - we do not want a copy
	// making size 0 so it just started to append straight away
	arr := make([][]string, 0)

	for i := 0; i < len(pc); i++ {
		arr = append(arr, splitPostcode(pc[i][0]))
	}

	return &arr
}

func SaveCleanPostcodes(filePC, fileSave string) {
	r := ReadPCCSV(filePC)
	c := CleanPostcodes(r)
	WriteCSVpointer(fileSave, c)
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
