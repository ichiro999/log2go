log2go
======
关于

log2go是一个简单的日志系统，能够输出格式化的日志信息并能按级别进行记录，
同时提供了按文件大小和按时间来进行日志的备份功能

特性

1, 日志可以按照DEBUG, INFO, WARN, ERROR以及FATL五种级别进行记录，并按级
   别的高低进行记录
2, 提供了按时间和按文件大小两种方式进行回滚
3, 日志记录提供了日期，时间，文件名，函数名以及行号等额外信息
4, 具有线性安全，能够提供多个goroutine对其进行记录
5, 可以将日志写到文件中，也可以将日志写到标准错误输出


API说明
func New(log_file_name string, time_interval int64, log_file_max_size uint64, log_level int) (Log2goer, error)
返回一个Log2goer接口类型值和error值，若error为nil，则表示成功
arg：
log_file_name： 为日志文件名，若为空值，则将日志输出到stderr中
time_interval： 回滚的时间间隔，单位为分钟
log_file_max_size： 回滚的文件最大值，只有在time_interval为0时，才会起作用
log_level: 日志记录的最低级别，记录的级别从小到大依次为EBUG, INFO, WARN, ERROR, FATAL

各种级别的记录日志接口，其中format才用fmt.Printf中一样的格式化字符串
func (l *logObject) LOG_DEBUG(format string, v...interface{})

func (l *logObject) LOG_INFO(format string, v...interface{})

func (l *logObject) LOG_WARN(format string, v...interface{})

func (l *logObject) LOG_ERROR(format string, v...interface{})

func (l *logObject) LOG_FATAL(format string, v...interface{})

示例

package main
import "log2go"
import "time"
import "fmt"

func main() {
    l,err := log2go.New("", 0, 100, log2go.WARN)
    if err != nil {
        fmt.Println(err)
    }
    for i := 1; i < 10; i++ {
        l.LOG_DEBUG("test %d", i)
        l.LOG_WARN("test %d", i)
        l.LOG_ERROR("test %d", i)
    }
    fmt.Println("main over")
    time.Sleep(30 * time.Second)
}

