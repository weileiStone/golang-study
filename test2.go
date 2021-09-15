package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var schem = make(chan struct{}, 20)
func main()  {
	var wg sync.WaitGroup
	var dir *string
	var tagFile *string
	var findMod *string
	currtPath, _ := GetCurrentPath()
	dir = flag.String("dir", currtPath, "find dir")
	tagFile = flag.String("filename", "", "find file name")
	findMod = flag.String("findmod", "exact", "find file mod exact /suffix")
	flag.Parse()
	if *tagFile == ""{
		fmt.Fprintf(os.Stderr, "filename  is not empty")
		os.Exit(1)
	}
	fileInfo := make(chan string)
	wg.Add(1)
	go recursionDir(*dir, &wg, fileInfo, *tagFile, *findMod)
	go func() {
		wg.Wait()
		close(fileInfo)
	}()
	var count int64  =0
	loop:
		for{
			select {
			case v, ok := <-fileInfo:
				if !ok {
					break loop
				}
				fmt.Fprintf(os.Stdout, "dir %s \n", v)
			}
			count ++
		}
	if count ==0 {
		fmt.Fprintf(os.Stdout, "not exists %s \n", *tagFile)
	}else {
		fmt.Fprintf(os.Stdout, "count %s %d \n", *tagFile, count)
	}
}

func recursionDir(dir string, wg *sync.WaitGroup, fileinfo chan <-string, targfile, mod string)  {
	defer wg.Done()
	for _, file := range readDir(dir){
		if file.IsDir() {
			wg.Add(1)
			filePath := filepath.Join(dir, file.Name())
			if isMatch(targfile, file.Name(), mod) {
				fileinfo <- filepath.Join(dir, file.Name())
			}
			go recursionDir(filePath, wg, fileinfo, targfile, mod)
		}else {
			if isMatch(targfile, file.Name(), mod) {
				fileinfo <- filepath.Join(dir, file.Name())
			}
		}
	}
}

func isMatch(regx, filename, mod string)bool  {
	var reg bool
	switch mod {
	case "exact":
		if regx == filename {
			return true
		}
		return false
	case "suffix":
		patten := fmt.Sprintf(`^.*%s$`, regx)
		reg, _ = regexp.Match(patten, []byte(filename))
		return reg
	default:
		return reg
	}

}

//读取文件夹
func readDir(dir string) []os.FileInfo  {
	schem <- struct{}{}
	defer func() {<-schem}()
	files, err := ioutil.ReadDir(dir)
	if err != nil{
		fmt.Fprintf(os.Stderr, "open dir err %s \n", err)
	}
	return files
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

