package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Pair struct {
	Hash, Path string
}
type fileList []string
type results map[string]fileList

func hashFile(path string) Pair {
	file, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}

	return Pair{fmt.Sprintf("%x", hash.Sum(nil)), path}
}
func searchTree(dir string) (results, error) {
	hashes := make(results)
	visit := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Mode().IsRegular() && fi.Size() > 0 {
			h := hashFile(path)
			hashes[h.Hash] = append(hashes[h.Hash], h.Path)
		}
		return nil
	}

	err := filepath.Walk(dir, visit)

	return hashes, err
}
func main() {
	//var hashes
	dirName := "test"

	if hashes, err := searchTree(dirName); err == nil {
		for hash, files := range hashes {
			if len(files) > 1 {
				fmt.Println(hash, len(files))

				for _, file := range files {
					fmt.Println(" ", file)
				}
			}
		}
	}

}
