package replaceall

import (
	"io"
	"time"

	"github.com/stipo42/stringaling/internal/util"
)

type AllReplacer struct {
	StartAt       int64 // The byte number to start at
	GoUntil       int64 // The byte number to consume
	StartToken    string
	EndToken      string
	Token         string
	ReaderSpawner func() (io.Reader, error)
	WriterSpawner func() (io.Writer, error)
	ReaderCleanup *func()
	WriterCleanup *func()
}

// Replace performs the replacement for the configured AllReplacer
// optionally an id may be supplied for keeping track of threading when
// output is verbose
// This function returns a flag representing the confidence that this function
// caught all the replacements.
// When the function is not confident it basically means the start and end token count
// is uneven.
func (s AllReplacer) Replace(id ...int) (confident bool, err error) {
	start := time.Now().UnixNano()
	if s.StartToken == s.EndToken {
		confident, err = s.replaceSameStartEnd(id...)
	} else {
		confident, err = s.replace(id...)
	}
	if s.ReaderCleanup != nil {
		util.Debug("%d: running reader cleanup", id)
		c := *s.ReaderCleanup
		c()
	}
	if s.WriterCleanup != nil {
		util.Debug("%d: running writer cleanup", id)
		c := *s.WriterCleanup
		c()
	}
	diff := time.Now().UnixNano() - start
	util.Info("Replace took %s to execute", util.HumanReadable(diff))

	return
}

// SpawnReader will spawn a new io.Reader for the AllReplacer to read from.
func (s AllReplacer) SpawnReader() (reader io.Reader, err error) {
	return s.ReaderSpawner()
}

// SpawnWriter will spawn a new io.Writer for the AllReplacer to write to.
func (s AllReplacer) SpawnWriter() (writer io.Writer, err error) {
	return s.WriterSpawner()
}

// replace is used to replace content between different start and end tokens
func (s AllReplacer) replace(id ...int) (confident bool, err error) {
	noWriteDepth := 0 // When not zero, don't write byte to output
	chunk := make([]byte, 1)
	sct := 0 // Increase every time the consecutively read byte matches that index of the start token, when longer than start token length, depth increases
	ect := 0 // Increase every time the consecutively read byte matches that index of the end token,   when longer than end   token length, depth decreases
	skipped := 0

	slen := len(s.StartToken)
	elen := len(s.EndToken)

	byteCtr := int64(0)
	var reader io.Reader
	var writer io.Writer
	reader, err = s.SpawnReader()
	// Holds the temporary skipped bytes
	// Only holds data when skipping, drops its value when not skipping
	// If the end of the stream is hit while skipping, it is written to the output
	var cupdate []byte
	if err != nil {
		util.Error("%d: could not spawn a reader struct: %s", id, err)
	} else {
		writer, err = s.SpawnWriter()
		if err != nil {
			util.Error("%d: could not spawn a writer struct: %s", id, err)
		} else {
			rerr := s.fastForward(reader)
			for rerr == nil {
				var b int
				b, rerr = reader.Read(chunk)
				byteCtr += int64(b)
				if rerr != nil {
					if rerr == io.EOF {
						util.Debug("%d: end of file: %s", id, rerr)
					} else {
						util.Error("%d: couldn't read chunk: %s", id, rerr)
						err = rerr
					}
					if len(cupdate) > 0 {
						s.write(cupdate, writer, id...)
					}
				} else {
					if noWriteDepth > 0 {
						skipped += 1
					}
					startBackfill := s.missCheck(chunk[0], noWriteDepth, s.StartToken, &sct, &ect, id...)
					if len(startBackfill) > 0 {
						cupdate = removeLastIndexes(cupdate, len(startBackfill))
						s.write(startBackfill, writer, id...)
					}
					endBackfill := s.missCheck(chunk[0], noWriteDepth, s.EndToken, &ect, &sct, id...)
					if len(endBackfill) > 0 {
						cupdate = removeLastIndexes(cupdate, len(endBackfill))
						s.write(endBackfill, writer, id...)
					}
					if sct >= slen {
						sct = 0
						noWriteDepth += 1
					}
					if ect >= elen {
						ect = 0
						noWriteDepth -= 1
						if noWriteDepth < 0 {
							// Mismatched end to start, write end back, reduce cupdate
							cupdate = removeLastIndexes(cupdate, elen-1)
							s.writeS(s.EndToken, writer, id...)
							noWriteDepth = 0
						} else if noWriteDepth <= 0 {
							skipped += slen
							util.Debug("%d: replaced %d bytes", id, skipped)
							s.writeS(s.Token, writer, id...)
							cupdate = nil
						}
					} else if noWriteDepth == 0 && sct == 0 && ect == 0 {
						s.write(chunk, writer, id...)
					} else {
						util.Debug("%d: Appending '%s' to cupdate, noWriteDepth = %d, sct = %d, ect = %d, cupdate = %s", id, string(chunk), noWriteDepth, sct, ect, string(cupdate))
						cupdate = append(cupdate, chunk[0])
					}
				}
				if byteCtr >= s.GoUntil {
					util.Debug("%d: Hit end of byte duty", id)
					confident = len(cupdate) == 0
					if len(cupdate) > 0 {
						s.write(cupdate, writer, id...)
					}
					break
				}
			}
		}
	}

	return
}

// replaceSameStartEnd is used when the start and end tokens are the same,
// the logic is slightly different / simplified in this scenario
func (s AllReplacer) replaceSameStartEnd(id ...int) (confident bool, err error) {
	noWriteDepth := 0 // When not zero, don't write byte to output
	chunk := make([]byte, 1)
	ct := 0 // Increase every time the consecutively read byte matches that index of the start token, when longer than start token length, depth increases
	skipped := 0

	slen := len(s.StartToken)

	byteCtr := int64(0)
	var reader io.Reader
	var writer io.Writer
	reader, err = s.SpawnReader()
	// Holds the temporary skipped bytes
	// Only holds data when skipping, drops its value when not skipping
	// If the end of the stream is hit while skipping, it is written to the output
	var cupdate []byte
	if err != nil {
		util.Error("could not spawn a reader struct: %s", err)
	} else {
		writer, err = s.SpawnWriter()
		if err != nil {
			util.Error("%d: could not spawn a writer struct: %s", id, err)
		} else {
			rerr := s.fastForward(reader)
			for rerr == nil {
				var b int
				b, rerr = reader.Read(chunk)
				byteCtr += int64(b)
				if rerr != nil {
					if rerr == io.EOF {
						util.Debug("%d: End of file: %s", id, rerr)
					} else {
						util.Error("%d: Couldn't read chunk: %s", id, rerr)
						err = rerr
					}
					if len(cupdate) > 0 {
						s.write(cupdate, writer, id...)
					}
				} else {
					if noWriteDepth > 0 {
						skipped += 1
					}
					backfill := s.missCheck(chunk[0], noWriteDepth, s.StartToken, &ct, nil, id...)
					if len(backfill) > 0 {
						cupdate = removeLastIndexes(cupdate, len(backfill))
						s.write(backfill, writer, id...)
					}
					if ct >= slen {
						ct = 0
						if noWriteDepth == 0 {
							noWriteDepth = 1
						} else if noWriteDepth == 1 {
							skipped += slen
							util.Debug("%d: replaced %d bytes", id, skipped)
							s.writeS(s.Token, writer, id...)
							cupdate = nil
							noWriteDepth = 0
						}
					} else if noWriteDepth == 0 && ct == 0 {
						s.write(chunk, writer, id...)
					} else {
						util.Debug("%d: Appending '%s' to cupdate, noWriteDepth = %d, ct = %d, cupdate = %s", id, string(chunk), noWriteDepth, ct, string(cupdate))
						cupdate = append(cupdate, chunk[0])
					}
				}
				if byteCtr >= s.GoUntil {
					util.Debug("%d: Hit end of byte duty", id)
					confident = len(cupdate) == 0
					if len(cupdate) > 0 {
						s.write(cupdate, writer, id...)
					}
					break
				}
			}
		}
	}

	return
}

// missCheck handles misses on tokens
// a missCheck can happen even when counts are at zero, so the reset return value indicates if a reset happened
func (s AllReplacer) missCheck(thisByte byte, noWriteDepth int, myToken string, myCount *int, otherCount *int, id ...int) (backfill []byte) {
	var miss bool
	if thisByte == myToken[*myCount] {
		*myCount += 1
	} else {
		miss = true
	}
	if miss && *myCount > 0 && noWriteDepth == 0 {
		if otherCount == nil || *otherCount == 0 {
			// Backfill for this miss.
			backfill = []byte(myToken[0:*myCount])
			util.Debug("%d: token '%s' missed on %s, need to backfill '%s'", id, myToken, string(thisByte), backfill)
		} else {
			util.Debug("%d: token '%s' missed on %s, but it partially matches other token, not backfilling", id, myToken, string(thisByte))
		}
	}
	if miss {
		*myCount = 0
	}
	return
}

func (s AllReplacer) writeS(str string, writer io.Writer, id ...int) (wroteBytes int) {
	return s.write([]byte(str), writer, id...)
}
func (s AllReplacer) write(ibytes []byte, writer io.Writer, id ...int) (wroteBytes int) {
	if len(ibytes) > 0 {
		var werr error
		wroteBytes, werr = writer.Write(ibytes)
		if werr != nil {
			util.Error("%d: couldn't write bytes: %s", id, werr)
		} else {
			util.Debug("%d: wrote %d bytes: '%s'", id, wroteBytes, string(ibytes))
		}
	}
	return
}

func (s AllReplacer) fastForward(reader io.Reader) (err error) {
	var rerr error
	if s.StartAt > 0 {
		fastForward := make([]byte, s.StartAt)
		_, rerr = reader.Read(fastForward)
		if rerr != nil {
			if rerr == io.EOF {
				util.Debug("Fast forwarded past end of file: %s", rerr)
			} else {
				util.Error("couldn't fast forward: %s", rerr)
			}
			err = rerr
		}
		// Deallocate this because we don't care
		fastForward = nil
	}
	return
}

func removeLastIndexes(slice []byte, rcount int) []byte {
	if len(slice) > 0 && rcount > 0 {
		top := len(slice) - rcount
		tcup := make([]byte, top)
		copy(tcup, slice[0:top])
		slice = tcup
	}
	return slice
}
