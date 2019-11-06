package combine

import (
	"github.com/stipo42/stringaling/internal/util"
	"io"
)

type StreamCombiner struct {
	Streams []io.Reader
	Output  io.Writer
	Buffer  int64
}

func (c StreamCombiner) Combine() (err error) {
	for _, o := range c.Streams {
		chunk := make([]byte, c.Buffer)
		var rerr error
		for ; rerr == nil; {
			var read int
			read, rerr = o.Read(chunk)
			if rerr != nil {
				if rerr != io.EOF {
					util.Error("error reading bytes: %s", rerr)
					err = rerr
					break
				}
			}
			c.write(chunk, read)
		}
	}
	return
}

func (c StreamCombiner) write(ibytes []byte, writenum ...int) (wroteBytes int) {
	var wn int
	if len(writenum) > 0 {
		wn = writenum[0]
	}
	if wn > len(ibytes) {
		wn = len(ibytes)
	}
	if wn > 0 {
		var werr error
		wroteBytes, werr = c.Output.Write(ibytes[0:wn])
		if werr != nil {
			util.Error("couldn't write bytes: %s", werr)
		} else {
			util.Debug("wrote %d bytes: '%s'", wroteBytes, string(ibytes[0:wn]))
		}
	}
	return
}
