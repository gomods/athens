package download

import (
	"context"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
)

// Protocol is the download protocol which mirrors
// the http requests that cmd/go makes to the proxy.
type Protocol interface {
	// List implements GET /{module}/@v/list
	List(ctx context.Context, mod string) ([]string, error)

	// Info implements GET /{module}/@v/{version}.info
	Info(ctx context.Context, mod, ver string) ([]byte, error)

	// Latest implements GET /{module}/@latest
	Latest(ctx context.Context, mod string) (*storage.RevInfo, error)

	// GoMod implements GET /{module}/@v/{version}.mod
	GoMod(ctx context.Context, mod, ver string) ([]byte, error)

	// Zip implements GET /{module}/@v/{version}.zip
	Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error)
}

// Wrapper helps extend the main protocol's functionality with addons.
type Wrapper func(Protocol) Protocol

// Opts specifies download protocol options to avoid long func signature.
type Opts struct {
	Storage storage.Backend
	Stasher stash.Stasher
	Lister  UpstreamLister
}

// New returns a full implementation of the download.Protocol
// that the proxy needs. New also takes a variadic list of wrappers
// to extend the protocol's functionality (see addons package).
// The wrappers are applied in order, meaning the last wrapper
// passed is the Protocol that gets hit first.
func New(opts *Opts, wrappers ...Wrapper) Protocol {
	var p Protocol = &protocol{opts.Storage, opts.Stasher, opts.Lister}
	for _, w := range wrappers {
		p = w(p)
	}

	return p
}

type protocol struct {
	storage storage.Backend
	stasher stash.Stasher
	lister  UpstreamLister
}

func (p *protocol) List(ctx context.Context, mod string) ([]string, error) {
	const op errors.Op = "protocol.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	lggr := log.EntryFromContext(ctx)

	// if the storage implements ListCacher, check the expiry. If the list
	// cache for this module isn't expired yet, just return what's in storage.
	// Otherwise, just continue on to the normal storage + VCS check
	listCacher, hasListCacher := p.storage.(storage.ListCacher)
	skipVCS := false
	if hasListCacher {
		expiresIn, err := listCacher.ExpiresIn(ctx, mod)
		if err == nil && expiresIn > 0 {
			skipVCS = true
		}
		// don't use an else here, and don't return the error, because we
		// need to continue down to the below list logic
		if err != nil {
			lggr.Warnf(
				"The list cacher returned an error (%s). Athens will continue to fetch the version list for module %s from the VCS",
				err,
				mod,
			)
		}
	}

	if skipVCS {
		vers, err := p.storage.List(ctx, mod)
		if err == nil {
			// see the below call to removePseudoVersions for details on why
			// this function must be called
			return removePseudoVersions(vers), nil
		}
		// if the storage list call failed, we still want to move on, but
		// log
		lggr.Warnf(
			"The list cache is turned on and is not expired, so Athens requested module %s from storage only. Storage failed with %s, but Athens will continue to fetch from VCS",
			mod,
			err,
		)
	}

	var strList, goList []string
	var sErr, goErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		strList, sErr = p.storage.List(ctx, mod)
	}()

	go func() {
		defer wg.Done()
		_, goList, goErr = p.lister.List(ctx, mod)
	}()

	wg.Wait()

	// if we got an unexpected storage err then we can not guarantee that the end result contains all versions
	// a tag or repo could have been deleted
	if sErr != nil {
		return nil, errors.E(op, sErr)
	}

	// if i.e. github is unavailable we should fail as well so that the behavior of the proxy is stable.
	// otherwise we will get different results the next time because i.e. GH is up again
	isUnexpGoErr := goErr != nil && !errors.IsRepoNotFoundErr(goErr)
	if isUnexpGoErr {
		return nil, errors.E(op, goErr)
	}

	isRepoNotFoundErr := goErr != nil && errors.IsRepoNotFoundErr(goErr)
	storageEmpty := len(strList) == 0
	if isRepoNotFoundErr && storageEmpty {
		return nil, errors.E(op, errors.M(mod), errors.KindNotFound, goErr)
	}

	strListSemVers := removePseudoVersions(strList)
	// if the repo does not exist but athens already saved some versions
	// return those so that running go get github.com/my/mod gives us the newest saved version
	// we should only do that if exclusively pseudo-versions have been saved
	// otherwise @latest would not return the latest stable version but latest commit
	if isRepoNotFoundErr && len(strListSemVers) == 0 {
		return strList, nil
	}
	// if the repo exists we have to filter out pseudo versions to prevent following scenario:
	// user does go get github.com/my/mod
	// there is no sem-ver and so the /list endpoint returns nothing, then /latests gets hit
	// Athens saves the pseudo version x1
	// from now on every time user runs go get github.com/my/mod she/he will get pseudo version x1 even if a newer version x2 exists
	// this is because /list returns non-empty list of versions (x1) and so /latest wont get hit
	return union(goList, strListSemVers), nil
}

var pseudoVersionRE = regexp.MustCompile(`^v[0-9]+\.(0\.0-|\d+\.\d+-([^+]*\.)?0\.)\d{14}-[A-Za-z0-9]+(\+incompatible)?$`)

func removePseudoVersions(allVersions []string) []string {
	var vers []string
	for _, v := range allVersions {
		// copied from go cmd https://github.com/golang/go/blob/master/src/cmd/go/internal/modfetch/pseudo.go#L93
		isPseudoVersion := strings.Count(v, "-") >= 2 && pseudoVersionRE.MatchString(v)
		if !isPseudoVersion {
			vers = append(vers, v)
		}
	}
	return vers
}

func (p *protocol) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "protocol.Latest"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	lr, _, err := p.lister.List(ctx, mod)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return lr, nil
}

func (p *protocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "protocol.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	info, err := p.storage.Info(ctx, mod, ver)
	var newVer string
	if errors.IsNotFoundErr(err) {
		newVer, err = p.stasher.Stash(ctx, mod, ver)
		if err != nil {
			return nil, errors.E(op, err)
		}
		info, err = p.storage.Info(ctx, mod, newVer)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}

	return info, nil
}

func (p *protocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "protocol.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	goMod, err := p.storage.GoMod(ctx, mod, ver)
	var newVer string
	if errors.IsNotFoundErr(err) {
		newVer, err = p.stasher.Stash(ctx, mod, ver)
		if err != nil {
			return nil, errors.E(op, err)
		}
		goMod, err = p.storage.GoMod(ctx, mod, newVer)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}

	return goMod, nil
}

func (p *protocol) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
	const op errors.Op = "protocol.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	zip, err := p.storage.Zip(ctx, mod, ver)
	var newVer string
	if errors.IsNotFoundErr(err) {
		newVer, err = p.stasher.Stash(ctx, mod, ver)
		if err != nil {
			return nil, errors.E(op, err)
		}
		zip, err = p.storage.Zip(ctx, mod, newVer)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}

	return zip, nil
}

// union concatenates two version lists and removes duplicates
func union(list1, list2 []string) []string {
	if list1 == nil {
		list1 = []string{}
	}
	if list2 == nil {
		list2 = []string{}
	}
	list := append(list1, list2...)
	unique := []string{}
	m := make(map[string]struct{})
	for _, v := range list {
		if _, ok := m[v]; !ok {
			unique = append(unique, v)
			m[v] = struct{}{}
		}
	}
	return unique
}
