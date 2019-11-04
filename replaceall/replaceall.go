package replaceall

import (
	"github.com/stipo42/stringaling/internal/util"
	"io"
	"time"
)

type AllReplacer struct {
	Reader     io.Reader
	Writer     io.Writer
	StartToken string
	EndToken   string
	Token      string
}

func (s AllReplacer) Replace() (err error) {
	start := time.Now().UnixNano()
	if s.StartToken == s.EndToken {
		err = s.replaceSameStartEnd()
	} else {
		err = s.replace()
	}

	diff := time.Now().UnixNano() - start
	util.Info("Replace took %s to execute", util.HumanReadable(diff))

	return
}

func (s AllReplacer) replace() (err error) {
	noWriteDepth := 0 // When not zero, don't write byte to output
	chunk := make([]byte, 1)
	sct := 0 // Increase every time the consecutively read byte matches that index of the start token, when longer than start token length, depth increases
	ect := 0 // Increase every time the consecutively read byte matches that index of the end token,   when longer than end   token length, depth decreases
	skipped := 0

	var rerr error
	for ; rerr == nil; {
		_, rerr = s.Reader.Read(chunk)
		if rerr != nil {
			if rerr == io.EOF {
				util.Debug("End of file: %s", rerr)
			} else {
				util.Error("Couldn't read chunk: %s", rerr)
				err = rerr
			}
			// Final chunker check
			if sct > 0 && noWriteDepth == 0 && ect == 0 {

				// Backfill for this miss.
				tt := s.StartToken[0:sct]
				util.Debug("final start miss!: %s", tt)
				s.writeS(tt)
			}
			if ect > 0 && noWriteDepth == 0 && sct == 0 {
				// Backfill for this miss.
				tt := s.EndToken[0:ect]
				util.Debug("final end miss!: %s", tt)
				s.writeS(tt)
			}
		} else {
			if noWriteDepth > 0 {
				skipped += 1
			}
			if chunk[0] == s.StartToken[sct] {
				sct += 1
			} else if sct > 0 {
				if noWriteDepth == 0 {
					if ect == 0 {
						// Backfill for this miss if no ect.
						tt := s.StartToken[0:sct]
						util.Debug("start miss!: %s", tt)
						s.writeS(tt)
					}
				}
				sct = 0
			}
			if chunk[0] == s.EndToken[ect] {
				ect += 1
			} else if ect > 0 {
				if noWriteDepth == 0 {
					if sct == 0 {
						// Backfill for this miss if no sct.
						tt := s.EndToken[0:ect]
						util.Debug("end miss!: %s", tt)
						s.writeS(tt)
					}
				}
				ect = 0
			}

			if sct >= len(s.StartToken) {
				sct = 0
				noWriteDepth += 1
			}
			if ect >= len(s.EndToken) {
				ect = 0
				noWriteDepth -= 1
				if noWriteDepth < 0 {
					noWriteDepth = 0
				} else if noWriteDepth == 0 {
					skipped += len(s.StartToken) + len(s.EndToken)
					util.Debug("replaced %d bytes", skipped)
					s.writeS(s.Token)
				}
			} else if noWriteDepth == 0 && sct == 0 && ect == 0 {
				s.write(chunk)
			}
		}
	}
	return
}
func (s AllReplacer) replaceSameStartEnd() (err error) {
	noWriteDepth := 0 // When not zero, don't write byte to output
	chunk := make([]byte, 1)
	ct := 0 // Increase every time the consecutively read byte matches that index of the start token, when longer than start token length, depth increases
	skipped := 0
	var rerr error
	for ; rerr == nil; {
		_, rerr = s.Reader.Read(chunk)
		if rerr != nil {
			if rerr == io.EOF {
				util.Debug("End of file: %s", rerr)
			} else {
				util.Error("Couldn't read chunk: %s", rerr)
				err = rerr
			}
			// Final chunker check
			if ct > 0 && noWriteDepth == 0 {
				// Backfill for this miss.
				s.write([]byte(s.StartToken[0:ct]))
			}
		} else {
			if noWriteDepth > 0 {
				skipped += 1
			}
			if chunk[0] == s.StartToken[ct] {
				ct += 1
			} else if ct > 0 {
				if noWriteDepth == 0 {
					// Backfill for this miss.
					s.write([]byte(s.StartToken[0:ct]))
				}
				ct = 0
			}
			if ct >= len(s.StartToken) {
				ct = 0
				if noWriteDepth == 0 {
					noWriteDepth = 1
				} else if noWriteDepth == 1 {
					skipped += len(s.StartToken) + len(s.EndToken)
					util.Debug("replaced %d bytes", skipped)
					s.writeS(s.Token)
					noWriteDepth = 0
				}
			} else if noWriteDepth == 0 && ct == 0 {
				s.write(chunk)
			}
		}
	}
	return
}

func (s AllReplacer) writeS(str string) (wroteBytes int) {
	return s.write([]byte(str))
}
func (s AllReplacer) write(ibytes []byte) (wroteBytes int) {
	if len(ibytes) > 0 {
		var werr error
		wroteBytes, werr = s.Writer.Write(ibytes)
		if werr != nil {
			util.Error("couldn't write bytes: %s", werr)
		} else {
			util.Debug("wrote %d bytes: '%s'", wroteBytes, string(ibytes))
		}
	}
	return
}
