package render

import "strings"

// renderパッケージ内で共通利用するビルダ構造体
type MyStringBuilder struct {
	strings.Builder
}

func (b *MyStringBuilder) WriteStrings(text ...string) {
	for _, t := range text {
		b.WriteString(t)
	}
}
