package api

type DescribeTeamArtifactsReq struct {
	Action     string
	PageNumber int
	PageSize   int
	Rule       *DescribeTeamArtifactsRule
}

type DescribeTeamArtifactsRule struct {
	ProjectName []string
	Repository  []string
}

type DescribeTeamArtifactsResp struct {
	Response *DescribeTeamArtifactsResponse
}

type DescribeTeamArtifactsResponse struct {
	RequestId string
	Error     *Error
	Data      *DescribeTeamArtifactsData
}

type DescribeTeamArtifactsData struct {
	InstanceSet []*Artifacts
	PageNumber  int
	PageSize    int
	TotalCount  int
}

type Artifacts struct {
	Package        string
	PackageVersion string
}

type Error struct {
	Message string
	Code    string
}

type DescribeRepoFileListReq struct {
	Action            string
	ContinuationToken string
	PageSize          int
	Project           string
	Repository        string
}

type DescribeRepoFileListResp struct {
	Response *DescribeRepoFileListResponse
}

type DescribeRepoFileListResponse struct {
	RequestId string
	Error     *Error
	Data      *RepoFileListData
}

type RepoFileListData struct {
	ContinuationToken string
	InstanceSet       []*RepoFile
}

type RepoFile struct {
	DownloadUrl  string
	Path         string
	ArtifactType string
	Host         string
	Project      string
	Repository   string
	PackageName  string
	VersionName  string
	Hash         string
}

type CreateArtPropertiesReq struct {
	Action         string
	ProjectName    string
	Repository     string
	Package        string
	PackageVersion string
	PropertySet    []*Property
}

type Property struct {
	Name  string
	Value string
}

type CreateArtPropertiesResp struct {
	Response *CreateArtPropertiesResponse
}
type CreateArtPropertiesResponse struct {
	RequestId string
	Error     *Error
}
