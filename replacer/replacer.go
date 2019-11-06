package replacer

import "io"

type Replacer interface {
	Replace() error
	SpawnReader() (io.Reader, error)
	SpawnWriter() (io.Writer, error)
}
