package types

import (
	"fmt"
	"io"
	"strings"
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
		// IsTls is remote url is tls
		IsTls bool `json:"isTls"`

		// Path is image path to repository
		Path string `json:"path"`

		// Count is count of image of the repository
		Count int `json:"-"`

		// Images is image info of the repository
		Images []*Image `json:"images,omitempty"`
	}

	Image struct {
		SrcPath    string `json:"srcPath,omitempty"`
		PkgName    string `json:"pkgName,omitempty"`
		Version    string `json:"version,omitempty"`
		Tag        string `json:"tag,omitempty"`
		SrcPkgName string `json:"SrcPkgName,omitempty"`
	}
)

func (r *Repository) Render(w io.Writer) {
	data := make([][]string, len(r.Images))
	for i, v := range r.Images {
		data[i] = []string{v.PkgName, v.Version, v.SrcPath}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Artifact", "Version", "SrcPath"})
	table.SetFooter([]string{"", "Total Images", fmt.Sprintf("%d", r.Count)})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (r *Repository) CheckDuplication(w io.Writer) bool {
	var duplication []*Image
	imageMap := make(map[string]*Image, len(r.Images))
	for _, v := range r.Images {
		image, ok := imageMap[v.Tag]
		if !ok {
			imageMap[v.Tag] = v
			continue
		}
		duplication = append(duplication, image, v)
	}
	if len(duplication) == 0 {
		log.Info("congratulations! no duplicate anomalies")
		return false
	}
	duplication = RemoveDuplication(duplication)

	data := make([][]string, len(duplication))
	for i, v := range duplication {
		data[i] = []string{v.PkgName, v.Version, v.SrcPath}
	}

	log.Info("Duplicate Artifacts Info:")
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Artifact", "Version", "SrcPath"})
	table.SetFooter([]string{"", "Total Images", fmt.Sprintf("%d", len(duplication))})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
	return true
}

func (r *Repository) ForEach(fn func(image *Image, srcRepo, dstRepo string, isTlsSrc, isTlsDst bool) error) error {
	isTlsDst, dstRepo := parseDstUrl(settings.GetDstWithoutSlash())
	srcRepo := strings.Trim(r.Path, "/")
	for _, image := range r.Images {
		if err := fn(image, srcRepo, dstRepo, r.IsTls, isTlsDst); err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *Repository) ParallelForEach(fn func(image *Image, srcRepo, dstRepo string, isTlsSrc, isTlsDst bool) error) error {
	if settings.Concurrency <= 1 {
		return r.ForEach(fn)
	}
	isTlsDst, dstRepo := parseDstUrl(settings.GetDstWithoutSlash())
	srcRepo := strings.Trim(r.Path, "/")

	dataChan := make(chan *Image)
	go queueutil.Producer(r.Images, dataChan)

	if settings.Verbose {
		log.Debug("parallel foreach do migrate docker artifacts",
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
		go queueutil.Consumer(dataChan, errChan, &wg, &execJobNum[i], func(image *Image) error {
			atomic.AddInt32(&goroutineCount, 1)
			err := fn(image, srcRepo, dstRepo, r.IsTls, isTlsDst)
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

func (r *Repository) Tag(fn func(image *Image)) {
	for _, image := range r.Images {
		fn(image)
	}
}

func parseDstUrl(dstUrl string) (isTls bool, registryUrl string) {
	isTls = strings.HasPrefix(dstUrl, "https://")
	if isTls {
		registryUrl = strings.TrimPrefix(dstUrl, "https://")
	} else {
		registryUrl = strings.TrimPrefix(dstUrl, "http://")
	}
	return
}

func RemoveDuplication(images []*Image) []*Image {
	m := make(map[string]struct{})
	result := make([]*Image, 0)

	for _, image := range images {
		if _, ok := m[image.SrcPath]; !ok {
			result = append(result, image)
			m[image.SrcPath] = struct{}{}
		}
	}
	return result
}
