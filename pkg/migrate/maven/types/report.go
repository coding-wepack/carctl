package types

import (
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Report struct {
	SucceededResult []Result `json:"succeeded,omitempty"`
	SkippedResult   []Result `json:"skipped,omitempty"`
	FailedResult    []Result `json:"failed,omitempty"`
}

type Result struct {
	// Name: group:artifact:version
	Name    string `json:"name"`
	Path    string `json:"path"`
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

func (r *Report) AddSucceededResult(group, artifact, version, path, msg string) {
	r.SucceededResult = append(r.SucceededResult, Result{
		Name:    strings.Join([]string{group, artifact, version}, ":"),
		Path:    path,
		Message: msg,
	})
}

func (r *Report) AddSkippedResult(group, artifact, version, path, msg string) {
	r.SkippedResult = append(r.SkippedResult, Result{
		Name:    strings.Join([]string{group, artifact, version}, ":"),
		Path:    path,
		Message: msg,
	})
}

func (r *Report) AddFailedResult(group, artifact, version, path, msg string) {
	r.FailedResult = append(r.FailedResult, Result{
		Name:    strings.Join([]string{group, artifact, version}, ":"),
		Path:    path,
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
	table.SetHeader([]string{"Artifact", "Path", "Result"})
	table.SetFooter([]string{
		"Total", strconv.Itoa(size), "",
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
