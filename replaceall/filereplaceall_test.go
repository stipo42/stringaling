package replaceall

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestReplaceAll(t *testing.T) {
	inputFileName := "testdata/TestReplaceAll-input.xml"
	expectedFileName := "testdata/TestReplaceAll-expected.xml"
	outputFileName := "testdata/results/results-clean.xml"
	startToken := "<phi>"
	endToken := "</phi>"
	token := "<redacted></redacted>"
	threads := 5

	err := ReplaceAll(inputFileName, outputFileName, startToken, endToken, token, threads)
	if err != nil {
		t.Errorf("error during execution: %s", err)
		t.Fail()
	} else {
		var expected string
		expected, err = quickRead(expectedFileName)
		if err != nil {
			t.Errorf("could not read expected file (%s): %s", expectedFileName, err)
			t.Fail()
		} else {
			var actual string
			actual, err = quickRead(outputFileName)
			if err != nil {
				t.Errorf("could not read output file (%s): %s", outputFileName, err)
				t.Fail()
			} else if actual != expected {
				t.Errorf("actual did not equal expected: %s != %s", actual, expected)
				t.Fail()
			}
		}
	}
}

func quickRead(fileName string) (content string, err error) {
	var f *os.File
	f, err = os.Open(fileName)
	if err == nil {
		var out []byte
		out, err = ioutil.ReadAll(f)
		content = string(out)
	}
	return
}
