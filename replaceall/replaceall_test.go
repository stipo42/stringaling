package replaceall

import (
	"bytes"
	"github.com/stipo42/stringaling/internal/util"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	util.DEBUG = true
	os.Exit(m.Run())
}

func TestStringalinger_ReplaceAllShort(t *testing.T) {
	expectedString := "Hello billy CRACKS this"
	inputString := "Hello billy <kw> SPAM  </kw> this"
	sw := bytes.NewBufferString("")

	strgr := AllReplacer{
		Reader:     bytes.NewReader([]byte(inputString)),
		Writer:     sw,
		StartToken: "<kw",
		EndToken:   "/kw>",
		Token:      "CRACKS",
	}
	strgr.Replace()

	outputString := sw.String()
	if outputString != expectedString {
		t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
		t.Fail()
	}
}

func TestStringalinger_ReplaceAllSameToken(t *testing.T) {
	expectedString := "Hello billy CRACKS this"
	inputString := "Hello billy <kw> SPAM <kw> this"
	sw := bytes.NewBufferString("")

	strgr := AllReplacer{
		Reader:     bytes.NewReader([]byte(inputString)),
		Writer:     sw,
		StartToken: "<kw>",
		EndToken:   "<kw>",
		Token:      "CRACKS",
	}
	strgr.Replace()

	outputString := sw.String()
	if outputString != expectedString {
		t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
		t.Fail()
	}
}

func TestStringalinger_ReplaceAllResetStart(t *testing.T) {
	expectedString := "Hello <kbilly CRACKS this"
	inputString := "Hello <kbilly <kw> SPAM  </kw> this"
	sw := bytes.NewBufferString("")

	strgr := AllReplacer{
		Reader:     bytes.NewReader([]byte(inputString)),
		Writer:     sw,
		StartToken: "<kw",
		EndToken:   "/kw>",
		Token:      "CRACKS",
	}
	strgr.Replace()

	outputString := sw.String()
	if outputString != expectedString {
		t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
		t.Fail()
	}
}
func TestStringalinger_ReplaceAllResetEnd(t *testing.T) {
	expectedString := "Hello billy CRACKS this /k"
	inputString := "Hello billy <kw> SPAM  </kw> this /k"
	sw := bytes.NewBufferString("")

	strgr := AllReplacer{
		Reader:     bytes.NewReader([]byte(inputString)),
		Writer:     sw,
		StartToken: "<kw",
		EndToken:   "/kw>",
		Token:      "CRACKS",
	}
	strgr.Replace()

	outputString := sw.String()
	if outputString != expectedString {
		t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
		t.Fail()
	}
}
func TestStringalinger_ReplaceAllMediumMultiLine(t *testing.T) {
	expectedString := `HELLO
THIS
IS
A
LARGE
BLOCK OF TEXT
REPLACE EVERYTHING BETWEEN THE CRACKS

CRACKS`
	inputString := `HELLO
THIS
IS
A
LARGE
BLOCK OF TEXT
REPLACE EVERYTHING BETWEEN THE <kw
AND /kw>

<kw bannaana> banaba
help
</kw>`
	sw := bytes.NewBufferString("")

	strgr := AllReplacer{
		Reader:     bytes.NewReader([]byte(inputString)),
		Writer:     sw,
		StartToken: "<kw",
		EndToken:   "/kw>",
		Token:      "CRACKS",
	}
	strgr.Replace()

	outputString := sw.String()
	if outputString != expectedString {
		t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
		t.Fail()
	}
}

func TestStringalinger_ReplaceAllNestedBlocks(t *testing.T) {
	expectedString := `CRACKS EVERYTHING BETWEEN THE CRACKS

CRACKS`
	// first <kw is split on the size field
	inputString := `<kwHELLO 
<kwTHIS
IS
A
LARGE/kw>
BLOCK OF TEXT REPLACE /kw> EVERYTHING BETWEEN THE <kw
AND /kw>

<kw bannaana> banaba
help
</kw>`
	sw := bytes.NewBufferString("")

	strgr := AllReplacer{
		Reader:     bytes.NewReader([]byte(inputString)),
		Writer:     sw,
		StartToken: "<kw",
		EndToken:   "/kw>",
		Token:      "CRACKS",
	}

	strgr.Replace()

	outputString := sw.String()
	if outputString != expectedString {
		t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
		t.Fail()
	}
}
