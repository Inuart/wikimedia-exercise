package shortdescription

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	ContactInfo string
	CacheSize   int           // defaults to DefaultCacheSize
	CachedTTL   time.Duration // defaults to DefaultCachedTTl
	HttpClient  HttpDoer
}

const (
	DefaultCacheSize = 500
	DefaultCachedTTl = time.Hour
)

type HttpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	contactInfoMsg = "You need to provide an your contact info. See https://meta.wikimedia.org/wiki/User-Agent_policy"
	userAgentFmt   = "ShortDescriptionAPI/v0.0.0 (%s) hiring-exercise/v0.0.0"
)

func New(cfg Config) (Describer, error) {
	if cfg.ContactInfo == "" {
		return Describer{}, errors.New("shortdescription.New: " + contactInfoMsg)
	}

	if cfg.CacheSize == 0 {
		cfg.CacheSize = DefaultCacheSize
	}

	if cfg.CachedTTL == 0 {
		cfg.CachedTTL = DefaultCachedTTl
	}

	if cfg.HttpClient == nil {
		cfg.HttpClient = http.DefaultClient
	}

	cache, err := newCache(cfg.CacheSize, cfg.CachedTTL)
	if err != nil {
		return Describer{}, fmt.Errorf("cache creation failed: %w", err)
	}

	return Describer{
		userAgent:  fmt.Sprintf(userAgentFmt, cfg.ContactInfo),
		httpClient: cfg.HttpClient,
		cache:      cache,
	}, nil
}

type Describer struct {
	userAgent  string
	httpClient HttpDoer
	cache      cache
}

func (d Describer) ShortDescription(ctx context.Context, person, userAgent string) (ShortDescription, error) {
	if person == "" {
		return ShortDescription{}, fmt.Errorf("%w: person is empty", ErrInvalidArgument)
	}

	if userAgent == "" {
		return ShortDescription{}, fmt.Errorf("%w: userAgent is empty", ErrInvalidArgument)
	}

	person, err := url.QueryUnescape(person)
	if err != nil {
		return ShortDescription{}, fmt.Errorf("%w: person is wrongly encoded: %v", ErrInvalidArgument, err)
	}

	// Normalize person title first so that caching is more effective. According to
	// https://www.mediawiki.org/wiki/API:Query this means capitalizing the first character and
	// replacing underscores with spaces.
	person = strings.Split(person, "|")[0] // deal with only one query
	person = strings.ToUpper(person[:1]) + person[1:]
	person = strings.ReplaceAll(person, "_", " ")

	shortDescription, ok := d.cache.Get(person)
	if ok {
		return ShortDescription{
			Person:      person,
			Description: shortDescription,
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", getShortDescriptionURL(person), nil)
	if err != nil {
		return ShortDescription{}, fmt.Errorf("cannot create request: %w", err)
	}

	// required by the API
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Api-User-Agent", d.userAgent)

	// the "Accept-Encoding" is automatically set so there's no need to add "gzip".

	res, err := d.httpClient.Do(req)
	if err != nil {
		return ShortDescription{}, fmt.Errorf("failed to initiate fetch: %w", err)
	}

	defer res.Body.Close()

	if err := responseError(res); err != nil {
		return ShortDescription{}, err
	}

	shortDescription, err = readShortDescription(res.Body)
	if err != nil {
		return ShortDescription{}, err
	}

	d.cache.Add(person, shortDescription)

	return ShortDescription{
		Person:      person,
		Description: shortDescription,
	}, nil
}
