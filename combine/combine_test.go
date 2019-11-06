package combine

import (
	"bytes"
	"github.com/stipo42/stringaling/internal/util"
	"io"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	util.DEBUG = true
	os.Exit(m.Run())
}

func TestStringalinger_CombineShort(t *testing.T) {
	var streams []io.Reader
	streams = append(streams, bytes.NewReader([]byte("ONE")))
	streams = append(streams, bytes.NewReader([]byte("TWO")))
	streams = append(streams, bytes.NewReader([]byte("THREE")))
	streams = append(streams, bytes.NewReader([]byte("FOUR")))
	sw := bytes.NewBufferString("")
	expectedString := "ONETWOTHREEFOUR"

	strcmb := StreamCombiner{
		Streams: streams,
		Output:  sw,
		Buffer:  2,
	}

	err := strcmb.Combine()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		t.Fail()
	} else {
		outputString := sw.String()
		if outputString != expectedString {
			t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
			t.Fail()
		}
	}

}
