package main

import (
	"errors"
	"fmt"
	"github.com/stipo42/stringaling/combine"
	"github.com/stipo42/stringaling/internal/util"
	"github.com/stipo42/stringaling/replaceall"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	start := time.Now().UnixNano()
	var cd int
	var err error
	if len(os.Args) > 1 {
		util.DEBUG = getDebugFlag()
		cmd := os.Args[1]
		if cmd == "replaceall" || cmd == "rall" {
			startToken, endToken, inputFileName, outputFileName, replaceWith, threads := getReplaceAllArgs()

			if validateArgs(startToken, endToken, inputFileName, outputFileName) {

				var stats os.FileInfo
				stats, err = os.Stat(inputFileName)
				if err != nil {
					util.Error("couldn't get file stats on input file (%s): %s", inputFileName, err)
				} else {
					// Need to determine thread size
					tSize := stats.Size() / int64(threads)

					confidenceChannel := make(chan bool, threads)

					for i := 0; i < threads; i++ {
						pOutputFile := fmt.Sprintf("%d_%s", i, outputFileName)

						strgr := replaceall.AllReplacer{
							StartToken: startToken,
							EndToken:   endToken,
							Token:      replaceWith,
							StartAt:    tSize * int64(i),
							GoUntil:    tSize,
						}

						strgr.WriterSpawner = func() (writer io.Writer, err error) {
							var outputFile *os.File
							outputFile, err = getCleanFile(pOutputFile)
							if err != nil {
								util.Error("[%d]: couldn't create temp partial file (%s): %s", i, pOutputFile, err)
							}
							cleanup := func() {
								err = outputFile.Close()
								if err != nil {
									util.Error("[%d]: couldn't close temp output file (%s): %s", i, pOutputFile, err)
								}
							}
							strgr.WriterCleanup = &cleanup
							return outputFile, err
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
						go Replace(strgr, confidenceChannel, i)

					}
					var eb strings.Builder
					confident := true
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
						var outputFile *os.File
						outputFile, err = getCleanFile(outputFileName)
						if err == nil {

							defer outputFile.Close()
							// Combine the files
							cmbr := combine.StreamCombiner{
								Output: outputFile,
								Buffer: 1024,
							}

							for i := 0; i < threads; i++ {
								pOutputFile := fmt.Sprintf("%d_%s", i, outputFileName)

								var pFile *os.File
								pFile, err = getCleanFile(pOutputFile)
								if err != nil {
									break
								}
								cmbr.Streams = append(cmbr.Streams, pFile)
							}

							err = cmbr.Combine()
						}
					}

					if !confident {
						util.Info("First round was not confident, going again!")
					}
				}

			} else {
				printReplaceAllHelp()
			}
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

func Replace(r replaceall.AllReplacer, rChan chan bool, id int) {
	confident, err := r.Replace(id)
	if err != nil {
		util.Error("A replacement resulted in an error: %s", err)
	}
	rChan <- confident
}

func getDebugFlag() bool {
	for _, arg := range os.Args {
		if arg == "-v" {
			return true
		}
	}
	return false
}

// getReplaceAllArgs can accept input in the following order:
// 0: startToken
// 1: endToken
// 2: inputFile
// 3: outputFile
// 4: replaceWith
// OR flagging can be used in any order, taking precedence over direct arguments
// the next index is then used as the flag value where applicable
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

func validateArgs(startToken string, endToken string, inputFile string, outputFile string) bool {
	util.Debug("-s %s -e %s -i %s -o %s", startToken, endToken, inputFile, outputFile)
	return inputFile != "" && outputFile != "" && startToken != "" && endToken != ""
}

func printHelp() {
	fmt.Println("")
	fmt.Println("This utility is a streaming string replacement tool.")
	fmt.Println("Massive amounts of memory are not needed for extremely large files.")
	fmt.Println("")
	fmt.Println("Usage : stringaling COMMAND [-v] ...")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("        COMMAND : A command to perform by stringaling")
	fmt.Println("        -v      : Verbose output")
	fmt.Println("")
	fmt.Println("Available Commands:")
	fmt.Println("        replaceall, rall - This will replace all characters between two tokens, including those tokens. ")
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
	fmt.Println("Usage : stringaling replaceall|rall -i INPUTFILE -o OUTPUTFILE -s STARTTOKEN -e ENDTOKEN [-w TOKEN]")
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

func getCleanFile(fileName string) (file *os.File, err error) {
	_, err = os.Stat(fileName)
	if err == nil {
		util.Debug("%s exists, deleting it", fileName)
		rerr := os.Remove(fileName)
		if rerr != nil {
			util.Debug("cannot delete %s", fileName)
		} else {
			err = errors.New("ok")
		}
	}
	if err != nil {
		file, err = os.Create(fileName)
	}
	return
}
