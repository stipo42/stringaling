package replaceall

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

func TestReplaceAll_Short(t *testing.T) {
	inputString := "Hello billy <kw> SPAM  </kw> this"
	expectedString := "Hello billy CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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

func TestReplaceAll_SameToken(t *testing.T) {
	inputString := "Hello billy <kw> SPAM <kw> this"
	expectedString := "Hello billy CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	strgr.StartToken = "<kw>"
	strgr.EndToken = "<kw>"
	_,err := strgr.Replace()
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

func TestReplaceAll_SameTokenReset(t *testing.T) {
	inputString := "Hello billy <kw> SPAM <kw> <kthis"
	expectedString := "Hello billy CRACKS <kthis"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	strgr.StartToken = "<kw>"
	strgr.EndToken = "<kw>"
	_,err := strgr.Replace()
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

func TestReplaceAll_SameTokenResetAtStart(t *testing.T) {
	inputString := "<kHello billy <kw> SPAM <kw> this"
	expectedString := "<kHello billy CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	strgr.StartToken = "<kw>"
	strgr.EndToken = "<kw>"
	_,err := strgr.Replace()
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

func TestReplaceAll_SameTokenResetAtEnd(t *testing.T) {
	inputString := "Hello billy <kw> SPAM <kw> this<k"
	expectedString := "Hello billy CRACKS this<k"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	strgr.StartToken = "<kw>"
	strgr.EndToken = "<kw>"
	_,err := strgr.Replace()
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

func TestReplaceAll_ResetStart(t *testing.T) {
	inputString := "Hello <kbilly <kw> SPAM  </kw> this"
	expectedString := "Hello <kbilly CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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
func TestReplaceAll_ResetEnd(t *testing.T) {
	inputString := "Hello /kbilly <kw> SPAM  </kw> this"
	expectedString := "Hello /kbilly CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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
func TestReplaceAll_ResetEndAtEnd(t *testing.T) {
	inputString := "Hello billy <kw> SPAM  </kw> this /k"
	expectedString := "Hello billy CRACKS this /k"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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
func TestReplaceAll_ResetStartAtEnd(t *testing.T) {
	inputString := "Hello billy <kw> SPAM  </kw> this <k"
	expectedString := "Hello billy CRACKS this <k"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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
func TestReplaceAll_ResetEndAtStart(t *testing.T) {
	inputString := "/kHello billy <kw> SPAM  </kw> this"
	expectedString := "/kHello billy CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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
func TestReplaceAll_ResetStartAtStart(t *testing.T) {
	inputString := "<kHello billy <kw> SPAM  </kw> this"
	expectedString := "<kHello billy CRACKS this"
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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
func TestReplaceAll_MediumMultiLine(t *testing.T) {
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
	expectedString := `HELLO
THIS
IS
A
LARGE
BLOCK OF TEXT
REPLACE EVERYTHING BETWEEN THE CRACKS

CRACKS`
	sw := bytes.NewBufferString("")

	strgr := createReplacer(inputString, sw)
	_,err := strgr.Replace()
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

func TestReplaceAll_NestedBlocks(t *testing.T) {
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
	expectedString := `CRACKS EVERYTHING BETWEEN THE CRACKS

CRACKS`
	sw := bytes.NewBufferString("")
	strgr := createReplacer(inputString, sw)

	_,err := strgr.Replace()
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

func TestReplaceAll_Split(t *testing.T) {
	inputString := `<kwHELLO 
<kwTHIS
IS
A
LARGE/kw>
BLOCK OF TEXT REPLACE /kw> EVERYTHING BETWEEN THE <kw
AND /kw>

<kw bannaana> banaba
help
</kw>` //128 chars
	expectedString := `CRACKS EVERYTHING BETWEEN THE CRACKS

CRACKS`
	sw := bytes.NewBufferString("")
	sw2 := bytes.NewBufferString("")
	strgr := createReplacer(inputString, sw)
	strgr.GoUntil = 64
	strgr.StartAt = 0
	strgr2 := createReplacer(inputString, sw2)
	strgr2.GoUntil = 64
	strgr2.StartAt = 64

	c1,err := strgr.Replace()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		t.Fail()
	} else {
		var c2 bool
		c2,err = strgr2.Replace()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			t.Fail()
		}else{
			outputString := sw.String() + sw2.String()
			if outputString != expectedString {
				t.Errorf("expected\n'%s'\nbut got\n'%s'", expectedString, outputString)
				t.Fail()
			}
			if !c1 {
				t.Errorf("first replacement wasn't confident and should have been")
				t.Fail()
			}
			if !c2 {
				t.Errorf("second replacement wasn't confident and should have been")
				t.Fail()
			}
		}
	}
}

func createReplacer(inputString string, output io.Writer) AllReplacer {

	strgr := AllReplacer{
		StartToken: "<kw",
		EndToken:   "/kw>",
		Token:      "CRACKS",
		StartAt:    0,
		GoUntil:    int64(len(inputString)),
	}

	strgr.ReaderSpawner = func() (io.Reader, error) {
		return bytes.NewReader([]byte(inputString)), nil
	}

	strgr.WriterSpawner = func() (io.Writer, error) {
		return output, nil
	}

	return strgr
}
