package main

import (
	"fmt"
	"strconv"
	"net/http"
	"os"
	"regexp"
	"strings"
)

//存储内容和标题
type TitleContent struct {
	Title string
	Content string
}

func main(){
	var start, end int
	fmt.Println("请输入起始页（>=1）")
	fmt.Scan(&start)

	fmt.Println("请输入终止页（>起始页）")
	fmt.Scan(&end)

	DoWork(start, end)
}


func DoWork(start, end int){
	fmt.Printf("准备爬取%d 到 %d页数据", start, end)
	page := make(chan int)
	for i := start; i <= end; i++ {
		go SpiderPape(i, page)
	}

	for i := start; i <= end; i++{
		fmt.Printf("第%d页数据爬取完毕", i)
		<-page
	}
}

//读取数据到管道
func SpiderPape(i int, page chan<- int){
	url := "http://www.81uav.cn/uav-news/?page=" + strconv.Itoa(i)
	fmt.Println("爬取网页", url)

	//开始爬取玩野
	result, err := HttpGet(url)
	if err != nil{
		fmt.Println("HttpGet err ", err)
		return
	}

	//读取内容
	re := regexp.MustCompile(`<h5><a target="blank" href="(?s:(.*?))"`)
	if re == nil{
		fmt.Println("regexp.MustCompile err")
		return
	}

	fileTitle := make(map[int]TitleContent)

	//取数据
	joyUrls := re.FindAllStringSubmatch(result, -1)

	//去网址
	n := 0
	for _, data := range joyUrls {

		title, content, err := SpiderOneJoy(data[1])
		if err != nil {
			fmt.Println("SpiderOneJoy err = ", err)
			continue
		}

		println(title, content)

		fileTitle[n] = TitleContent{
			title, content,
		}
		n++
	}

	//把内容写到文件中
	StoreJoyTOFile(fileTitle, i)
	//向管道中写数据
	page <- i
}

//将数据写到文件中
func StoreJoyTOFile(fileTitle map[int]TitleContent, i int){
	//新建文件
	f, err := os.Create(strconv.Itoa(i) + ".txt")
	if err != nil {
		fmt.Println("os.Create err ", err)
		return
	}
	defer f.Close()

	//写到文件
	for _, file := range fileTitle{
		f.WriteString("title: " + file.Title + "\n")
		f.WriteString("content: " + file.Content + "\n")
		f.WriteString("===================================\n")
	}
}

func SpiderOneJoy(url string)(title string, content string, err error){
	result, err := HttpGet(url)
	if err != nil{
		fmt.Println("HttpGet err ", err)
		return
	}

	//取内容
	re := regexp.MustCompile(`<h1>(?s:(.*?))<\/h1>`)
	if re == nil {
		fmt.Println("regexp.MustCompile err ", err)
		return
	}

	//取数据
	temTitle := re.FindAllStringSubmatch(result, -1)
	for _, data := range temTitle{
		title = data[1]
		title = strings.Replace(title, "\t", "", -1)
		break
	}

	//取内容
	//re2 := regexp.MustCompile(`id="article"><p>(?s:(.*?))<a id="prev" href="`)
	re2 := regexp.MustCompile(`id="article">(?s:(.*?))</div>
</div>
<div class="b10 c_b">`)
	if re2 == nil {
		fmt.Println("re2 regexp.MustCompile err ", err)
		return
	}

	//去睡
	temContent := re2.FindAllStringSubmatch(result, -1)
	for _, data := range temContent{
		TempContent := data[1]
		content = strings.Replace(TempContent, "\t", "", -1)
		break
	}

	return
}

func HttpGet(url string)(result string, err error){
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http.Get err")
		return
	}
	defer resp.Body.Close()

	//开始读取网页内容
	buf := make([]byte, 4 * 1024)
	var tmp string
	for{
		n, err := resp.Body.Read(buf)
		if n == 0{
			fmt.Println("resp.Body.Read = ", err)
			break
		}

		tmp += string(buf)
	}

	result = tmp
	return
}


