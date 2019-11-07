package main

import (
	"errors"
	"fmt"
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
		err = replaceall.ReplaceAll(inputFileName, outputFileName, startToken, endToken, token, threads)
	} else {
		printReplaceAllHelp()
	}
	return
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
		err = combine.Combine(files, outputFileName, deleteFiles)
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
