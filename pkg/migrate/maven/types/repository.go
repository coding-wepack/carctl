package types

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/sliceutil"
	"github.com/pkg/errors"

	"github.com/olekukonko/tablewriter"
)

const (
	ArtifactNameRegexStr = `^[\w.-]+$`
)

var (
	ArtifactNameRegex = regexp.MustCompile(ArtifactNameRegexStr)
)

var (
	ErrForEachContinue = errors.New("continue")
)

type (
	Repository struct {
		// Path is file path to repository
		Path string `json:"path"`

		// FileCount is count of files of the repository
		FileCount int `json:"fileCount"`

		Groups []*Group `json:"groups,omitempty"`
	}

	Group struct {
		Name string `json:"name,omitempty"`

		Artifacts []*Artifact `json:"artifacts,omitempty"`
	}

	Artifact struct {
		Name string `json:"name,omitempty"`

		Versions []*Version `json:"versions,omitempty"`
	}

	Version struct {
		Name string `json:"name,omitempty"`

		Files []*VersionFile `json:"files,omitempty"`
	}

	VersionFile struct {
		// Name: e.g., spring-context-4.3.14.RELEASE.jar | spring-context-4.3.14.RELEASE.pom
		Name string `json:"name,omitempty"`

		// Path: /home/user/.m2/repository/org/springframework/spring-context/4.3.14.RELEASE/spring-context-4.3.14.RELEASE.jar
		Path string `json:"path,omitempty"`

		// DownloadUrl: fullUrl From Nexus Or Coding, e.g., http://localhost:8081/repository/maven-public/net/sf/json-lib/json-lib/2.2.2/json-lib-2.2.2-jdk15.jar
		DownloadUrl string `json:"downloadUrl,omitempty"`
	}

	Maven struct {
		group       string
		artifact    string
		version     string
		path        string
		downloadUrl string
	}
)

type (
	FlattenRepository struct {
		Path string `json:"path"`

		GroupCount    int `json:"-"`
		ArtifactCount int `json:"-"`
		VersionCount  int `json:"-"`
		FileCount     int `json:"-"`

		Files []FlattenVersionFile `json:"files,omitempty"`
	}

	FlattenVersionFile struct {
		Group    string `json:"group"`
		Artifact string `json:"artifact"`
		Version  string `json:"version"`
		Filename string `json:"filename"`
		FilePath string `json:"filePath"`
	}
)

func (r *Repository) Render(w io.Writer) {
	flattenRepository := r.Flatten()
	flattenRepository.Render(w)
}

func (r *Repository) Flatten() *FlattenRepository {
	flattenRepository := &FlattenRepository{
		Path:          r.Path,
		GroupCount:    r.GroupCount(),
		ArtifactCount: r.ArtifactCount(),
		VersionCount:  r.VersionCount(),
		FileCount:     r.GetFileCount(),
		Files:         make([]FlattenVersionFile, 0),
	}

	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					flattenRepository.Files = append(flattenRepository.Files, FlattenVersionFile{
						Group:    g.Name,
						Artifact: a.Name,
						Version:  v.Name,
						Filename: f.Name,
						FilePath: f.Path,
					})
				}
			}
		}
	}

	return flattenRepository
}

func (r *Repository) GetFileCount() int {
	if r.FileCount > 0 {
		return r.FileCount
	}

	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for range v.Files {
					r.FileCount++
				}
			}
		}
	}
	return r.FileCount
}

func (r *Repository) ForEach(fn func(group, artifact, version, path, downloadUrl string) error) error {
	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					if err := fn(g.Name, a.Name, v.Name, f.Path, f.DownloadUrl); err != nil {
						if err == ErrForEachContinue {
							continue
						}
						return err
					}
				}
			}
		}
	}
	return nil
}

func (r *Repository) ParallelForEach(fn func(group, artifact, version, path, downloadUrl string) error) error {
	if settings.Concurrency <= 1 {
		return r.ForEach(fn)
	}
	mavens := make([]*Maven, 0)
	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					mavens = append(mavens, &Maven{
						group:       g.Name,
						artifact:    a.Name,
						version:     v.Name,
						path:        f.Path,
						downloadUrl: f.DownloadUrl,
					})
				}
			}
		}
	}

	var wg sync.WaitGroup
	chunks := sliceutil.Chunk(mavens, settings.Concurrency)
	errChan := make(chan error, len(chunks))
	if settings.Verbose {
		log.Debug("parallel foreach do migrate maven artifacts",
			logfields.Int("file size", len(mavens)),
			logfields.Int("concurrency", settings.Concurrency),
			logfields.Int("chunk size", len(chunks)))
	}
	for i, items := range chunks {
		if len(items) == 0 {
			continue
		}
		wg.Add(1)
		if settings.Verbose {
			log.Debug(fmt.Sprintf("do migrate maven artifacts with chunk[%d]", i), logfields.Int("size", len(items)))
		}
		go func(items []*Maven) {
			defer wg.Done()
			for _, item := range items {
				if err := fn(item.group, item.artifact, item.version, item.path, item.downloadUrl); err != nil {
					if err == ErrForEachContinue {
						continue
					}
					errChan <- err
				}
			}
		}(items)
	}
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

func (r *Repository) AddVersionFile(groupName, artifactName, versionName, filename, filePath string) {
	r.AddVersionFileBase(groupName, artifactName, versionName, filename, filePath, "")
}

func (r *Repository) AddVersionFileBase(groupName, artifactName, versionName, filename, filePath, downloadUrl string) {
	if !r.HasGroup(groupName) {
		r.AddGroupName(groupName)
	}
	for _, g := range r.Groups {
		if g.Name == groupName {
			if !g.HasArtifact(artifactName) {
				g.AddArtifactName(artifactName)
			}
			for _, art := range g.Artifacts {
				if art.Name == artifactName {
					if !art.HasVersion(versionName) {
						art.AddVersion(versionName)
					}
					for _, v := range art.Versions {
						if v.Name == versionName {

							if !v.HasFile(filename) {
								v.AddFile(filename, filePath, downloadUrl)
							}

						}
					}
				}
			}
		}
	}
}

func (r *Repository) HasGroup(groupName string) bool {
	for _, g := range r.Groups {
		if g.Name == groupName {
			return true
		}
	}
	return false
}

func (r *Repository) AddGroupName(group string) {
	r.Groups = append(r.Groups, &Group{Name: group})
}

func (r *Repository) GroupCount() int {
	return len(r.Groups)
}

func (r *Repository) ArtifactCount() int {
	var count int
	for _, g := range r.Groups {
		count += len(g.Artifacts)
	}
	return count
}

func (r *Repository) VersionCount() int {
	var count int
	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			count += len(a.Versions)
		}
	}
	return count
}

func (r *Repository) CleanInvalidMetadata(fileCount int) int {
	for _, g := range r.Groups {
		var invalidIndex []int
		for i, a := range g.Artifacts {
			// 仅有一个 metadata 文件，则需要过滤掉
			if len(a.Versions) == 1 && strings.EqualFold(a.Versions[0].Name, "Metadata") {
				invalidIndex = append(invalidIndex, i)
				fileCount--
			}
		}
		g.Artifacts = deleteIndexes(g.Artifacts, invalidIndex)
	}
	var invalidIndex []int
	for i, g := range r.Groups {
		if len(g.Artifacts) == 0 {
			invalidIndex = append(invalidIndex, i)
		}
	}
	r.Groups = deleteIndexes(r.Groups, invalidIndex)
	return fileCount
}

func deleteIndexes[T any](s []T, indexes []int) []T {
	if len(indexes) == 0 {
		return s
	}
	// 降序排列索引以确保正确删除
	sort.Sort(sort.Reverse(sort.IntSlice(indexes)))
	for _, index := range indexes {
		s = append(s[:index], s[index+1:]...)
	}
	return s
}

func (g *Group) HasArtifact(artifactName string) bool {
	for _, art := range g.Artifacts {
		if art.Name == artifactName {
			return true
		}
	}
	return false
}

func (g *Group) AddArtifactName(artifactName string) {
	g.Artifacts = append(g.Artifacts, &Artifact{Name: artifactName})
}

func (a *Artifact) HasVersion(version string) bool {
	for _, v := range a.Versions {
		if v.Name == version {
			return true
		}
	}
	return false
}

func (a *Artifact) AddVersion(version string) {
	a.Versions = append(a.Versions, &Version{Name: version})
}

func (v *Version) HasFile(filename string) bool {
	for _, f := range v.Files {
		if f.Name == filename {
			return true
		}
	}
	return false
}

func (v *Version) AddFile(filename, filePath, downloadUrl string) {
	v.Files = append(v.Files, &VersionFile{Name: filename, Path: filePath, DownloadUrl: downloadUrl})
}

func (f *FlattenRepository) Render(w io.Writer) {
	data := make([][]string, len(f.Files))
	for i, v := range f.Files {
		data[i] = []string{
			v.Group, v.Artifact, v.Version, v.Filename,
		}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Group ID", "Artifact ID", "Version", "File"})
	table.SetFooter([]string{
		f.renderFooterCount("Groups", f.GetGroupCount()),
		f.renderFooterCount("Artifacts", f.GetArtifactCount()),
		f.renderFooterCount("Versions", f.GetVersionCount()),
		f.renderFooterCount("Files", f.GetFileCount()),
	})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (f *FlattenRepository) GetGroupCount() int {
	return f.GroupCount
}

func (f *FlattenRepository) GetArtifactCount() int {
	return f.ArtifactCount
}

func (f *FlattenRepository) GetVersionCount() int {
	return f.VersionCount
}

func (f *FlattenRepository) GetFileCount() int {
	return f.FileCount
}

func (f *FlattenRepository) renderFooterCount(itemName string, count int) string {
	return fmt.Sprintf("Total %s: %d", itemName, count)
}

func (f *FlattenRepository) ForEach(fn func(group, artifact, version, filePath string) error) error {
	for _, r := range f.Files {
		if err := fn(r.Group, r.Artifact, r.Version, r.FilePath); err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}
