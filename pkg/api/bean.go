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
