package benchmark

import (
	"bytes"
	"os"
	"testing"

	"github.com/taahakh/speed/traverse"
)

func BenchmarkParse(b *testing.B) {
	loc := "C:\\Users\\taaha\\go\\src\\github.com\\taahakh\\speed\\data\\spd\\amazonscrape.html"
	file, err := os.ReadFile(loc)
	if err != nil {
		return
	}

	for n := 0; n < b.N; n++ {
		traverse.HTMLDoc(bytes.NewReader(file))
	}
}

func BenchmarkScrape(b *testing.B) {
	loc := "C:\\Users\\taaha\\go\\src\\github.com\\taahakh\\speed\\data\\spd\\amazonscrape.html"
	file, err := os.ReadFile(loc)
	if err != nil {
		return
	}
	doc, _ := traverse.HTMLDoc(bytes.NewReader(file))

	for n := 0; n < b.N; n++ {
		doc.FindStrictlyOnce("a")
	}
}
