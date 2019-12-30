package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/robfig/cron"
)

func main() {
	logPath := "/data/logs"
	taskCron := os.Getenv("cron")
	if taskCron == "" {
		taskCron = "0 * */1 * * ?"
	}
	log.Println("cron  is  ", taskCron)
	Init(logPath, taskCron)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>A Prometheus Exporter</title></head>
			<body>
			<h1>清理K8S应用日志</h1>
			<p>清理方式只将日志文件置空，不会删除日志文件</p>
		    <p>环境变量说明：</p>
			<p>folder_max_size 日志总文件夹大小阈值 默认值 10G</p>
			<p>file_max_size 单个日志文件内容大小阈值 默认值 2G</p>
			<p>cron 定时规则 默认 0 * */1 * * ? 每隔一个小时执行一次</p>
			<p>容器内扫描日志路径为   /data/logs  请将此路径映射到节点上的日志目录</p>
			</body>
			</html>`))
	})
	log.Printf("Starting Server at http://localhost:%d", 80)
	log.Fatal(http.ListenAndServe(":"+fmt.Sprintf("%d", 80), nil))
}

//初始化定时任务
func Init(logPath, taskCron string) {

	c := cron.New()
	//taskCron := mapCron[taskName]
	_ = c.AddFunc(taskCron, func() {
		schedule, _ := cron.Parse(taskCron)

		timeNow := time.Now()                       //当前时间
		nextTime := schedule.Next(timeNow)          //任务下次执行时间
		seconds := nextTime.Unix() - timeNow.Unix() //距离下次执行的秒数
		expire, _ := strconv.Atoi(strconv.FormatInt(seconds, 10))
		expire = expire - 1 //提现一秒释放锁
		if expire <= 0 {    //如果定时任务没有下个时间，key的失效时间默认设置为1分钟之后
			expire = 60
		}
		clean("/data/logs")
	})
	c.Start()
}

func clean(logPath string) {

	log.Printf("定时任务开始执行:  ")

	//文件夹最大阈值
	folder_max_size := os.Getenv("folder_max_size")
	if folder_max_size == "" {
		folder_max_size = "10"
	}
	//文件最大阈值
	file_max_size := os.Getenv("file_max_size")
	if file_max_size == "" {
		file_max_size = "2"
	}

	folderMaxSize, err := strconv.ParseFloat(folder_max_size, 64)
	if err != nil {
		folderMaxSize = 10
	}

	fileMaxSize, err := strconv.ParseFloat(file_max_size, 64)
	if err != nil {
		fileMaxSize = 2
	}

	folderMaxSize *= 1073741824
	fileMaxSize *= 1073741824

	log.Printf("folderMaxSize %s ", UnitAndSizeWithKb(int64(folderMaxSize)))

	log.Printf("fileMaxSize %s ", UnitAndSizeWithKb(int64(fileMaxSize)))

	log.Printf("container logs path is %s", logPath)

	// log.Println("EmptyFileBy folder max size  begin:")
	//文件夹大小
	folderTotalSize, err := GetFolderSize(logPath)
	if err == nil {
		log.Printf("%s size is %s ", logPath, UnitAndSizeWithKb(folderTotalSize))

		if folderTotalSize > int64(folderMaxSize) {
			//执行清理
			err = EmptyLogFile(logPath)
			if err != nil {
				log.Printf("EmptyLogFile by folderMaxSize fail,cause %s", err.Error())
			}
		}

	} else {
		log.Printf("GetFolderSize fail, %s", err.Error())
	}

	//log.Println("EmptyFileByFileMaxSize begin:")
	//文件清理
	err = EmptyFileByFileMaxSize(logPath, int64(fileMaxSize))
	if err != nil {
		log.Printf("EmptyFileByFileMaxSize fail ,%s", err.Error())
	}

	log.Printf("定时任务执行结束。 ")
}

func EmptyFileByFileMaxSize(pathname string, maxSize int64) error {
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			log.Printf("Foder is [%s]\n", pathname+string(os.PathSeparator)+fi.Name())
			_ = EmptyFileByFileMaxSize(pathname+string(os.PathSeparator)+fi.Name(), maxSize)
		} else {
			if fi.Size() > maxSize {
				log.Printf("%s will empty , length is %s ", pathname+string(os.PathSeparator)+fi.Name(), UnitAndSizeWithKb(fi.Size()))
				command := ":> " + pathname + string(os.PathSeparator) + fi.Name()

				log.Printf("command is %s", command)
				cmd := exec.Command("sh", "-c", command)

				bytes, err := cmd.Output()
				if err != nil {
					log.Println(err)
				} else {
					resp := string(bytes)
					log.Println(resp)
				}
			} else {
				log.Printf("[%s],size is  %s", pathname+string(os.PathSeparator)+fi.Name(), UnitAndSizeWithKb(fi.Size()))
			}
		}
	}
	return err
}

//计算文件夹的大小
func GetFolderSize(pathname string) (int64, error) {
	rd, err := ioutil.ReadDir(pathname)
	var totalSize int64
	totalSize = 0
	for _, fi := range rd {
		if fi.IsDir() {
			totalSizeTmp, _ := GetFolderSize(pathname + string(os.PathSeparator) + fi.Name())
			totalSize = totalSize + totalSizeTmp
			log.Printf("dir %s total size is  %s", fi.Name(), UnitAndSizeWithKb(totalSizeTmp))

		} else {
			totalSize = totalSize + fi.Size()
			log.Printf("File %s  size is  %s", fi.Name(), UnitAndSizeWithKb(fi.Size()))
		}
	}
	return totalSize, err
}

//全部文件清理
func EmptyLogFile(pathname string) error {
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			log.Printf("Foder is [%s]\n", pathname+string(os.PathSeparator)+fi.Name())
			_ = EmptyLogFile(pathname + string(os.PathSeparator) + fi.Name())
		} else {
			if fi.Size() > 0 {
				log.Printf("%s will empty , size is %s ", pathname+string(os.PathSeparator)+fi.Name(), UnitAndSizeWithKb(fi.Size()))
				command := ":> " + pathname + string(os.PathSeparator) + fi.Name()
				log.Printf("command is %s", command)
				cmd := exec.Command("sh", "-c", command)
				bytes, err := cmd.Output()
				if err != nil {
					log.Println(err)
				} else {
					resp := string(bytes)
					log.Println(resp)
				}

			}
		}
	}
	return err
}

// f:需要处理的浮点数，n：要保留小数的位数
// Pow10（）返回10的n次方，最后一位四舍五入，对ｎ＋１位加０．５后四舍五入
func round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10

}

func UnitAndSizeWithKb(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d %s", size, "byte")
	}
	size = size / 1024
	if size < (1 << 10) {
		return fmt.Sprintf("%f %s", float64(size), "KB")
	} else if size < (1 << 20) {
		totalCapacity := round(float64(size)/(1<<10), 4)
		return fmt.Sprintf("%f %s", totalCapacity, "MB")
	} else if size < (1 << 30) {
		totalCapacity := round(float64(size)/(1<<20), 4)
		return fmt.Sprintf("%f %s", totalCapacity, "GB")
	} else if size < (1 << 40) {
		totalCapacity := round(float64(size)/(1<<30), 4)
		return fmt.Sprintf("%f %s", totalCapacity, "TB")
	} else if size < (1 << 50) {
		totalCapacity := round(float64(size)/(1<<40), 4)
		return fmt.Sprintf("%f %s", totalCapacity, "PB")
	} else if size < (1 << 60) {
		totalCapacity := round(float64(size)/(1<<50), 4)
		return fmt.Sprintf("%f %s", totalCapacity, "EB")
	}
	return fmt.Sprintf("%f %s", float64(size), "KB")
}
