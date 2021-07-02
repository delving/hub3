package ead

import (
	"encoding/json"
	"time"

	"github.com/allegro/bigcache"
	cfg "github.com/delving/hub3/config"
	"github.com/rs/zerolog"
)

var httpCache *bigcache.BigCache

func newBigCache() {
	config := bigcache.Config{
		Shards:           1024,
		HardMaxCacheSize: cfg.Config.Cache.HardMaxCacheSize,
		LifeWindow:       time.Duration(cfg.Config.Cache.LifeWindowMinutes) * time.Minute,
		CleanWindow:      5 * time.Minute,
		MaxEntrySize:     cfg.Config.Cache.MaxEntrySize,
	}

	var err error

	httpCache, err = bigcache.NewBigCache(config)
	if err != nil {
		cfg.Config.Logger.Warn().Err(err).Msg("cannot start bigcache running without cache; %#v")
	}

	rlog := cfg.Config.Logger.With().Str("test", "sublogger").Logger()
	rlog.Info().Msg("starting bigCache for request caching")
}

func getCachedRequest(requestKey string, rlog *zerolog.Logger) *SearchResponse {
	entry, cacheErr := httpCache.Get(requestKey)
	if cacheErr != nil {
		rlog.Debug().Str("cache_key", requestKey).Err(cacheErr).Msg("cache miss")
		return nil
	}

	var eadResponse SearchResponse

	jsonErr := json.Unmarshal(entry, &eadResponse)
	if jsonErr != nil {
		rlog.Warn().Err(jsonErr).Msg("unable to unmarshall cached response")
		return nil
	}

	rlog.Debug().Str("cache_key", requestKey).Msg("returning response from cache")

	return &eadResponse
}

func storeResponseInCache(requestKey string, response *SearchResponse, rlog *zerolog.Logger) {
	b, err := json.Marshal(response)
	if err != nil {
		rlog.Error().Err(err).Msg("unable to marshal eadResponse for caching")
	} else {
		cacheErr := httpCache.Set(requestKey, b)
		if cacheErr != nil {
			rlog.Error().Err(cacheErr).Msg("unable to cache searchResponse")
		}
		rlog.Debug().Str("cache_key", requestKey).Msg("set cache for key")
	}
}
