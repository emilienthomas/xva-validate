// Package xva contains functions to work with xva files.
package xva

import (
	"archive/tar"
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	patternChecksum = `^Ref:([0-9]+)/([0-9]{8})\.checksum$`
	patternBlock    = `^Ref:([0-9]+)/([0-9]{8})$`
)

var (
	regexpChecksum = regexp.MustCompile(patternChecksum)
	regexpBlock    = regexp.MustCompile(patternBlock)
)

// Tests integrity of the xva file.
// When verbosity >= 2, outputs each validation performed.
func Validate(xvaFileName string, verbosity uint) (isValid bool, validationIssue string, err error) {

	// Ensure file exists
	f, err := os.OpenFile(xvaFileName, os.O_RDONLY, 0755)
	if err != nil {
		return false, "", err
	}
	defer f.Close()

	if f == nil {
		return false, "", errors.New(fmt.Sprintf("Invalid file: %s", xvaFileName))
	}

	bufferedReader := bufio.NewReader(f)
	if bufferedReader == nil {
		return false, "", errors.New(fmt.Sprintf("Unable to create buffered reader for %s", xvaFileName))
	}

	tarReader := tar.NewReader(bufferedReader)
	if tarReader == nil {
		return false, "", errors.New(fmt.Sprintf("Unable to open %s as tar", xvaFileName))
	}

	sums := make(map[string]string)
	fileContent := make([]byte, 1048576) // Build a 1MB array, since vm disks are store in 1MB chunks
	checksumFromFile := make([]byte, 40)

	if verbosity >= 2 {
		log.Println("Iterating on file content")
	}

	// When err is EOF, tar file is finished
	for header, err := tarReader.Next(); err == nil; header, err = tarReader.Next() {
		if verbosity >= 2 {
			log.Printf("Found %s", header.Name)
		}

		if header.Typeflag == tar.TypeReg {
			// Current entry is a file

			if regexpChecksum.MatchString(header.Name) {
				// Checksum file
				if verbosity >= 3 {
					log.Println("It is a checksum file")
				}
				blockName := strings.Replace(header.Name, ".checksum", "", -1)

				_, err = io.ReadFull(tarReader, checksumFromFile)
				if err != nil {
					return false, "", err
				}
				hexSumFromFile := string(checksumFromFile)

				if len(sums[blockName]) == 0 {
					// Checksum comes first, put it in map
					sums[blockName] = hexSumFromFile
				} else {
					// Checksum comes second, compare it with value in map
					if sums[blockName] != hexSumFromFile {
						if verbosity >= 2 {
							log.Printf("Invalid checksum for %s: expected %s, got %s", blockName, hexSumFromFile, sums[blockName])
						}
						return false, fmt.Sprintf("Invalid checksum for %s: expected %s, got %s", blockName, hexSumFromFile, sums[blockName]), nil
					} else if verbosity >= 2 {
						log.Printf("Checksum valid for %s : %s", blockName, hexSumFromFile)
					}
					// Remove entry from map
					delete(sums, blockName)
				}

			} else if regexpBlock.MatchString(header.Name) {
				// Block file
				if verbosity >= 3 {
					log.Println("It is a block file")
				}
				// Read the complete file into fileContent
				for i, j := 0, 0; err == nil; {
					j, err = tarReader.Read(fileContent[i:header.Size])
					i = i + j
				}
				if err != io.EOF {
					return false, "", err
				}
				fileSum := sha1.Sum(fileContent[:header.Size])
				fileSumAsHex := hex.EncodeToString(fileSum[:])

				if len(sums[header.Name]) == 0 {
					// Data file comes first, put sum in map
					sums[header.Name] = fileSumAsHex
				} else {
					// Data file comes second, compare to sum in map
					if sums[header.Name] != fileSumAsHex {
						if verbosity >= 2 {
							log.Printf("Invalid checksum for %s: expected %s, got %s", header.Name, sums[header.Name], fileSumAsHex)
						}
						return false, fmt.Sprintf("Invalid checksum for %s: expected %s, got %s", header.Name, sums[header.Name], fileSumAsHex), nil
					} else if verbosity >= 2 {
						log.Printf("Checksum valid for %s : %s", header.Name, fileSumAsHex)
					}
					// Remove entry from map
					delete(sums, header.Name)
				}
			}
		}
	}
	if verbosity >= 2 {
		log.Println("Finished iterating")
	}

	// Loop exited, if not on EOF then return error
	if err != nil && err != io.EOF {
		return false, "", err
	}

	// When the whole xva file has been iterated over, no entry should remain in sums.
	if len(sums) > 0 {
		remains := make([]string, len(sums))
		i := 0
		for blockName := range sums {
			remains[i] = blockName
			i++
		}
		return false, fmt.Sprintf("Missing checksums or data blocks: %s", remains), nil
	}

	return true, "", nil
}
