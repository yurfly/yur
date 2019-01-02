package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func printUsage() {
	fmt.Println(`
Usage:

yur command [args...]

	初始化博客文件夹
	yur init

	新建 markdown 文件
	yur new [filename]

	编译博客
	yur build

	打开本地服务器 [http://localhost]
	yur http
	
	打开本地服务器 [http://localhost:8080]
	yur http :8080

`)
}

var (
	htmldocs   string = "htmldocs"
	publicPath string = "public"
	meta       string = `
	{{ define "meta" }}
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8;width=device-width, initial-scale=1" />
    <title>hello word</title>
    <link href="css/style.css" rel="stylesheet">
	{{ end }}
	`
	index string = `
	<!doctype html>
	<html>
	<head>
	{{template "meta"}}
	</head>
	<body>
	hello word !
	</body>
	</html>
	`
	jsonConfig string = `{
            "menu111": [{
			"title":"url1",
			"url":"#",
			"submenu":{}
		}],
		            "menu222": [{
			"title":"url2",
			"url":"#",
			"submenu":{}
		}]
    }`
	css string = `body {
		color:red;
		fontsize:36px;
	}`
)

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func initProject() {
	if IsExist(htmldocs) {
		fmt.Println("项目文件", htmldocs, "已经存在, 无法init")
		return
	}
	err := os.MkdirAll(htmldocs, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(publicPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	commonPath := path.Join(htmldocs, "common")
	err = os.MkdirAll(commonPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	//创建静态资源文件夹以及常用的子文件夹
	staticPath := path.Join(htmldocs, "static")
	err = os.MkdirAll(staticPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	cssPath := path.Join(staticPath, "css")
	err = os.MkdirAll(cssPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	jsPath := path.Join(staticPath, "js")
	err = os.MkdirAll(jsPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	imagesPath := path.Join(staticPath, "images")
	err = os.MkdirAll(imagesPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	//创建默认demo文件
	jsonPath := path.Join(htmldocs, "data.json")
	createFile(jsonPath, jsonConfig)
	indexname := path.Join(htmldocs, "index.html")
	createFile(indexname, index)
	cssname := path.Join(cssPath, "style.css")
	createFile(cssname, css)
	metaPath := path.Join(commonPath, "meta.html")
	createFile(metaPath, meta)
}

func createFile(filename string, content string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = file.Write([]byte(content))
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
}

func httpstart(addr string) {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(publicPath))))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln(err)
	}
}

func build() {
	var data map[string]interface{}
	jsonPath := path.Join(htmldocs, "data.json")
	datajson, err := os.Open(jsonPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer datajson.Close()
	jsonRead, err := ioutil.ReadAll(datajson)
	jsonStr := string(jsonRead)
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("JsonToMapDemo err: ", err)
	}
	fmt.Println("data.json:\n", data)

	//获取所有的页面模板,htmldocs/根目录下每个html文件都作为一个页面
	pagefiles, err := filepath.Glob(path.Join(htmldocs, "*.html"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range pagefiles {
		fmt.Println(".......扫描到page文件.....", f)
	}
	//获取所有的公共模板,htmldocs/common/根目录下每个html文件都作为公共部分
	commonfiles, err := filepath.Glob(path.Join(htmldocs, "common", "*.html"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range commonfiles {
		fmt.Println(".......扫描到common文件.....", f)
	}

	pageTmpls := make(map[string]*template.Template)
	for _, pageFile := range pagefiles {
		files := append(commonfiles, pageFile)
		pageTmpls[filepath.Base(pageFile)] = template.Must(template.ParseFiles(files...))
	}

	err = os.MkdirAll(publicPath, os.ModePerm)
	if err != nil {
		println("\t创建文件夹[", publicPath, "]失败")
	}

	for tmplName, tmpl := range pageTmpls {
		fmt.Println(".......................创建文件..........................", tmplName)
		filename := path.Join(publicPath, tmplName)
		tmplFile, err := os.Create(filename)
		if err != nil {
			log.Fatalln("创建文件" + tmplName + "失败")
		}
		fmt.Println(".......", tmpl.Name())
		tmpls := tmpl.Templates()
		for _, t := range tmpls {
			fmt.Println(".......\t", tmpl.Name(), " ---> ", t.Name())
		}
		tmpl.ExecuteTemplate(tmplFile, tmplName, data)
	}

	//复制静态资源
	staticPath := path.Join(htmldocs, "static")
	ScanDir(staticPath, func(fullpath string, isDir bool) bool {
		//计算新的dst路径
		dstPath := strings.Replace(fullpath, staticPath, publicPath, -1)
		if isDir {
			fmt.Printf("MkdirAll [%s]\n", dstPath)
			err = os.MkdirAll(dstPath, os.ModePerm)
			if err != nil {
				log.Fatalln(err)
				return false
			}
		} else {
			src, err := os.Open(fullpath)
			if err != nil {
				log.Fatalln(err)
				return false
			}
			defer src.Close()
			dst, err := os.Create(dstPath)
			if err != nil {
				log.Fatalln(err)
				return false
			}
			defer dst.Close()
			fmt.Printf("[%s] copy to [%s]\n", fullpath, dstPath)
			io.Copy(dst, src)
		}
		return true
	})
	fmt.Println("...........................end..........................")
}

func ScanDir(dir string, scancb func(fullpath string, isDir bool) bool) error {
	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("[", dir, "]", err)
		return err
	}
	for _, item := range rd {
		fullname := path.Join(dir, item.Name())
		if item.IsDir() {
			fmt.Printf("[%s]\n", fullname)
			scancb(fullname, true)
			ScanDir(fullname, scancb)
		} else {
			fmt.Printf("[%s]\n", item.Name())
			scancb(fullname, false)
		}
	}
	return err
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || len(args) > 3 {
		printUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "init":
		// 初始化项目结构
		initProject()
	case "new":
		//新建 MarkDown
		if len(args) != 2 {
			printUsage()
			os.Exit(1)
		}
		//name := args[1]
		//createMarkdown(name)
	case "build":
		// 编译成静态html网站
		build()
	case "http":
		if len(args) > 2 {
			printUsage()
			os.Exit(1)
		}
		addr := ":80"
		if len(args) == 2 {
			addr = args[1]
		}
		// 运行本地服务器
		fmt.Printf("http service: [%s]\n", addr)
		httpstart(addr)
	case "help":
		printUsage()
	default:
		printUsage()
	}
}
