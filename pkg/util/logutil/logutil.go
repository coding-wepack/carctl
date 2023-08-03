package logutil

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/settings"
)

func WriteGoroutineFile(goroutineCount *int32, execJobNum []int32) {
	file, err := os.Create("goroutine.log")
	if err != nil {
		log.Error("failed to create goroutine.log", logfields.Error(err))
		return
	}
	defer file.Close()
	write := bufio.NewWriter(file)

	writeString(write, "时间\t\t\t协程数\t")
	for i := 0; i < settings.Concurrency; i++ {
		writeString(write, fmt.Sprintf("协程%d\t", i))
	}
	writeString(write, "\n")
	for {
		writeString(write, fmt.Sprintf("%s\t%d\t\t", time.Now().Format("04:05.000"), *goroutineCount))
		for i := 0; i < settings.Concurrency; i++ {
			writeString(write, fmt.Sprintf(" %d\t\t", execJobNum[i]))
		}
		writeString(write, "\n")
		write.Flush()
		time.Sleep(time.Millisecond * 300)
	}
}

func writeString(w *bufio.Writer, str string) {
	_, err := w.WriteString(str)
	if err != nil {
		log.Error("failed to write string to goroutine.log", logfields.Error(err))
	}
}
