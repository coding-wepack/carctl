package types

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type Report struct {
	SucceededResult []Result `json:"succeeded,omitempty"`
	SkippedResult   []Result `json:"skipped,omitempty"`
	FailedResult    []Result `json:"failed,omitempty"`
}

type Result struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Time    int64  `json:"time"`
	Message string `json:"message"`
}

func NewReport() *Report {
	return &Report{
		SucceededResult: make([]Result, 0),
		SkippedResult:   make([]Result, 0),
		FailedResult:    make([]Result, 0),
	}
}

func (r *Report) TotalCount() int {
	return len(r.SucceededResult) + len(r.SkippedResult) + len(r.FailedResult)
}

func (r *Report) AddSucceededResult(name, path, msg string) {
	r.SucceededResult = append(r.SucceededResult, Result{
		Name:    name,
		Path:    path,
		Message: msg,
	})
}

func (r *Report) AddSkippedResult(name, path, msg string) {
	r.SkippedResult = append(r.SkippedResult, Result{
		Name:    name,
		Path:    path,
		Message: msg,
	})
}

func (r *Report) AddFailedResult(name, path, msg string) {
	r.FailedResult = append(r.FailedResult, Result{
		Name:    name,
		Path:    path,
		Message: msg,
	})
}

func (r *Report) AddSucceededResult2(name, path, msg string, size, time int64) {
	r.SucceededResult = append(r.SucceededResult, Result{
		Name:    name,
		Path:    path,
		Size:    size,
		Time:    time,
		Message: msg,
	})
}

func (r *Report) AddSkippedResult2(name, path, msg string, size, time int64) {
	r.SkippedResult = append(r.SkippedResult, Result{
		Name:    name,
		Path:    path,
		Size:    size,
		Time:    time,
		Message: msg,
	})
}

func (r *Report) AddFailedResult2(name, path, msg string, size, time int64) {
	r.FailedResult = append(r.FailedResult, Result{
		Name:    name,
		Path:    path,
		Size:    size,
		Time:    time,
		Message: msg,
	})
}

func (r *Report) Render(w io.Writer) {
	totalResult := r.mergeIntoOneResult(true)
	size := len(totalResult)
	data := make([][]string, size)
	for i, result := range totalResult {
		data[i] = []string{
			result.Name, result.Path, result.Message,
		}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Artifact", "Src Path", "Result"})
	table.SetFooter([]string{
		"Total", strconv.Itoa(size), "",
	})
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (r *Report) Render2(w io.Writer) {
	totalResult := r.mergeIntoOneResult(true)
	count := len(totalResult)
	var totalSize float64 = 0
	var totalTime float64 = 0
	data := make([][]string, count)
	for i, result := range totalResult {
		data[i] = []string{
			result.Name, result.Path, fmt.Sprintf("%d", result.Size), fmt.Sprintf("%d", result.Time), result.Message,
		}
		totalSize += float64(result.Size) / 1024 / 1024
		totalTime += float64(result.Time) / 1000
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Artifact", "Src Path", "File Size(kb)", "Migrate Time(s)", "Result"})
	table.SetFooter([]string{
		"Total", strconv.Itoa(count), fmt.Sprintf("%f", totalSize), fmt.Sprintf("%f", totalTime), "",
	})
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (r *Report) mergeIntoOneResult(sortByName bool) []Result {
	totalResult := append(r.SucceededResult, r.SkippedResult...)
	totalResult = append(totalResult, r.FailedResult...)

	if sortByName {
		sort.Slice(totalResult, func(i, j int) bool {
			return totalResult[i].Name < totalResult[j].Name
		})
	}

	return totalResult
}
