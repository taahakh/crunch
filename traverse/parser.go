package traverse

import (
	"log"
	"regexp"
	"strings"
)

// Customisable search struct
type Search struct {
	Tag      string
	Selector []Attribute
	Attr     []Attribute
}

type SearchBuilder struct {
	Bracket, FinishBracket, ValueState, EqualState, KeyPair bool
	Left, Right, SelectorState                              bool
	Tracking                                                []string
	Attr                                                    []Attribute
	Selector                                                []Attribute
	Tag                                                     string
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func CreateSearch(sb *SearchBuilder) *Search {

	return &Search{
		Tag:      sb.Tag,
		Selector: sb.Selector,
		Attr:     sb.Attr,
	}
}

func FinderParser(s string) *Search {
	lol := []string{}
	for _, v := range s {
		lol = append(lol, string(v))
	}
	build := &SearchBuilder{}
	build.Bracket = false
	build.FinishBracket = false
	build.SelectorState = false
	build.ValueState = false
	build.EqualState = false
	build.Left = false
	build.Right = false
	build.KeyPair = false
	if len(s) == 0 {
		log.Println("string empty")
	}
	finderParser(lol, 0, build)
	checkBuild(build)
	if len(build.Tracking) > 0 {
		build.Tag = Reverse(strings.Join(build.Tracking, ""))
		build.Tracking = nil
	}
	// fmt.Println(build)
	return CreateSearch(build)
}

func finderParser(r []string, i int, b *SearchBuilder) {
	i++
	if i != len(r) {
		finderParser(r, i, b)
	}
	checkStringParse(r[i-1], b, i-1)
}

func checkBuild(b *SearchBuilder) {
	if b.Bracket && !b.FinishBracket {
		log.Fatal("Brackets are not closed")
	}
}

func checkStringParse(r string, b *SearchBuilder, i int) {

	var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

	var appendstring = func(s []string) string {
		join := strings.Join(s, "")
		return join
	}

	var attrAppend = func() {
		if b.KeyPair {
			b.KeyPair = false
			item := &b.Attr[len(b.Attr)-1]
			item.Name = Reverse(appendstring(b.Tracking))
		} else {
			b.Attr = append(b.Attr, Attribute{Name: Reverse(appendstring(b.Tracking))})
		}
		b.Tracking = nil
	}

	var selectorAppend = func() {
		if r == "." {
			r = "class"
		} else {
			r = "id"
		}
		b.Selector = append(b.Selector, Attribute{Name: r, Value: Reverse(appendstring(b.Tracking))})
		b.Tracking = nil
	}

	var valueAppend = func() {
		b.Attr = append(b.Attr, Attribute{Value: Reverse(appendstring(b.Tracking))})
		b.Tracking = nil
	}

	var left = func() {
		if r == "." || r == "#" {
			if b.SelectorState {
				log.Fatal("selector state")
			}
			selectorAppend()
			b.SelectorState = true
			return
		}

		if isAlphaNumeric(r) || r == "-" {
			b.SelectorState = false
			b.Tracking = append(b.Tracking, r)
			return
		} else {
			log.Fatal("Incorrect ", r, " on LEFT")
		}

	}

	var right = func() {
		if r == "]" {
			log.Fatal("Another ] bracket")
		}

		if r == "[" {
			b.FinishBracket = true
			attrAppend()
			return
		}

		if r == "," {
			attrAppend()
			return
		}

		// value state symbol
		if r == "'" {
			if !b.ValueState {
				b.ValueState = true
			} else {
				b.ValueState = false
				b.EqualState = true
			}
			return
		} else if b.EqualState {
			if r != "=" {
				log.Fatal("Closed value state that has no equal")
			} else {
				// attrAppend()
				valueAppend()
				b.EqualState = false
				b.KeyPair = true
				return
			}
		} else if b.ValueState {
			b.Tracking = append(b.Tracking, r)
			return
		}

		if isAlphaNumeric(r) || r == "-" {
			b.Tracking = append(b.Tracking, r)
			return
		} else {
			if r == " " {
				return
			}
			log.Fatal("final alphanumeric")
		}

	}

	if b.Left {
		left()
		return
	} else if b.Right {

		if b.FinishBracket {
			b.Left = true
			b.Right = false
			left()
			return
		}
		right()
		return
	}

	if r == "]" {
		b.Bracket = true
		b.Right = true
	} else if isAlphaNumeric(r) {
		b.Left = true
		left()
	}
}
