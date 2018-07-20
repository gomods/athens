package download

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/gomods/athens/pkg/log"

	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/google/go-github/github"
	"github.com/marwan-at-work/vgop/semver"
	errs "github.com/pkg/errors"
)

// PathList URL.
const PathList = "/{module:.+}/@v/list"

// ListHandler implements GET baseURL/module/@v/list
func ListHandler(
	lister storage.Lister,
	eng *render.Engine,
	isProxy bool,
	lggr *log.Logger,
) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		sp := buffet.SpanFromContext(c)
		sp.SetOperationName("listHandler")
		mod, err := paths.GetModule(c)
		if err != nil {
			return err
		}

		versions, err := lister.List(mod)
		if storage.IsNotFoundError(err) || len(versions) == 0 {
			if isProxy {
				// Go uses the http.DefaultClient which implicitly handles up to 10 reqdirects.
				// therefore, it's safe to redirect to whatever server/endpoint we like as long
				// as that server returns an answer that Go expects.
				return c.Redirect(http.StatusMovedPermanently, env.GetOlympusEndpoint())
			} else if !isGithubPath(mod) {
				c.Response().WriteHeader(http.StatusNotFound)
				return nil
			}

			info, err := getInfoFromPath(c, mod)
			if err != nil {
				lggr.WithFields(logrus.Fields{
					"module": mod,
					"method": "ListHandler.getInfoFromPath",
				}).SystemErr(err)
			}

			if err != nil || info.typ != sourceGithub {
				c.Response().WriteHeader(http.StatusNotFound)
				return nil
			}

			versions = info.tags
		} else if err != nil {
			return errs.WithStack(err)
		}

		return c.Render(http.StatusOK, eng.String(strings.Join(versions, "\n")))
	}
}

func isGithubPath(path string) bool {
	els := strings.Split(path, "/")
	return els[0] == "github.com"
}

type importInfo struct {
	typ        sourceType
	githubInfo githubInfo
	tags       []string
}

type githubInfo struct {
	owner string
	repo  string
}

type sourceType int

const (
	sourceGithub sourceType = iota + 1
	sourceGitlab
	sourceBitbucket
	sourceVCS
)

func getInfoFromPath(ctx context.Context, path string) (importInfo, error) {
	els := strings.Split(path, "/")
	switch els[0] {
	case "github.com":
		if len(els) != 3 {
			return importInfo{}, errors.New("unparsable github path: " + path)
		}
		owner := els[1]
		repo := els[2]

		gc := github.NewClient(nil)
		var allTags []*github.RepositoryTag
		page := 1
		for {
			tags, _, err := gc.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
				Page:    page,
				PerPage: 100,
			})
			if err != nil {
				return importInfo{}, err
			}

			if len(tags) == 0 {
				break
			}

			allTags = append(allTags, tags...)
			page++
		}

		ai := importInfo{
			typ: sourceGithub,
			githubInfo: githubInfo{
				owner: owner,
				repo:  repo,
			},
			tags: []string{},
		}

		for _, t := range allTags {
			if tag := t.GetName(); semver.IsValid(tag) {
				ai.tags = append(ai.tags, tag)
			}
		}

		return ai, nil
	}

	return importInfo{}, errors.New("unsupported API")
}
