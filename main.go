package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

/**
定时器
用途：
	通知我多久之后访问某个地址，发送哪些数据
接口：
	接收通知请求
		参数：
			url：要求被请求的地址
			interval：多久之后被请求
			params：需要携带的参数，json字符串
*/

// 定时器结构体
type Task struct {
	Url      string        `json:"url" binding:"required"`
	Interval time.Duration `json:"interval" binding:"required"`
	Params   string        `json:"params"`
}

func main() {
	port := "12345"
	if len(os.Args) > 1 {
		// 自定义端口处理
		_, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("输入的端口号不合法，使用默认端口号：", port)
		} else {
			port = os.Args[1]
		}
	}

	// 启动服务
	r := gin.Default()

	r.POST("/timer", func(c *gin.Context) {
		var task Task
		if err := c.ShouldBindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": 400,
				"error":  err.Error(),
			})
			return
		}
		//fmt.Println(time.Now().Unix())
		log.Println(time.Now().Unix(), " - get request: ", task)
		go process(task)
		c.JSON(http.StatusOK, gin.H{
			"status":   200,
			"url":      task.Url,
			"params":   task.Params,
			"interval": task.Interval,
		})
	})

	r.Run(":" + port)
}

// 初始化后日志
func init() {
	file, err := os.OpenFile(`./debug.log`, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	// 设置存储位置
	log.SetOutput(file)
}

// 协程任务
func process(task Task) {
	timer := time.NewTimer(task.Interval * time.Second)
	select {
	case <-timer.C:
		//fmt.Println(time.Now().Unix())
		//fmt.Println(task.Params)

		client := http.Client{}
		request, _ := http.NewRequest("POST", task.Url, strings.NewReader(task.Params))
		request.Header.Set("Content-type", "application/json")
		do, err := client.Do(request)
		log.Println(time.Now().Unix(), " - request: ", task)
		if err != nil {
			log.Println("请求" + task.Url + "失败，参数：" + task.Params)
		}
		defer do.Body.Close()

		//fmt.Println("我请求了地址" + task.Url)
		all, _ := ioutil.ReadAll(do.Body)
		//fmt.Println(string(all))
		log.Println(time.Now().Unix(), " - response: ", string(all))
	}
}
