package types

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/logutil"
	"github.com/coding-wepack/carctl/pkg/util/queueutil"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

var (
	ErrForEachContinue = errors.New("continue")
)

type (
	Repository struct {
		// Path is image path to repository
		Path string `json:"path"`

		// Count is count of image of the repository
		Count int `json:"-"`

		// Images is image info of the repository
		Files []*File `json:"files,omitempty"`
	}

	File struct {
		FileName string `json:"fileName,omitempty"`
		FilePath string `json:"filePath,omitempty"`
		Size     int64  `json:"size,omitempty"`
	}
)

func (r *Repository) Render(w io.Writer) {
	data := make([][]string, len(r.Files))
	for i, f := range r.Files {
		data[i] = []string{f.FileName, f.FilePath}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Artifact", "SrcPath"})
	table.SetFooter([]string{"Total Files", fmt.Sprintf("%d", r.Count)})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (r *Repository) ForEach(fn func(fileName, filePath string, size int64) error) error {
	for _, f := range r.Files {
		err := fn(f.FileName, f.FilePath, f.Size)
		if err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *Repository) ParallelForEach(fn func(fileName, filePath string, size int64) error) error {
	if settings.Concurrency <= 1 {
		return r.ForEach(fn)
	}

	dataChan := make(chan *File)
	go queueutil.Producer(r.Files, dataChan)

	if settings.Verbose {
		log.Debug("parallel foreach do migrate generic artifacts",
			logfields.Int("file size", r.Count),
			logfields.Int("concurrency", settings.Concurrency))
	}
	var wg sync.WaitGroup
	var goroutineCount int32 = 0
	errChan := make(chan error)
	execJobNum := make([]int32, settings.Concurrency)
	for i := 0; i < settings.Concurrency; i++ {
		wg.Add(1)
		execJobNum[i] = 0
		go queueutil.Consumer(dataChan, errChan, &wg, &execJobNum[i], func(f *File) error {
			atomic.AddInt32(&goroutineCount, 1)
			err := fn(f.FileName, f.FilePath, f.Size)
			atomic.AddInt32(&goroutineCount, -1)
			if err != nil && err == ErrForEachContinue {
				return nil
			}
			return err
		})
	}

	go logutil.WriteGoroutineFile(&goroutineCount, execJobNum)

	go func() {
		wg.Wait()
		// 关闭通道，表示所有的 goroutine 已经执行完毕
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
