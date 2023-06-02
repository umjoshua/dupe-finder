package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func findDuplicates(targetPath string) (map[string][]string, error) {
	duplicates := make(map[string][]string)

	fileChan := make(chan string)
	hashChan := make(chan string)

	wg := sync.WaitGroup{}

	go func() {
		for file := range fileChan {
			fileHash, err := hashFile(file)
			if err != nil {
				continue
			}
			hashChan <- fileHash
			hashChan <- file
			wg.Done()
		}

	}()

	go func() {
		for hashVal := range hashChan {
			filePath := <-hashChan
			duplicates[hashVal] = append(duplicates[hashVal], filePath)
		}
	}()

	err := filepath.Walk(targetPath, func(fileName string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if err != nil {
			return err
		}

		wg.Add(1)
		fileChan <- fileName

		return nil
	})

	close(fileChan)
	wg.Wait()
	close(hashChan)

	if err != nil {
		fmt.Println(err)
	}
	return duplicates, nil
}

func hashFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashValue := hash.Sum(nil)
	return fmt.Sprintf("%X", hashValue), nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: dupe-finder [path]")
		return
	}

	targetDirectory := os.Args[1]

	if _, err := os.Stat(targetDirectory); os.IsNotExist(err) {
		fmt.Println("Invalid Path")
		return
	}
	duplicates, err := findDuplicates(targetDirectory)
	if err != nil {
		fmt.Println("Couldn't Complete")
		return
	}

	count := 0
	for _, f := range duplicates {
		if len(f) > 1 {
			count++
			fmt.Println(f[1], ": Duplicate found:-> ", strings.Join(f, ", "))
		}
	}

	if count == 0 {
		return
	}

	fmt.Printf("Delete %v duplicates? (Y/n): ", count)
	var input string
	fmt.Scan(&input)

	if input == "Y" || input == "y" {
		for _, f := range duplicates {
			if len(f) > 1 {
				for _, fi := range f[1:] {
					err = os.Remove(fi)
					if err != nil {
						fmt.Println("Couldn't delete ", fi)
					}
				}
			}
		}
	}
}
