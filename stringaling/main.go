package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stipo42/stringaling/combine"
	"github.com/stipo42/stringaling/internal/util"
	"github.com/stipo42/stringaling/replaceall"
)

func main() {
	start := time.Now().UnixNano()
	var cd int
	var err error
	if len(os.Args) > 1 {
		util.DEBUG = getDebugFlag()
		cmd := os.Args[1]
		if cmd == "replace-all" || cmd == "ra" {
			err = doReplaceAll()
		} else if cmd == "combine" || cmd == "c" {
			err = doCombine()
		} else if cmd == "help" {
			printHelp()
		} else {
			err = errors.New("unrecognized command")
		}
		if err != nil {
			util.Error("error executing %s: %s", cmd, err)
			os.Exit(2)
		}
	} else {
		util.Error("Please supply a command. ")
		printHelp()
		os.Exit(1)
	}
	util.Info("stringaling took %s to complete", util.HumanReadable(time.Now().UnixNano()-start))
	os.Exit(cd)
}

func doReplaceAll() (err error) {
	startToken, endToken, inputFileName, outputFileName, token, threads := getReplaceAllArgs()

	if validateReplaceAllArgs(startToken, endToken, inputFileName, outputFileName) {
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
			if confident {
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
		err = os.Rename(useInputFileName, outputFileName)
		if err != nil {
			util.Error("couldnt rename %s to %s: %s", useInputFileName, outputFileName, err)
		}
		for c := 0; c < ct; c++ {
			tFileName := getNextTempFile(outputFileName, c)
			err = os.Remove(tFileName)
			if err != nil {
				util.Error("error deleting temp file %s: %s", tFileName, err)
			}
		}
	} else {
		printReplaceAllHelp()
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
			pTempFileName := fmt.Sprintf("%d_%s", i, tempFileName)

			strgr := replaceall.AllReplacer{
				StartToken: startToken,
				EndToken:   endToken,
				Token:      token,
				StartAt:    tSize * int64(i),
				GoUntil:    tSize,
			}

			strgr.WriterSpawner = func() (writer io.Writer, err error) {
				var tempFile *os.File
				tempFile, err = getCleanFile(pTempFileName)
				if err != nil {
					util.Error("[%d]: couldn't create temp partial file (%s): %s", i, pTempFileName, err)
				}
				cleanup := func() {
					err = tempFile.Close()
					if err != nil {
						util.Error("[%d]: couldn't close temp output file (%s): %s", i, pTempFileName, err)
					}
				}
				strgr.WriterCleanup = &cleanup
				return tempFile, err
			}

			strgr.ReaderSpawner = func() (reader io.Reader, err error) {
				var inputFile *os.File
				inputFile, err = os.Open(inputFileName)
				if err != nil {
					util.Error("[%d]: couldn't open input file (%s): %s", i, inputFileName, err)
				}
				cleanup := func() {
					err = inputFile.Close()
					if err != nil {
						util.Error("[%d]: couldn't close input file: %s", i, err)
					}
				}
				strgr.ReaderCleanup = &cleanup
				return inputFile, err
			}

			// Do this in it's own thread
			go replaceWorker(strgr, confidenceChannel, i)

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
			tempFile, err = getCleanFile(tempFileName)
			if err == nil {

				defer tempFile.Close()
				// Combine the files
				cmbr := combine.StreamCombiner{
					Output: tempFile,
					Buffer: 1024,
				}

				for i := 0; i < threads; i++ {
					pTempFileName := fmt.Sprintf("%d_%s", i, tempFileName)

					var pTempFile *os.File
					pTempFile, err = os.Open(pTempFileName)
					defer pTempFile.Close()
					if err != nil {
						util.Error("cannot open partial file %s: %s", pTempFileName, err)
						break
					}
					cmbr.Streams = append(cmbr.Streams, pTempFile)
				}

				err = cmbr.Combine()
			} else {
				util.Error("cannot open file %s: %s", tempFileName, err)
			}

			// Cleanup
			for i := 0; i < threads; i++ {
				pTempFileName := fmt.Sprintf("%d_%s", i, tempFileName)
				err = os.Remove(pTempFileName)
				if err != nil {
					util.Error("error deleteing temp work file: %s", err)
				}
			}
		}
	}
	return
}

func getNextTempFile(outputFileName string, pass int) string {
	return fmt.Sprintf("stringalinger_tmp%d_%s", pass, outputFileName)
}

// replaceWorker fires off replaceall.AllReplacer r in a new thread, reporting its confidence back to
// the supplied confidenceChannel reporting on an id
func replaceWorker(r replaceall.AllReplacer, confidenceChannel chan bool, id int) {
	confident, err := r.Replace(id)
	if err != nil {
		util.Error("[%d]: replacement resulted in an error: %s", id, err)
	}
	confidenceChannel <- confident
}

// getReplaceAllArgs gets the arguments from the os.Args slice relevant to the replaceall command
func getReplaceAllArgs() (startToken string, endToken string, inputFile string, outputFile string, replaceWith string, threads int) {
	args := os.Args[2:]
	skip := false
	for a, arg := range args {
		if skip {
			skip = false
			continue
		}
		isFlag := strings.Index(arg, "-") == 0
		if isFlag {
			if arg == "-s" {
				skip = true
				startToken = args[a+1]
			} else if arg == "-e" {
				skip = true
				endToken = args[a+1]
			} else if arg == "-i" {
				skip = true
				inputFile = args[a+1]
			} else if arg == "-o" {
				skip = true
				outputFile = args[a+1]
			} else if arg == "-w" {
				skip = true
				replaceWith = args[a+1]
			} else if arg == "-t" {
				skip = true
				var err error
				threads, err = strconv.Atoi(args[a+1])
				if err != nil || threads <= 0 {
					threads = 1
				}
			}
			util.Debug("found %s, set to %s", arg, args[a+1])
		}
	}
	if threads == 0 {
		threads = 1
	}
	return
}

func validateReplaceAllArgs(startToken string, endToken string, inputFile string, outputFile string) bool {
	util.Debug("-s %s -e %s -i %s -o %s", startToken, endToken, inputFile, outputFile)
	return inputFile != "" && outputFile != "" && startToken != "" && endToken != ""
}

func doCombine() (err error) {
	files, outputFileName, deleteFiles := getCombineArgs()
	if validateCombineArgs(files, outputFileName) {
		var outputFile *os.File
		outputFile, err = getCleanFile(outputFileName)
		if err == nil {
			defer outputFile.Close()
			cmbr := combine.StreamCombiner{
				Output: outputFile,
				Buffer: 1024,
			}

			for i := 0; i < len(files); i++ {
				var inputFile *os.File
				inputFile, err = os.Open(files[i])
				defer inputFile.Close()
				if err != nil {
					util.Error("cannot open file %s: %s", files[i], err)
					break
				}
				cmbr.Streams = append(cmbr.Streams, inputFile)
			}

			err = cmbr.Combine()
		} else {
			util.Error("cannot open outputfile %s: %s", outputFileName, err)
		}

		// Cleanup
		if deleteFiles {
			util.Debug("delete flag supplied, deleting input files")
			for i := 0; i < len(files); i++ {
				err = os.Remove(files[i])
				if err != nil {
					util.Error("error deleteing file: %s", err)
				}
			}
		}
	} else {
		printCombineHelp()
	}
	return
}

func getCombineArgs() (files []string, outputFile string, deleteFiles bool) {
	args := os.Args[2:]
	skip := false
	for a, arg := range args {
		if skip {
			skip = false
			continue
		}
		isFlag := strings.Index(arg, "-") == 0
		if isFlag {
			if arg == "-f" {
				skip = true
				files = append(files, args[a+1])
			} else if arg == "-d" {
				deleteFiles = true
			} else if arg == "-o" {
				skip = true
				outputFile = args[a+1]
			}
		}
	}
	return
}

func validateCombineArgs(files []string, outputFile string) bool {
	return len(files) > 1 && outputFile != ""
}

// getDebugFlag returns true if the verbose flag was supplied
func getDebugFlag() bool {
	for _, arg := range os.Args {
		if arg == "-v" {
			return true
		}
	}
	return false
}
func printHelp() {
	fmt.Println("")
	fmt.Println("This utility is a streaming string replacement tool.")
	fmt.Println("Massive amounts of memory are not needed for extremely large files.")
	fmt.Println("")
	fmt.Println(fmt.Sprintf("Usage : %s COMMAND [-v] ...", os.Args[0]))
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("        COMMAND : A command to perform by stringaling")
	fmt.Println("        -v      : Verbose output")
	fmt.Println("")
	fmt.Println("Available Commands:")
	fmt.Println("        replace-all, ra  - This will replace all characters between two tokens, including those tokens. ")
	fmt.Println("        combine, c       - This will combine a set of files into a single file, in the order provided. ")
	fmt.Println("        help             - This will show this help screen")
	fmt.Println("")

}

func printReplaceAllHelp() {
	fmt.Println("")
	fmt.Println("replaceall,rall - This will replace all characters between two tokens, including those tokens.")
	fmt.Println("                  this streams the input one byte at a time, which is why there are strict limitations. ")
	fmt.Println("")
	fmt.Println("This command does NOT support REGEX and requires strict tokens to be given for marking the beginning and end of replacement.")
	fmt.Println("This command supports the beginning and end tokens being the same token.")
	fmt.Println("")
	fmt.Println(fmt.Sprintf("Usage : %s replace-all|ra -i INPUTFILE -o OUTPUTFILE -s STARTTOKEN -e ENDTOKEN [-w TOKEN]", os.Args[0]))
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("        -i INPUTFILE  : The file to stringaling process ")
	fmt.Println("        -o OUTPUTFILE : The file to write the result of the stringaling process to.")
	fmt.Println("        -s STARTTOKEN : The token to mark the beginning of replacement. ")
	fmt.Println("        -e ENDTOKEN   : The token to mark the end of replacement. ")
	fmt.Println("        -w TOKEN      : The token to replace the marked characters with, if not supplied, defaults to emptystring. ")
	fmt.Println("        -t THREADS    : (Experimental) The number of threads to split work against. The higher this count, ")
	fmt.Println("                        the less accurate replacement is, as it is unknown if the start of a thread should be written. ")
	fmt.Println("                        However, the more threads there are, the faster the program will complete. ")
	fmt.Println("                        For peak performance, set this to the total number of cores available. ")
	fmt.Println("")
}

func printCombineHelp() {
	fmt.Println("")
	fmt.Println("combine,c - This will combine a set of files into a single file, optionally deleting the originals. ")
	fmt.Println("")
	fmt.Println("This command does NOT support REGEX and requires strict filenames, if regex is required it must be used before this command.")
	fmt.Println("")
	fmt.Println(fmt.Sprintf("Usage : %s combine|c [-d] -f FILENAME [-f FILENAME]... -o OUTPUTFILE", os.Args[0]))
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("        -d            : Deletes the source files when supplied.")
	fmt.Println("        -f FILENAME   : Adds a file to the combination pool.")
	fmt.Println("        -o OUTPUTFILE : Sets the name of the file to write the combination to.")
	fmt.Println("")
}

// getCleanFile gets a file by fileName, deleting it first if it already exists.
func getCleanFile(fileName string) (file *os.File, err error) {
	_, err = os.Stat(fileName)
	if err == nil {
		util.Debug("%s exists, deleting it", fileName)
		rerr := os.Remove(fileName)
		if rerr != nil {
			util.Debug("cannot remove %s: %s", fileName, rerr)
		} else {
			err = errors.New("ok")
		}
	}
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			util.Debug("error creating file %s: %s", fileName, err)
		} else {
			err = file.Chmod(0777)
		}
	}
	return
}
