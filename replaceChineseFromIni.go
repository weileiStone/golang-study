package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type bytePos struct {
	Line    int
	StartP  int
	EndP    int
	content []byte
}

func main() {
	cmpSourceIni := flag.String("source_ini", "", "cmp ini file")
	cmpTargetPhp := flag.String("target_php", "", "cmp php file")
	cfname := flag.String("cfname", "", "create php file name and dir")
	replaceType := flag.String("replace_type", "", "replace key or value")
	flag.Parse()


	zhIniMap, err := getZhIni(*cmpSourceIni, *replaceType)
	if err != nil {
		fmt.Printf(`ReadFile err cmpSourceIni %s %s`, *cmpSourceIni, err)
		os.Exit(1)
	}

	file, err := os.Open(*cmpTargetPhp) // 打开文件
	if err != nil {
		fmt.Printf(`ReadFile cmpTargetPhp  %s err %s`, *cmpTargetPhp, err)
		return
	}
	defer file.Close()
	ranges := []int{0}
	oribytes := make([]byte, 0)
	lastEnd := 0
	for _, pos := range zhIniMap {
		contentBytes := make([]byte, pos.EndP-pos.StartP)
		_, err := file.ReadAt(contentBytes, int64(pos.StartP))
		if err != nil {
			fmt.Printf(`ReadFile at err %s`, err)
			break
		}
		fmt.Printf("content %s \r\n", string(contentBytes))
		fmt.Printf("last end %d new start %d \r\n", lastEnd, pos.StartP)

		oriCont := make([]byte, pos.StartP-lastEnd)
		_, err = file.ReadAt(oriCont, int64(lastEnd))
		if err != nil {
			fmt.Printf(`ReadFile at err %s`, err)
			break
		}

		oriCont = append(oriCont, pos.content...)
		oribytes = append(oribytes, oriCont...)
		ranges = append(ranges, pos.StartP, pos.EndP)
		lastEnd = pos.EndP
	}
	// fmt.Printf("range %+v \r\n", ranges)
	//获取文件大小
	fileStat, err := os.Stat(*cmpTargetPhp)
	if err != nil {
		fmt.Printf(`fileStat err %s`, err)
		return
	}

	overContent := make([]byte, fileStat.Size()-int64(lastEnd))
	_, err = file.ReadAt(overContent, int64(lastEnd))
	if err != nil {
		fmt.Printf(`read over err %s`, err)
		return
	}
	oribytes = append(oribytes, overContent...)
	err = createFile(*cfname, oribytes)
	if err != nil {
		fmt.Printf(`write err %s`, err)
		return
	}
}

func createFile(path string, content []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	_, err = writer.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func getZhIni(path string, replaceType string) ([]bytePos, error) {

	zhIniContent := make([]bytePos, 0)
	file, err := os.Open(path) // 打开文件
	if err != nil {
		fmt.Printf(`ReadFile err %s`, err)
		return zhIniContent, err
	}
	defer file.Close()
	// 确保在函数结束时关闭文件
	scanner := bufio.NewScanner(file) // 创建新的Scanner对象
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		pos := bytePos{}
		oriContent := scanner.Text()
		oriSlice := strings.Split(oriContent, "=")
		orikey := oriSlice[0]
		oriValue := oriSlice[1]
		oriValue = strings.TrimRight(strings.TrimLeft(oriValue, `"`), ` "`)
		oriValue = strings.Replace(oriValue, `'`, `\'`, -1)
		switch replaceType {
		case `key`:
			pos.content = []byte(`{` + orikey + `}`)
		default:
			pos.content = []byte(oriValue)
		}

		intKeySlice := strings.Split(orikey, "_")
		lineKey := intKeySlice[1]
		lineint, _ := strconv.Atoi(lineKey)

		startKey := intKeySlice[2]
		startint, _ := strconv.Atoi(startKey)
		endKey := intKeySlice[3]
		endint, _ := strconv.Atoi(endKey)
		pos.Line = lineint
		pos.StartP = startint
		pos.EndP = endint
		zhIniContent = append(zhIniContent, pos)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf(`ReadFile err %s %s`, path, err)
		return zhIniContent, err
	}
	return zhIniContent, nil
}
