package combine

import (
	"github.com/stipo42/stringaling/internal/util"
	"os"
)

func Combine(files []string, outputFileName string, deleteFiles bool) (err error){
	var outputFile *os.File
	outputFile, err = util.GetCleanFile(outputFileName)
	if err == nil {
		cmbr := StreamCombiner{
			Output: outputFile,
			Buffer: 1024,
		}

		for i := 0; i < len(files); i++ {
			var inputFile *os.File
			inputFile, err = os.Open(files[i])
			defer inputFile.Close()
			if err != nil {
				util.Error("cannot open input file( %s): %s", files[i], err)
				break
			}
			cmbr.Streams = append(cmbr.Streams, inputFile)
		}

		err = cmbr.Combine()

		oerr := outputFile.Close()
		if oerr != nil {
			util.Error("error closing output file (%s): %s", outputFileName, oerr)
		}
	} else {
		util.Error("cannot open output file (%s): %s", outputFileName, err)
	}

	// Cleanup
	if deleteFiles {
		util.Debug("delete flag supplied, deleting input files")
		for i := 0; i < len(files); i++ {
			err = os.Remove(files[i])
			if err != nil {
				util.Error("error deleting file (%s): %s", files[i], err)
			}
		}
	}
	return
}