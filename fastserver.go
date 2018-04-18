package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Myhandler struct{}
type home struct {
	Title string
}

const (
	Template_Dir = "./view/"
)

var (
	keyValue   string
	Upload_Dir string
	h          bool
)

//初始化
func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&keyValue, "key", "", "http key")
	flag.StringVar(&Upload_Dir, "path", "./upload/", "upload path")
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/{fid:[0-9a-f]+}/{ext:[a-zA-Z]+}", FileHandler)
	r.HandleFunc("/upload", upload)
	r.HandleFunc("/file", StaticServer)
	server := http.Server{
		Addr:        ":8899",
		Handler:     r,
		ReadTimeout: 10 * time.Second,
	}
	if !isDirExists(Upload_Dir) {
		os.Mkdir(Upload_Dir, os.ModePerm)
	}
	f, _ := filepath.Abs(Upload_Dir)
	fmt.Println("Server is running，key is ", keyValue, ",upload path is ", f)
	if Upload_Dir == "" {
		Upload_Dir = "./upload/"
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["fid"]
	ext := vars["ext"]
	//fmt.Println(file, ".", ext)
	path := buildFilePath(file)
	//fmt.Println(path)
	filenaem := path + file + "." + ext
	f, _ := ioutil.ReadFile(filenaem)
	if !isImage(ext) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file+"."+ext))
	}
	w.Write(f)
}

func (*Myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		if keyValue != "" {
			key := r.FormValue("key")
			if key != keyValue {
				fmt.Fprintf(w, "%v", "没有操作权限")
				return
			}
		}

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
		if _, err := md5hash.Write(fd); err != nil {
			fmt.Fprintf(w, "%v", "md5"+err.Error())
			return
		}
		var m5 = hex.EncodeToString(md5hash.Sum([]byte("")))
		//计算md5 End
		filename := buildFilePath(m5) + m5 + fileext
		//fmt.Println("文件内容：", filename)
		err = ioutil.WriteFile(filename, fd, 0666)
		if err != nil {
			fmt.Fprintf(w, "%v", "上传失败")
			return
		}
		fmt.Fprintf(w, "%v", m5+fileext)
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
	//fmt.Println(dir)
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
	//fmt.Println(name)
	ext := []string{".exe", ".js"}
	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}

func isImage(name string) bool {
	//fmt.Println(name)
	ext := []string{"jpg", "png"}
	for _, v := range ext {
		if v == name {
			return true
		}
	}
	return false
}
