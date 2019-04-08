package kinesis_consumer

import "testing"

func TestDecompression(t *testing.T) {
	// All compression should result in the same output
	expectedOutput := []byte(`testdata1234!"£$testdata1234!"£$testdata1234!"£$`)

	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "snappy",
			input: []byte{51, 64, 116, 101, 115, 116, 100, 97, 116, 97, 49, 50, 51, 52, 33, 34, 194, 163, 36, 134, 17, 0},
		},
		{
			name:  "gzip",
			input: []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 42, 73, 45, 46, 73, 73, 44, 73, 52, 52, 50, 54, 81, 84, 58, 180, 88, 133, 176, 0, 32, 0, 0, 255, 255, 54, 208, 10, 134, 51, 0, 0, 0},
		},
	}

	for _, test := range tests {
		k := KinesisConsumer{
			CompressedMetrics: true,
			CompressionType:   test.name,
		}

		output, err := k.decompress(test.input)
		if err != nil {
			// Snappy can't actually fail but we put it in here to follow convention.
			t.Logf("Decompression of %s data failed. Error: %s", test.name, err)
			t.Fail()
		}

		if string(output) != string(expectedOutput) {
			t.Logf("%s produced a different result than expected.", test.name)
			t.Fail()
		}
	}
}
