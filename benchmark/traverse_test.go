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
		// doc.FindStrictly("[id='HLCXComparisonWidget_feature_div', class='celwidget', data-feature-name='HLCXComparisonWidget', data-csa-c-type='widget', data-csa-c-slot-id='.12', data-csa-c-component='HLCXComparisonWidget', data-csa-c-cs-type='DETAIL_PAGE_DYNAMIC', data-csa-c-id='j1i5cn-olyvyi-gt0c00-y3wchp', data-cel-widget='HLCXComparisonWidget_feature_div']")
		doc.FindStrictlyOnce("[id='HLCXComparisonWidget_feature_div', class='celwidget', data-feature-name='HLCXComparisonWidget', data-csa-c-type='widget', data-csa-c-slot-id='.12', data-csa-c-component='HLCXComparisonWidget', data-csa-c-cs-type='DETAIL_PAGE_DYNAMIC', data-csa-c-id='j1i5cn-olyvyi-gt0c00-y3wchp', data-cel-widget='HLCXComparisonWidget_feature_div']")
		// doc.FindStrictly("[id='HLCXComparisonWidget_feature_div', class='celwidget', data-feature-name='HLCXComparisonWidget', data-csa-c-type='widget', data-csa-c-slot-id='.12', data-csa-c-component='HLCXComparisonWidget', data-csa-c-cs-type='DETAIL_PAGE_DYNAMIC', data-csa-c-id='j1i5cn-olyvyi-gt0c00-y3wchp', data-cel-widget='HLCXComparisonWidget_feature_div', alpha='']")
		// doc.Find("[id='HLCXComparisonWidget_feature_div', class='celwidget']")
		// doc.Find("[id='HLCXComparisonWidget_feature_div']")
		// doc.FindOnce("[id='HLCXComparisonWidget_feature_div']")
		// doc.Query("[id='HLCXComparisonWidget_feature_div']")
		// doc.QueryOnce("[id='HLCXComparisonWidget_feature_div']")
	}
}

func hashCompare(x map[string]string) {

	sli := []string{"yes", "no"}
	for _, item := range sli {
		if _, ok := x[item]; ok {
		}
	}

}

func sliceComapre(x []string) {
	sli := []string{"yes", "no"}
	for _, item := range sli {
		for _, newlist := range x {
			if newlist == item {

			}
		}
	}
}

func BenchmarkHash(b *testing.B) {
	hash := make(map[string]string, 2)
	hash["yes"] = ""
	hash["no"] = ""
	for n := 0; n < b.N; n++ {
		hashCompare(hash)
	}
}

func BenchmarkSlice(b *testing.B) {
	sli := []string{"yes", "no"}
	for n := 0; n < b.N; n++ {
		sliceComapre(sli)
	}
}
