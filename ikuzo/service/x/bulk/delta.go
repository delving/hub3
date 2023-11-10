package bulk

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/hub3/fragments"
)

const (
	rdfType = "rdf"
	esType  = "es"
)

type redisStore struct {
	orgID            string
	spec             string
	revision         int
	previousRevision int
	c                *redis.Client
}

func (rs *redisStore) datasetKey() string {
	return fmt.Sprintf("%s:ds:%s", rs.orgID, rs.spec)
}

func (rs *redisStore) getStoreKey(hubID string) string {
	return hubID
}

func (rs *redisStore) storeRDFData(su fragments.SparqlUpdate) error {
	ctx := context.Background()
	key := rs.getStoreKey(su.HubID)
	err := rs.c.HMSet(ctx, key, map[string]interface{}{
		// rdfType:    su.Triples,
		"graphURI": su.NamedGraphURI,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func (rs *redisStore) addID(hubID, setType, hash string) (bool, error) {
	ctx := context.Background()
	err := rs.c.SAdd(ctx, rs.revisionSetName(setType, false), hubID).Err()
	if err != nil {
		return false, err
	}
	hasKey := hubID + ":" + hash
	if err := rs.c.SAdd(ctx, rs.revisionSetName(setType, true), hasKey).Err(); err != nil {
		return false, err
	}

	return rs.c.SIsMember(ctx, rs.previousSetName(setType, true), hasKey).Result()
}

func (rs *redisStore) revisionSetName(setType string, withHash bool) string {
	key := fmt.Sprintf("%s:rev:%d:%s", rs.datasetKey(), rs.revision, setType)
	if withHash {
		key += ":hash"
	}
	return key
}

func (rs *redisStore) previousSetName(setType string, withHash bool) string {
	key := fmt.Sprintf("%s:rev:%d:%s", rs.datasetKey(), rs.previousRevision, setType)
	if withHash {
		key += ":hash"
	}
	return key
}

func (rs *redisStore) dropOrphansQuery(orphans []string) (string, error) {
	var sb strings.Builder
	for _, orphan := range orphans {
		key := rs.getStoreKey(orphan)
		log.Printf("orphan key: %s", key)
		res, err := rs.c.HGet(context.Background(), key, "graphURI").Result()
		if err != nil {
			return "", fmt.Errorf("unable to get hash key: %w", err)
		}
		if res != "" {
			sb.WriteString(fmt.Sprintf("drop graph <%s> ;\n", res))
		}
	}

	return sb.String(), nil
}

func (rs *redisStore) findOrphans(setType string) ([]string, error) {
	ctx := context.Background()
	return rs.c.SDiff(ctx, rs.previousSetName(setType, false), rs.revisionSetName(setType, false)).Result()
}

func (rs *redisStore) SetRevision(current, previous int) error {
	rs.revision = current
	rs.previousRevision = previous
	ctx := context.Background()
	err := rs.c.HMSet(ctx, rs.datasetKey(), map[string]interface{}{
		"currentRevision":  current,
		"previousRevision": previous,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}
