package render

import "strings"

type MyStringBuilder struct {
	strings.Builder
}

func (b *MyStringBuilder) WriteStrings(text ...string) {
	for _, t := range text {
		b.WriteString(t)
	}
}
