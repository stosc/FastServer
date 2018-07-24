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
	m          map[string]string
)

//初始化
func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&keyValue, "key", "", "http key")
	flag.StringVar(&Upload_Dir, "path", "./upload/", "upload path")
	initFileTypes()
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/{fid:[0-9a-f]+}", FileHandler)
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
	f, err := filepath.Abs(Upload_Dir)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Server is running，key is ", keyValue, ",upload path is ", f)
	if Upload_Dir == "" {
		Upload_Dir = "./upload/"
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["fid"]
	//ext := vars["ext"]
	//fmt.Println(file, ".", ext)
	path := buildFilePath(file)
	//fmt.Println(path)
	filenaem := path + file //+ "." + ext
	f, _ := ioutil.ReadFile(filenaem)
	ext := GetFileType(f)

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
		fmt.Println(1)
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "%v", "上传错误:"+err.Error())
			return
		}
		fd, err := ioutil.ReadAll(file)
		fileext := GetFileType(fd)
		fmt.Println("文件类型：.", fileext)
		if check(fileext) == false {
			fmt.Fprintf(w, "%v", "不允许的上传类型")
			return
		}

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
		filename := buildFilePath(m5) + m5
		//fmt.Println("文件内容：", filename)
		err = ioutil.WriteFile(filename, fd, 0666)
		if err != nil {
			fmt.Fprintf(w, "%v", "上传失败:"+err.Error())
			return
		}
		fmt.Fprintf(w, "%v", m5)
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

func initFileTypes() {
	m = make(map[string]string)
	m["ffd8ffe000104a464946"] = "jpg"
	m["89504e470d0a1a0a0000"] = "png"
	m["47494638396126026f01"] = "gif"
	m["49492a00227105008037"] = "tif"
	m["424d228c010000000000"] = "bmp"
	m["424d8240090000000000"] = "bmp"
	m["424d8e1b030000000000"] = "bmp"
	m["41433130313500000000"] = "dwg"
	m["3c21444f435459504520"] = "html"
	m["3c21646f637479706520"] = "htm"
	m["48544d4c207b0d0a0942"] = "css"
	m["696b2e71623d696b2e71"] = "js"
	m["7b5c727466315c616e73"] = "rtf"
	m["38425053000100000000"] = "psd"
	m["46726f6d3a203d3f6762"] = "eml"
	m["d0cf11e0a1b11ae10000"] = "doc"
	m["d0cf11e0a1b11ae10000"] = "vsd"
	m["5374616E64617264204A"] = "mdb"
	m["252150532D41646F6265"] = "ps"
	m["255044462d312e350d0a"] = "pdf"
	m["2e524d46000000120001"] = "rmvb"
	m["464c5601050000000900"] = "flv"
	m["00000020667479706d70"] = "mp4"
	m["49443303000000002176"] = "mp3"
	m["000001ba210001000180"] = "mpg"
	m["3026b2758e66cf11a6d9"] = "wmv"
	m["52494646e27807005741"] = "wav"
	m["52494646d07d60074156"] = "avi"
	m["4d546864000000060001"] = "mid"
	m["504b0304140000000800"] = "zip"
	m["526172211a0700cf9073"] = "rar"
	m["235468697320636f6e66"] = "ini"
	m["504b03040a0000000000"] = "jar"
	m["4d5a9000030000000400"] = "exe"
	m["3c25402070616765206c"] = "jsp"
	m["4d616e69666573742d56"] = "mf"
	m["3c3f786d6c2076657273"] = "xml"
	m["494e5345525420494e54"] = "sql"
	m["7061636b616765207765"] = "java"
	m["406563686f206f66660d"] = "bat"
	m["1f8b0800000000000000"] = "gz"
	m["6c6f67346a2e726f6f74"] = "properties"
	m["cafebabe0000002e0041"] = "class"
	m["49545346030000006000"] = "chm"
	m["04000000010000001300"] = "mxp"
	m["504b0304140006000800"] = "docx"
	m["d0cf11e0a1b11ae10000"] = "wps"
	m["6431303a637265617465"] = "torrent"
	m["6D6F6F76"] = "mov"
	m["FF575043"] = "wpd"
	m["CFAD12FEC5FD746F"] = "dbx"
	m["2142444E"] = "pst"
	m["AC9EBD8F"] = "qdf"
	m["E3828596"] = "pwl"
	m["2E7261FD"] = "ram"
	m["FFD8FF"] = "jpg"
	m["89504E47"] = "png"
	m["47494638"] = "gif"
	m["49492A00"] = "tif"
	m["424D"] = "bmp"
	m["41433130"] = "dwg"
	m["38425053"] = "psd"
	m["7B5C727466"] = "rtf"
	m["3C3F786D6C"] = "xml"
	m["68746D6C3E"] = "html"
	m["44656C69766572792D646174653A"] = "eml"
	m["D0CF11E0"] = "doc"
	m["5374616E64617264204A"] = "mdb"
	m["252150532D41646F6265"] = "ps"
	m["255044462D312E"] = "pdf"
	m["504B0304"] = "zip"
	m["52617221"] = "rar"
	m["57415645"] = "wav"
	m["41564920"] = "avi"
	m["2E524D46"] = "rm"
	m["000001BA"] = "mpg"
	m["000001B3"] = "mpg"
	m["6D6F6F76"] = "mov"
	m["3026B2758E66CF11"] = "asf"
	m["4D546864"] = "mid"
	m["1F8B08"] = "gz"
}

func GetFileType(fileContent []byte) string {
	if len(fileContent) < 4 {
		return ""
	}
	ext := searchFileType(fileContent, 10)
	if ext != "" {
		return ext
	}
	ext = searchFileType(fileContent, 8)
	if ext != "" {
		return ext
	}
	ext = searchFileType(fileContent, 4)
	if ext != "" {
		return ext
	}
	ext = searchFileType(fileContent, 3)
	if ext != "" {
		return ext
	}
	ext = searchFileType(fileContent, 2)
	if ext != "" {
		return ext
	}
	return ""
}

func searchFileType(fileContent []byte, length int) string {
	fileHeader := fileContent[:length]
	fh := hex.EncodeToString(fileHeader)
	//fmt.Println("文件头:", fh)
	ext := m[fh]
	return ext
}
