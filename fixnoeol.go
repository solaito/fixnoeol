package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func isText(fp *os.File) (bool, error) {
	b := make([]byte, 512)

	n, err := fp.Read(b)
	if err != nil {
		return false, err
	} else if n == 0 {
		return false, nil
	}

	// MEMO: https://gist.github.com/rayrutjes/db9b9ea8e02255d62ce2
	ct := http.DetectContentType(b[:n])
	fp.Seek(0, 0)

	return strings.HasPrefix(ct, "text"), nil
}

func isNoEol(fp *os.File) (bool, error) {
	fp.Seek(-1, 2)
	b, err := ioutil.ReadAll(fp)
	if err != nil {
		return false, err
	}
	return string(b) != "\n", nil
}

func isEmpty(info fs.DirEntry) (bool, error) {
	fi, err := info.Info()
	if err != nil {
		return false, err
	}

	return fi.Size() == 0, nil
}

func check(path string, info fs.DirEntry) bool {
	fmt.Println("fixNoEol path: " + path)

	if info.IsDir() {
		return false
	}

	if r, err := isEmpty(info); err != nil {
		return false
	} else if r {
		return false
	}

	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	if r, err := isText(f); err != nil {
		return false
	} else if !r {
		return false
	}

	if r, err := isNoEol(f); err != nil {
		return false
	} else if !r {
		return false
	}

	return true
}

func fix(path string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("err ", err)
	}
	defer f.Close()

	if _, err := fmt.Fprint(f, "\n"); err != nil {
		fmt.Println(err)
	}
}

func callback(path string, info fs.DirEntry, err error) error {
	if check(path, info) {
		fix(path)
	}

	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: fixnoeol <file/dir>...")
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
	}
	for _, root := range flag.Args() {
		_, err := os.Stat(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: No such file or directory\n", root)
			continue
		}
		filepath.WalkDir(root, callback)
	}
}
