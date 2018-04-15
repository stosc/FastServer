package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var mux map[string]func(http.ResponseWriter, *http.Request)

type Myhandler struct{}
type home struct {
	Title string
}

const (
	Template_Dir = "./view/"
	Upload_Dir   = "./upload/"
)

var (
	keyValue string
	h        bool
)

//初始化
func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&keyValue, "key", "", "a string")
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}
	server := http.Server{
		Addr:        ":8899",
		Handler:     &Myhandler{},
		ReadTimeout: 10 * time.Second,
	}
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = index
	mux["/upload"] = upload
	mux["/file"] = StaticServer
	if !isDirExists(Upload_Dir) {
		os.Mkdir(Upload_Dir, os.ModePerm)
	}
	fmt.Println("Server is running...")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (*Myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}
	if ok, _ := regexp.MatchString("/css/", r.URL.String()); ok {
		http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))).ServeHTTP(w, r)
	} else {
		http.StripPrefix("/", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
	}

}

//判断文件夹是否存在
func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
	panic("not reached")
}

//上传文件处理
func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles(Template_Dir + "file.html")
		t.Execute(w, "上传文件")
	} else {
		r.ParseMultipartForm(32 << 20)
		key := r.FormValue("key")
		fmt.Println("key is ", key)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Fprintf(w, "%v", "上传错误")
			return
		}

		fileext := filepath.Ext(handler.Filename)
		if check(fileext) == false {
			fmt.Fprintf(w, "%v", "不允许的上传类型")
			return
		}

		fd, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Fprintf(w, "%v", "md5 read"+err.Error())
			return
		}
		//计算md5
		md5hash := md5.New()
		if _, err := io.Copy(md5hash, file); err != nil {
			fmt.Fprintf(w, "%v", "md5"+err.Error())
			return
		}
		var m5 = hex.EncodeToString(md5hash.Sum([]byte("")))

		//计算md5 End
		filename := buildFilePath(m5) + m5 + fileext
		//filename := m5 + fileext
		fmt.Println("文件内容：", filename)
		err = ioutil.WriteFile(filename, fd, 0666)
		if err != nil {
			fmt.Fprintf(w, "%v", "上传失败")
			return
		}
		fmt.Fprintf(w, "%v", filename)
	}
}

//把md5值的1-3位转换为十六进制数后除以4，作为第一级子目录；4-6位同样处理，作为第二级子目录；
//二级子目录下是以md5命名的文件夹
func buildFilePath(md5Value string) string {
	dir1 := hexToDirName(Substr(md5Value, 0, 3))
	if !isDirExists(Upload_Dir + dir1) {
		os.Mkdir(Upload_Dir+dir1, os.ModePerm)
	}
	dir2 := hexToDirName(Substr(md5Value, 3, 3))
	if !isDirExists(Upload_Dir + dir1 + dir2) {
		os.Mkdir(Upload_Dir+dir1+dir2, os.ModePerm)
	}
	dir := Upload_Dir + dir1 + dir2 + md5Value + "/"
	os.Mkdir(dir, os.ModePerm)
	fmt.Println(dir)
	return dir
}

func hexToDirName(h string) string {
	s, _ := strconv.ParseInt(h, 16, 32)
	p := s / 4
	return strconv.FormatInt(p, 10) + "/"
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func index(w http.ResponseWriter, r *http.Request) {
	title := home{Title: "首页"}
	t, _ := template.ParseFiles(Template_Dir + "index.html")
	t.Execute(w, title)
}

func StaticServer(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/file", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
}

func check(name string) bool {
	ext := []string{".exe", ".js", ".png"}

	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}
