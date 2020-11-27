package templating

import "testing"

func BenchmarkTemplateLess(b *testing.B) {
	a := templateSpec{
		template:  "aa|bb|cc|dd|ee|ff",
		separator: "|",
	}
	specs := templateSpecs{a, a}
	for i := 0; i < b.N; i++ {
		specs.Less(0, 1)
	}
}
