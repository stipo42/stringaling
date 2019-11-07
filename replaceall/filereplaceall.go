package replaceall

import (
	"errors"
	"fmt"
	"github.com/stipo42/stringaling/combine"
	"github.com/stipo42/stringaling/internal/util"
	"io"
	"math"
	"os"
	"strings"
)

func ReplaceAll(inputFileName string, outputFileName string, startToken string, endToken string, token string, threads int) (err error) {
	confident := false
	useThreads := threads
	useInputFileName := inputFileName
	var ct int
	for ct = 0; ct <= 100; ct++ {
		useInputFileName, confident, err = replaceAllPass(
			ct,
			useInputFileName,
			outputFileName,
			startToken,
			endToken,
			token,
			useThreads,
		)
		if err != nil {
			util.Error("pass %d resulted in an error, aborting: %s", ct, err)
			break
		}
		if confident {
			util.Info("pass %d was confident that all replacements occurred", ct)
			break
		} else {
			if useThreads == 1 {
				util.Info("not confident after 1 thread, giving up")
				break
			} else {
				useThreads = int(math.Ceil(float64(useThreads) / 2))
				util.Info("pass %d was not confident, reducing threads to %d and trying again", ct, useThreads)
			}
		}
	}
	if err == nil {
		err = os.Rename(useInputFileName, outputFileName)
		if err != nil {
			util.Error("could not rename %s to %s: %s", useInputFileName, outputFileName, err)
		}
		for c := 0; c < ct; c++ {
			tFileName := getNextTempFile(outputFileName, c)
			err = os.Remove(tFileName)
			if err != nil {
				util.Error("error deleting temp file %s: %s", tFileName, err)
			}
		}
	}
	return
}

// replaceAllPass executes a single pass of the replaceall function
// multiple passes are used when the confidence of the AllReplacer isn't unified on 'confident'
func replaceAllPass(
	pass int,
	inputFileName string,
	outputFileName string,
	startToken string,
	endToken string,
	token string,
	threads int,
) (
	tempFileName string,
	confident bool,
	err error,
) {
	var stats os.FileInfo
	stats, err = os.Stat(inputFileName)
	if err != nil {
		util.Error("couldn't get file stats on input file (%s): %s", inputFileName, err)
	} else {
		// Need to determine thread size
		tSize := int64(math.Ceil(float64(stats.Size()) / float64(threads)))

		util.Debug("pass-%d: Using a thread size of %d (file size %d)", pass, tSize, stats.Size())

		confidenceChannel := make(chan bool, threads)

		tempFileName = getNextTempFile(outputFileName, pass)

		for i := 0; i < threads; i++ {
			pTempFileName := getNextTempWorkFile(tempFileName, i)

			strgr := &AllReplacer{
				StartToken: startToken,
				EndToken:   endToken,
				Token:      token,
				StartAt:    tSize * int64(i),
				GoUntil:    tSize,
			}

			var threadedOutput *os.File
			strgr.WriterSpawner = func() (writer io.Writer, err error) {
				threadedOutput, err = util.GetCleanFile(pTempFileName)
				if err != nil {
					util.Error("[%d]: couldn't create temp partial file (%s): %s", i, pTempFileName, err)
				}
				return threadedOutput, err
			}
			writerCleanup := func() {
				err = threadedOutput.Close()
				if err != nil {
					util.Error("[%d]: couldn't close temp partial file (%s): %s", i, pTempFileName, err)
				} else {
					util.Debug("[%d]: closed temp partial file (%s)", i, pTempFileName)
				}
			}
			strgr.WriterCleanup = &writerCleanup

			var threadedInput *os.File
			strgr.ReaderSpawner = func() (reader io.Reader, err error) {
				threadedInput, err = os.Open(inputFileName)
				if err != nil {
					util.Error("[%d]: couldn't open input file (%s): %s", i, inputFileName, err)
				}
				return threadedInput, err
			}
			readerCleanup := func() {
				err = threadedInput.Close()
				if err != nil {
					util.Error("[%d]: couldn't close input file (%s): %s", i, inputFileName, err)
				} else {
					util.Debug("[%d]: closed input file (%s)", i, inputFileName)
				}
			}
			strgr.ReaderCleanup = &readerCleanup

			// Do this in it's own thread
			go replaceWorker(*strgr, confidenceChannel, i)

		}
		var eb strings.Builder
		confident = true
		// Consume
		for i := 0; i < threads; i++ {
			c := <-confidenceChannel
			if !c {
				confident = false
			}
		}

		if eb.Len() > 0 {
			err = errors.New(eb.String())
		}

		if err == nil {
			var tempFile *os.File
			tempFile, err = util.GetCleanFile(tempFileName)
			if err == nil {

				// Combine the files
				cmbr := combine.StreamCombiner{
					Output: tempFile,
					Buffer: 1024,
				}
				var tFiles []*os.File
				for i := 0; i < threads; i++ {
					pTempFileName := getNextTempWorkFile(tempFileName, i)

					var pTempFile *os.File
					pTempFile, err = os.Open(pTempFileName)
					if err != nil {
						util.Error("cannot open partial file %s: %s", pTempFileName, err)
						break
					}
					cmbr.Streams = append(cmbr.Streams, pTempFile)
					tFiles = append(tFiles, pTempFile)
				}

				err = cmbr.Combine()
				for _, file := range tFiles {
					terr := file.Close()
					if terr != nil {
						util.Error("error closing temporary partial file (%s): %s", file.Name(), terr)
					}
				}
				terr := tempFile.Close()
				if terr != nil {
					util.Error("error closing temporary output file (%s): %s", tempFile.Name(), terr)
				}
			} else {
				util.Error("cannot open temporary output file (%s): %s", tempFileName, err)
			}

			// Cleanup
			for i := 0; i < threads; i++ {
				pTempFileName := getNextTempWorkFile(tempFileName, i)
				err = os.Remove(pTempFileName)
				if err != nil {
					util.Error("error deleting temp partial file (%s): %s", pTempFileName, err)
				}
			}
		}
	}
	return
}
func getNextTempFile(outputFileName string, pass int) string {
	path, file := splitPath(outputFileName)
	file = fmt.Sprintf("stringalinger_tmp%d_%s", pass, file)
	return path + file
}

func getNextTempWorkFile(outputFileName string, pass int) string {
	path, file := splitPath(outputFileName)
	file = fmt.Sprintf("%d_%s", pass, file)
	return path + file
}

func splitPath(fullpath string) (path string, filename string) {
	pieces := strings.Split(fullpath, "/")
	filename = pieces[len(pieces)-1]
	path = strings.Join(pieces[0:len(pieces)-1], "/")
	if path == "" && len(pieces) > 1 {
		path = "/"
	} else if path != "" {
		path = path + "/"
	}
	return
}

// replaceWorker fires off replaceall.AllReplacer r in a new thread, reporting its confidence back to
// the supplied confidenceChannel reporting on an id
func replaceWorker(r AllReplacer, confidenceChannel chan bool, id int) {
	confident, err := r.Replace(id)
	if err != nil {
		util.Error("[%d]: replacement resulted in an error: %s", id, err)
	}
	confidenceChannel <- confident
}
