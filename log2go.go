package log2go

import (
    "fmt"
    "os"
    "io"
    "strings"
    "errors"
    "runtime"
    "strconv"
    "log"
    "time"
    "sync"
)

type logObject struct {
    //locker *sync.Mutex
    name string
    loger *log.Logger
    loglevel int
    t_interval int64
    max_size uint64
    back_num int64
    closer_f func()
    sendto  chan log_line 
}

type log_line struct {
    level int
    msg string
}

//level of log
const (
    DEBUG = 1 << iota
    INFO 
    WARN
    ERROR
    FATAL
    CHAN_MAX_SIZE = 1024 * 10
)

var g_locker *sync.Mutex

//Log2goer interface
type Log2goer interface {
    LOG_DEBUG(format string, v...interface{})
    LOG_INFO(format string, v...interface{})
    LOG_WARN(format string, v...interface{})
    LOG_ERROR(format string, v...interface{})
    LOG_FATAL(format string, v...interface{})
}

//New create Log2goer pointer
func New(log_file_name string, time_interval int64, log_file_max_size uint64, log_level int) (Log2goer, error){
    if (time_interval < 0 && log_file_max_size < 0) || log_level <= 0 {
        fmt.Fprintf(os.Stderr, "invail arg: %s %d %d %d\n", log_file_name, time_interval, log_file_max_size, log_level)
        return nil, errors.New("invail arg")
    }
    file, closer, err := createfile(log_file_name)
    if err != nil {
        return nil, err
    }

    send_c := make(chan log_line, CHAN_MAX_SIZE)
    l_o := log.New(file, "", (log.Ldate | log.Ltime))
    log_object := logObject{name: log_file_name, loger: l_o, loglevel: log_level, t_interval: time_interval,
        back_num: 0, max_size: log_file_max_size, sendto: send_c, closer_f: closer}
    g_locker = new(sync.Mutex)
    go log_object.doit()
    return &log_object, nil
}

func createfile(file_name string) (io.Writer, func(), error){
    if file_name == "" {
        return os.Stderr, nil, nil  //out to terminal
    }
    var file *os.File = nil
    var errOpen error = nil

    if _, err := os.Stat(file_name); err !=nil {
        //file is not exit
        file, errOpen = os.OpenFile(file_name, os.O_CREATE | os.O_RDWR, 0644)
        if errOpen != nil {
            fmt.Fprintf(os.Stderr, "create log file %s errpr: %s", file_name, err.Error())
            return nil, nil, errOpen
        }
    } else if err == nil {
        file, errOpen = os.OpenFile(file_name, os.O_APPEND | os.O_RDWR, 0644)
        if errOpen != nil {
            fmt.Fprintf(os.Stderr, "open log file %s errpr: %s", file_name, err.Error())
            return nil, nil, errOpen
        }
    }
    closer := func(){file.Close()}
    return file, closer, nil
}


//LOG_DEBUG write log string by level DEBUG
func (l *logObject) LOG_DEBUG(format string, v...interface{}) {
    if strings.HasSuffix(format, "\n") {
        strings.TrimRight(format, "\n")
    }
    l.printtolog(DEBUG, fmt.Sprintf("[DEBUG] " + format + "\n", v...))
}
//LOG_INFO write log string by level INFO
func (l *logObject) LOG_INFO(format string, v...interface{}) {
    if strings.HasSuffix(format, "\n") {
        strings.TrimRight(format, "\n")
    }
    l.printtolog(INFO, fmt.Sprintf("[INFO] " + format + "\n", v...))
}
//LOG_WARN write log string by level WARN
func (l *logObject)LOG_WARN(format string, v...interface{}) {
    l.printtolog(WARN, fmt.Sprintf("[WARN] " + format + "\n", v...))

}

//LOG_WARN write log string by level ERROR
func (l *logObject)LOG_ERROR(format string, v...interface{}) {
    l.printtolog(ERROR, fmt.Sprintf("[ERROR] " + format + "\n", v...))
}

//LOG_FATAL write log string by level ERROR
func (l *logObject)LOG_FATAL(format string, v...interface{}) {
    l.printtolog(FATAL, fmt.Sprintf("[FATAL] " + format + "\n", v...))
}
//printtotlog write to chan withe level 
func (l *logObject) printtolog(level int, msg string) {
    pc, file, line, ok := runtime.Caller(2)
    if !ok {
        file = "???"
        line = 0
    }
    func_name := runtime.FuncForPC(pc).Name()
    msg = changPathStr(file) + ":[" + changPathStr(func_name) + ":" + strconv.Itoa(line) + "] " + msg 
    l.sendto <- log_line{level, msg}
}

//doit write the log string to file or stdout at new goroutine
func (l *logObject) doit() {
    if l.closer_f != nil {
        //make l.closer_f can change by black log file
        defer func() { l.closer_f() }()
    }
    time_ch := time.Tick(time.Duration(l.t_interval))
    if l.t_interval > 0 {
        l.max_size = 0
    }
    var curr_size uint64
    if l.name != "" {
        file_stat, err := os.Stat(l.name)
        if err != nil {
            log.Fatal("get stat of ", l.name, " eror: ", err.Error())
        }
        curr_size = uint64(file_stat.Size())
    }
    for {
        select {
            case log_msg := <- l.sendto: {
                if l.name != "" && l.max_size > 0 && curr_size >= l.max_size {
                    err := back_log_file(l, false)
                    if err != nil {
                        l.loger.Println("back log by size error");
                        continue
                    }
                    curr_size = 0
                }

                if log_msg.level < l.loglevel {
                    continue
                }
                l.loger.Printf("%s", log_msg.msg)
                if l.name != "" && l.max_size > 0 {
                    curr_size += uint64(len([]byte(log_msg.msg)))
                }
            }
            case <-time_ch: {
                if l.name != "" {
                    err := back_log_file(l, true)
                    if err != nil {
                        l.loger.Println("back log by time error");
                        continue
                    }
                }
            }
        }

    }
}

func back_log_file(l_obj *logObject,  use_time bool) error {
    l_obj.closer_f()
    curr_time := time.Now()
    year, moth, date := curr_time.Date()

    time_str := fmt.Sprintf("%d%02d%02d.%d", year, moth, date, l_obj.back_num)
    back_file_name := l_obj.name + "." + time_str 

    g_locker.Lock()
    defer g_locker.Unlock()

    if err := os.Rename(l_obj.name, back_file_name); err != nil {
        return err
    }

    file_writer, new_closer, err := createfile(l_obj.name)
    if err != nil {
        return err
    }
    if !use_time {
        l_obj.back_num++
    } else {
        if l_obj.t_interval >= 60 {
            l_obj.back_num += l_obj.t_interval / 60
        } else {
            l_obj.back_num += l_obj.t_interval
        }
    }

    l_obj.closer_f = new_closer
    l_obj.loger = log.New(file_writer, "", (log.Ldate | log.Ltime))
    return nil
}

func changPathStr(path string) string {
    short := path
    for i := len(path) - 1; i > 0; i-- {
        if path[i] == '/' {
            short = path[i+1:]
            break
        }
    }
    return short
}
