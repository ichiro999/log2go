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
