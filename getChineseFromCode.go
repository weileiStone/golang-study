package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
)

var paths = make([]string, 0)

var lines = make([]string, 0)

var iniContent = make([]string, 0)
var cfname *string

func main() {

	currtPath, _ := GetCurrentPath()
	dir := flag.String("dir", currtPath, "find dir")
	extStr := flag.String("ext", "php", "file ext name")
	cfname = flag.String("cfname", "test", "create file name")
	flag.Parse()
	readFilesInDirectory(*dir)
	for _, item := range paths {
		ReadFileAll(item, *extStr)
	}
	// file, err := os.Create(`chinaNew.txt`)
	// if err != nil {
	// 	fmt.Fprintf(os.Stdout, `err %s \n`, err)
	// }
	// defer file.Close()
	// writer := bufio.NewWriter(file)
	// defer writer.Flush()
	// for _, item := range lines {
	// 	_, err := writer.WriteString(item + "\n")
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stdout, `err %s \n`, err)
	// 		break
	// 	}
	//	fmt.Fprintf(os.Stdout, "count %s \n", item)
	// }
	WriteFileByLine(`chinaNew-`+*cfname+`.txt`, lines)
	WriteFileByLine("zh-CN-"+*cfname+`.ini`, iniContent)
}

func WriteFileByLine(path string, contentLines []string) {
	if FileIsExist(path) {
		err := os.Remove(path)
		if err != nil {
			fmt.Fprintf(os.Stdout, `err %s \n`, err)
			return
		}
	}
	file, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stdout, `err %s \n`, err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for _, item := range contentLines {
		_, err := writer.WriteString(item + "\n")
		if err != nil {
			fmt.Fprintf(os.Stdout, `err %s \n`, err)
			break
		}
		fmt.Fprintf(os.Stdout, "count %s \n", item)
	}
}

func readFilesInDirectory(directory string) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	// 遍历文件和子目录
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", directory, file.Name())
		if file.IsDir() {
			readFilesInDirectory(filePath) // 递归调用
		} else {
			// 处理文件
			// fmt.Println(filePath)
			// if file.Name() != *exceptFile {
			paths = append(paths, filePath)
			// }
		}
	}
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
		return "", fmt.Errorf(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

func fileAppend(path string, content string) {

	// 打开文件，如果文件不存在则创建
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 追加内容
	_, err = file.WriteString(content + "\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	// fmt.Println("Content appended successfully!")
}

type LineStruct struct {
	PathName string `json:"path_name"`
	Lines    []int  `json:"lines"`
}

func (l *LineStruct) String() string {
	str := `file name ` + l.PathName + `: `
	for _, item := range l.Lines {
		lineStr := strconv.Itoa(item)
		str += str + lineStr + `,\r\n`
	}
	return str
}

func ReadFileAll(path, ext string) error {
	code, _ := ReadFileOs(path, ext)
	parser := tree_sitter.NewParser()
	defer parser.Close()
	// parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_javascript.Language()))
	// parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_php.LanguagePHPOnly()))
	switch ext {
	case `phponly`:
		parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_php.LanguagePHPOnly()))
	case `php`:
		parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_php.LanguagePHP()))
	case `go`:
		parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_go.Language()))
	case `js`:
		parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_javascript.Language()))
	case `vue`:
		parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_html.Language()))
	}

	tree := parser.Parse(code, nil)
	defer tree.Close()

	root := tree.RootNode()
	treeCurs := root.Walk()
	defer treeCurs.Close()
	NodeWalkOneFile(treeCurs, code, path, ext)
	return nil
}

func NodeWalkOneFile(ts *tree_sitter.TreeCursor, sourceVal []byte, path, ext string) {
	defer ts.Close()
	content := ``
	iniLine := ``
	root := ts.Node()
	for _, nd := range root.Children(ts) {
		NodeWalkOneFile(nd.Walk(), sourceVal, path, ext)
	}
	languageMap := map[string]bool{
		`phponly`: (root.KindId() == 363 || root.KindId() == 367),
		`php`:     (root.KindId() == 197 || root.KindId() == 367),
		`go`:      (root.KindId() == 81 || root.KindId() == 83 || root.KindId() == 188 || root.KindId() == 190 || root.KindId() == 80),
		// `js`: root.KindId() == 132 || root.KindId() == 129 || root.KindId() == 51 || root.KindId() == 52 || root.KindId() == 104 ||
		// 	root.KindId() == 105 || root.KindId() == 194 || root.KindId() == 219 || root.KindId() == 220 || root.KindId() == 251 || root.KindId() == 252 ||
		// 	root.KindId() == 254 || root.KindId() == 255 || root.KindId() == 256,
		`js`:  root.KindId() == 132,
		`vue`: root.KindId() == 16,
	}
	isStr := languageMap[ext]

	basePath := filepath.Base(path)
	valval := sourceVal[root.StartByte():root.EndByte()]
	// fmt.Printf("root grammar name %s root Kind %s, kind id %d, grammar id %d, start byte %d, end byte %d, val val %s, row %d \r\n",
	// 	root.GrammarName(), root.Kind(), root.KindId(), root.GrammarId(), root.StartByte(), root.EndByte(), string(valval), root.StartPosition().Row+1)
	if isStr {
		if IsContainChinese(string(valval)) {
			// fmt.Printf("root grammar name %s root Kind %s, kind id %d, grammar id %d, start byte %d, end byte %d, val val %s, row %d \r\n",
			// 	root.GrammarName(), root.Kind(), root.KindId(), root.GrammarId(), root.StartByte(), root.EndByte(), string(valval), root.StartPosition().Row+1)
			content = fmt.Sprintf("file name %s, row %d , chinese content %s, ", path, root.StartPosition().Row+1, string(valval))
			valContent := strings.ReplaceAll(string(valval), `"`, `\"`)
			iniLine = fmt.Sprintf(`%s_%d_%d_%d="%s"`, strings.TrimSuffix(basePath, filepath.Ext(basePath)), root.StartPosition().Row+1, root.StartByte(), root.EndByte(), valContent)

		}

	}
	if content != `` {
		lines = append(lines, content)
	}
	if iniLine != `` {
		iniContent = append(iniContent, iniLine)
	}

}

func NodeWalk(ts *tree_sitter.TreeCursor, sourceVal []byte, path, ext string) {
	defer ts.Close()
	iniLine := ``
	root := ts.Node()
	for _, nd := range root.Children(ts) {
		NodeWalk(nd.Walk(), sourceVal, path, ext)
	}
	languageMap := map[string]bool{
		`php`: (root.KindId() == 363),
		`go`:  root.KindId() == 81 || root.KindId() == 83 || root.KindId() == 189 || root.KindId() == 190 || root.KindId() == 212 || root.KindId() == 213,
		// `js`: root.KindId() == 132 || root.KindId() == 129 || root.KindId() == 51 || root.KindId() == 52 || root.KindId() == 104 ||
		// 	root.KindId() == 105 || root.KindId() == 194 || root.KindId() == 219 || root.KindId() == 220 || root.KindId() == 251 || root.KindId() == 252 ||
		// 	root.KindId() == 254 || root.KindId() == 255 || root.KindId() == 256,
		`js`:  root.KindId() == 132,
		`vue`: root.KindId() == 16,
	}
	isStr := languageMap[ext]
	basePath := filepath.Base(path)
	if isStr {
		valval := sourceVal[root.StartByte():root.EndByte()]
		if IsContainChinese(string(valval)) {
			// fmt.Printf("root grammar name %s root Kind %s, kind id %d, grammar id %d, start byte %d, end byte %d, val val %s, row %d \r\n",
			// 	root.GrammarName(), root.Kind(), root.KindId(), root.GrammarId(), root.StartByte(), root.EndByte(), string(valval), root.StartPosition().Row+1)
			// content = fmt.Sprintf("file name %s, row %d , chinese content %s, ", path, root.StartPosition().Row+1, string(valval))
			iniLine = fmt.Sprintf(`%s_%d_%d_%d="%s"`, strings.TrimSuffix(basePath, filepath.Ext(basePath)), root.StartPosition().Row+1, root.StartByte(), root.EndByte(), string(valval))
			dirName := strings.TrimSuffix(path, basePath)
			dirAllName := dirName + `/` + "zh-CN-" + *cfname + `.ini`
			fileAppend(dirAllName, iniLine)
		}

	}

}

func ReadFileOs(name string, ext string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if filepath.Ext(f.Name()) != `.`+ext {
		return nil, fmt.Errorf(`%s not is %s`, f.Name(), ext)
	}

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}

		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
	}
}

func ReadFile(path, ext string) error {
	line := 1
	file, err := os.Open(path) // 打开文件
	if err != nil {
		fmt.Printf(`ReadFile err %s`, err)
		return err
	}
	defer file.Close() // 确保在函数结束时关闭文件
	if filepath.Ext(file.Name()) != `.`+ext {
		return fmt.Errorf(`%s not is %s`, file.Name(), ext)
	}
	scanner := bufio.NewScanner(file) // 创建新的Scanner对象
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		if IsContainChinese(scanner.Text()) {
			content := fmt.Sprintf(`%s:%d`, path, line)
			lines = append(lines, content)
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf(`ReadFile err %s %s`, path, err)
		return err
	}
	return nil
}

func ReadFileByChuck(path string) (LineStruct, error) {
	line := 0
	lines := LineStruct{
		PathName: path,
		Lines:    make([]int, 0),
	}
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf(`ReadFile err %s`, err)
		return lines, err
	}
	defer file.Close()           // 确保在函数结束时关闭文件
	const chunkSize = 512 * 1024 // 512KB
	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			// 处理读取到的内容
			// fmt.Print(string(buf[:n]))
			content := string(buf[:n])
			for _, r := range content {
				if unicode.Is(unicode.Scripts["Han"], rune(r)) {
					lines.Lines = append(lines.Lines, line)
					continue
				}
			}
		}
		if err == io.EOF {
			break // 文件结束
		}
	}
	return lines, nil
}

func IsContainChinese(str string) bool {
	pattern := "[\u4e00-\u9fa5]" // 匹配中文字符的正则表达式
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		fmt.Printf(`ReadFile err %s`, err)
	}

	return matched
}

func FileIsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}
