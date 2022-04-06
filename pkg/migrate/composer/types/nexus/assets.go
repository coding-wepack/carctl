package nexus

import (
	"time"
)

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
		Sha256 string `json:"sha256"`
		Md5    string `json:"md5"`
	} `json:"checksum"`
	ContentType    string    `json:"contentType"`
	LastModified   time.Time `json:"lastModified"`
	BlobCreated    time.Time `json:"blobCreated"`
	LastDownloaded time.Time `json:"lastDownloaded"`
}

type ComposerItem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Dist    struct {
		URL       string `json:"url"`
		Type      string `json:"type"`
		Reference string `json:"reference"`
		Shasum    string `json:"shasum"`
	} `json:"dist"`
	Time        time.Time `json:"time"`
	UID         int       `json:"uid"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
}

type Packages struct {
	Packages PackageInfo `json:"packages"`
}

type VersionInfo map[string]*ComposerItem

type PackageInfo map[string]VersionInfo
