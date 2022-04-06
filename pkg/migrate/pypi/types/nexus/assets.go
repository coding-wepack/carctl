package nexus

import "time"

type GetAssetsResponse struct {
	Items             []Item `json:"items"`
	ContinuationToken string `json:"continuationToken"`
}

type Item struct {
	DownloadURL string `json:"downloadUrl"`
	Path        string `json:"path"`
	ID          string `json:"id"`
	Repository  string `json:"repository"`
	Format      string `json:"format"`
	Checksum    struct {
		Sha1   string `json:"sha1"`
		Sha512 string `json:"sha512"`
		Sha256 string `json:"sha256"`
		Md5    string `json:"md5"`
	} `json:"checksum"`
	ContentType    string      `json:"contentType"`
	LastModified   time.Time   `json:"lastModified"`
	BlobCreated    time.Time   `json:"blobCreated"`
	LastDownloaded interface{} `json:"lastDownloaded"`
	Pypi           struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Platform string `json:"platform"`
	} `json:"pypi"`
}
