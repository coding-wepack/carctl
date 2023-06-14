package repo

import (
	"os"
	"strings"

	repoapi "github.com/coding-wepack/carctl/pkg/api/repo"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/types/repo"
	repotypes "github.com/coding-wepack/carctl/pkg/types/repo"
	"github.com/coding-wepack/carctl/pkg/util/fileutil"
	"github.com/coding-wepack/carctl/pkg/util/jsonutil"
	"github.com/coding-wepack/carctl/pkg/util/randutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func AddProxySourceFromFile(filename string) error {
	log.Info("add proxy sources from the file", zap.String("filename", filename))

	exists, err := fileutil.IsFileExists(filename)
	if err != nil {
		return errors.Wrap(err, "failed to check if the file exists")
	}
	if !exists {
		return errors.New("file not found")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	if len(data) == 0 {
		return errors.New("invalid empty file")
	}

	var datasource map[string][]string
	if err = jsonutil.Unmarshal(data, &datasource); err != nil {
		return errors.Wrap(err, "failed to unmarshal datasource")
	}
	if len(datasource) == 0 {
		log.Info("No valid data found from the file")
		return nil
	}

	for u, sources := range datasource {
		if err = addProxySources(u, sources); err != nil {
			log.Warn("failed to add proxy sources",
				zap.String("proxyUrl", u), zap.String("err", err.Error()))
		}
	}

	return nil
}

func addProxySources(proxyUrl string, sources []string) error {
	_, err := repo.ValidateProxySourceUrl(proxyUrl)
	if err != nil {
		return err
	}

	for i, source := range sources {
		name := generateSourceName(source)
		if err = repoapi.AddProxySource(proxyUrl, settings.Cookie, &repotypes.ProxySourcePayload{
			Name:     name,
			Source:   source,
			Priority: i + 1,
		}); err != nil {
			log.Warn("failed to add proxy source",
				zap.String("repoProxyUrl", proxyUrl), zap.String("proxySource", source),
				zap.String("proxySourceName", name), zap.String("err", err.Error()))
		} else {
			log.Info("added a proxy source", zap.String("repoProxyUrl", proxyUrl),
				zap.String("proxySource", source), zap.String("proxySourceName", name))
		}
	}

	return nil
}

func generateSourceName(source string) string {
	source = strings.Trim(source, "/")
	chunks := strings.SplitAfter(source, "/")
	name := chunks[len(chunks)-1]
	name = name + "-" + randutil.RandString(4)

	const maxLen = 24
	if len(name) > maxLen {
		name = "auto-generated-" + randutil.RandString(5)
	}
	return name
}
