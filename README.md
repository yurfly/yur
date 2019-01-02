# yur

一款静态页面生成工具

模板引擎用的是golang html/template包

配置文件采用json格式

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
