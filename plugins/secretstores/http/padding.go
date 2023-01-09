package http

func PKCS7Trimming(in []byte) []byte {
	// 'count' number of bytes where padded to the end of the clear-text
	// each containing the value of 'count'
	count := int(in[len(in)-1])
	return in[:len(in)-count]
}
