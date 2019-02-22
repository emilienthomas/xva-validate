// Package xva contains functions to work with xva files.
package xva

import (
	"archive/tar"
	"bufio"
	"crypto/sha1"
	"encoding/base64"
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
func Validate(xvaFileName string, verbosity uint) (isValid bool, err error) {

	// Ensure file exists
	f, err := os.OpenFile(xvaFileName, os.O_RDONLY, 0755)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if f == nil {
		return false, errors.New(fmt.Sprintf("Invalid file: %s", xvaFileName))
	}

	bufferedReader := bufio.NewReader(f)
	if bufferedReader == nil {
		return false, errors.New(fmt.Sprintf("Unable to create buffered reader for %s", xvaFileName))
	}

	tarReader := tar.NewReader(bufferedReader)
	if tarReader == nil {
		return false, errors.New(fmt.Sprintf("Unable to open %s as tar", xvaFileName))
	}

	sums := make(map[string]string)
	fileContent := make([]byte, 1048576) // Build a 1MB array
	checksumFromFile := make([]byte, 40)

	if verbosity >= 2 {
		log.Println("Iterating on file content")
	}

	for {
		header, err := tarReader.Next()

		// When err is EOF, tar file is finished
		if err == io.EOF {
			if verbosity >= 2 {
				log.Println("EOF")
			}
			break
		}
		// Otherwise return current error
		if err != nil {
			return false, err
		}

		if header.Typeflag == tar.TypeReg {
			// Current entry is a file

			if regexpChecksum.MatchString(header.Name) {
				// Checksum file
				blockName := strings.Replace(header.Name, ".checksum", "", -1)

				_, err = io.ReadFull(tarReader, checksumFromFile)
				if err != nil {
					return false, err
				}
				base64sum := string(checksumFromFile)

				// Successfully read checksum in xva file
				if len(sums[blockName]) == 0 {
					// Checksum comes first, put it in map
					sums[blockName] = base64sum
				} else {
					// Checksum comes second, compare it with value in map
					if sums[blockName] == base64sum {
						return false, errors.New(fmt.Sprintf("Invalid checksum for %s: expected %s, got %s", blockName, base64sum, sums[blockName]))
					} else if verbosity >= 2 {
						log.Printf("Checksum valid for %s : %s", blockName, base64sum)
					}
					// Remove entry from map
					delete(sums, blockName)
				}

			} else if regexpBlock.MatchString(header.Name) {
				// Block file

				_, err = io.ReadFull(tarReader, fileContent)
				if err != nil {
					return false, err
				}

				fileSum := sha1.Sum(fileContent)
				fileSumAsBase64 := base64.StdEncoding.EncodeToString(fileSum[:])

				if len(sums[header.Name]) == 0 {
					// Data file comes first, put sum in map
					sums[header.Name] = fileSumAsBase64
				} else {
					// Data file comes second, compare to sum in map
					if sums[header.Name] == fileSumAsBase64 {
						return false, errors.New(fmt.Sprintf("Invalid checksum for %s: expected %s, got %s", header.Name, sums[header.Name], fileSumAsBase64))
					} else if verbosity >= 2 {
						log.Printf("Checksum valid for %s : %s", header.Name, fileSumAsBase64)
					}
					// Remove entry from map
					delete(sums, header.Name)
				}
			}
		}
	}

	// When the whole xva file has been iterated over, no entry should remain in sums.
	if len(sums) > 0 {
		remains := make([]string, len(sums))
		i := 0
		for blockName := range sums {
			remains[i] = blockName
			i++
		}
		return false, errors.New(fmt.Sprintf("Missing checksums or data blocks: %s", remains))
	}

	return true, nil
}
