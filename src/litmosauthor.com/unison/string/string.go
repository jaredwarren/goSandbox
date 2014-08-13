package string

func Reverse(s string) string {
	b := []bytes(s)
	for i := 0; i < len(b)/2; i++ {
		j := len(b)-i-1
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}