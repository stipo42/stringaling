package main

import (
	"errors"
	"fmt"
	"github.com/stipo42/stringaling/internal/util"
	"github.com/stipo42/stringaling/replaceall"
	"os"
	"strings"
)

func main() {
	var err error
	if len(os.Args) > 1 {
		util.DEBUG = getDebugFlag()
		cmd := os.Args[1]
		if cmd == "replaceall" || cmd == "rall" {
			startToken, endToken, inputFileName, outputFileName, replaceWith := getReplaceAllArgs()

			if validateArgs(startToken, endToken, inputFileName, outputFileName) {
				var inputFile *os.File
				inputFile, err = os.Open(inputFileName)
				if err != nil {
					util.Error("couldn't open input file %s: %s", inputFile, err)
				} else {
					defer inputFile.Close()
					var outputFile *os.File
					outputFile, err = os.Open(outputFileName)
					if err != nil {
						outputFile, err = os.Create(outputFileName)
						if err != nil {
							util.Error("couldn't create or open output file %s: %s", outputFileName, err)
						}
					}
					if outputFile != nil {
						defer outputFile.Close()

						//var stats os.FileInfo
						//stats,err = inputFile.Stat()
						//
						//stats.Size()
						strgr := replaceall.AllReplacer{
							Reader:     inputFile,
							Writer:     outputFile,
							StartToken: startToken,
							EndToken:   endToken,
							Token:      replaceWith,
						}

						err = strgr.Replace()
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
	os.Exit(0)
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
func getReplaceAllArgs() (startToken string, endToken string, inputFile string, outputFile string, replaceWith string) {
	args := os.Args[2:]
	skip := false
	flaggingDone := false
	for a, arg := range args {
		if !flaggingDone {
			if skip {
				skip = false
				continue
			}
			isFlag := strings.Index(arg, "-") == 0
			if isFlag {
				if arg == "-s" {
					skip = true
					startToken = args[a+1]
					util.Debug("found %s, set to %s", arg, args[a+1])
					continue
				} else if arg == "-e" {
					skip = true
					endToken = args[a+1]
					util.Debug("found %s, set to %s", arg, args[a+1])
					continue
				} else if arg == "-i" {
					skip = true
					inputFile = args[a+1]
					util.Debug("found %s, set to %s", arg, args[a+1])
					continue
				} else if arg == "-o" {
					skip = true
					outputFile = args[a+1]
					util.Debug("found %s, set to %s", arg, args[a+1])
					continue
				} else if arg == "-w" {
					skip = true
					replaceWith = args[a+1]
					util.Debug("found %s, set to %s", arg, args[a+1])
					continue
				}
			} else {
				flaggingDone = true
			}
		}
		if flaggingDone {
			if startToken == "" {
				startToken = arg
			} else if endToken == "" {
				endToken = arg
			} else if inputFile == "" {
				inputFile = arg
			} else if outputFile == "" {
				outputFile = arg
			} else if replaceWith == "" {
				replaceWith = arg
			}
		}
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
	fmt.Println("Usage : stringaling replaceall|rall -i INPUTFILE -o OUTPUTFILE -s STARTTOKEN -e ENDTOKEN -w TOKEN")
	fmt.Println(" or   : stringaling replaceall|rall STARTTOKEN ENDTOKEN INPUTFILE OUTPUTFILE [TOKEN]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("        -i INPUTFILE  : The file to stringaling process ")
	fmt.Println("        -o OUTPUTFILE : The file to write the result of the stringaling process to.")
	fmt.Println("        -s STARTTOKEN : The token to mark the beginning of replacement. ")
	fmt.Println("        -e ENDTOKEN   : The token to mark the end of replacement. ")
	fmt.Println("        -w TOKEN      : The token to replace the marked characters with, if not supplied, defaults to emptystring. ")
	fmt.Println("")
}
