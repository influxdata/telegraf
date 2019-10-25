package azure_blob

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"testing"
)

func TestGetLoginMethod(t *testing.T) {
	tests := []struct {
		plugin   *AzureBlob
		expected loginMethod
	}{
		{
			plugin: &AzureBlob{
				BlobAccount:    "blobAccount",
				BlobAccountKey: "blobAccountKey",
			},
			expected: accountLogin,
		},
		{
			plugin: &AzureBlob{
				BlobAccountKey: "blobAccountKey",
			},
			expected: invalidLogin,
		},
		{
			plugin: &AzureBlob{
				BlobAccountSasURL: "blobAccountSasUrl",
			},
			expected: sasLogin,
		},
	}

	for _, test := range tests {
		loginmethod := test.plugin.getLoginMethod()
		if loginmethod != test.expected {
			t.Errorf("Wrong login method. Expected %v, actual %v\n", test.expected, loginmethod)
		}
	}
}

func TestCompression(t *testing.T) {
	tests := []struct {
		filename string
		input    string
	}{
		{
			filename: "lala.txt",
			input:    "compress me",
		},
	}

	for _, test := range tests {
		b, err := compressBytesIntoFile(test.filename, []byte(test.input))

		zipReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
		if err != nil {
			t.Error(err)
		}

		for i, zipFile := range zipReader.File {
			if i >= 1 {
				t.Errorf("Zip should only include one file")
			}

			unzippedFileBytes, err := readZipFile(zipFile)
			if err != nil {
				t.Error(err)
				continue
			}

			if zipFile.Name != test.filename {
				t.Errorf("Wrong filename, expected %s, got %s", test.filename, zipFile.Name)
			}

			if string(unzippedFileBytes) != test.input {
				t.Errorf("Expected %s, got %s", test.input, string(unzippedFileBytes))
			}
		}
	}
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
