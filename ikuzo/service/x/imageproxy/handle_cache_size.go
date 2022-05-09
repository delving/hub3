package imageproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
)

func (s *Service) rebuildCacheMetrics(w http.ResponseWriter, r *http.Request) {
	if err := s.buildCacheMetrics(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, s.cm)
}

func (s *Service) handleCacheStats() http.HandlerFunc {
	type stats struct {
		MaxCacheSize     int
		CurrentCacheSize int
		PercentInUse     int
	}

	maxPercentage := s.maxSizeCacheDir / 100

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, err := s.externalCacheSize()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := stats{
			MaxCacheSize:     s.maxSizeCacheDir,
			CurrentCacheSize: size,
			PercentInUse:     size / maxPercentage,
		}

		render.JSON(w, r, resp)
	})
}

func (s *Service) removeOldestFiles(nr int) error {
	if s.cacheDir == "" {
		return nil
	}

	rawCmd := fmt.Sprintf(
		"find %s -type f -printf '%%T+ %%p\\n' | sort | head -n %d | xargs rm -rf {}",
		s.cacheDir,
		nr,
	)

	cmd := exec.Command("bash", "-c", rawCmd)

	log.Printf("%s", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n%s", cmd.String(), out)
		return err
	}

	return nil
}

func (s *Service) cachePercentageInUse() (int, error) {
	cacheSize, err := s.externalCacheSize()
	if err != nil {
		return 0, err
	}

	return (cacheSize / (s.maxSizeCacheDir / 100)), nil
}

func (s *Service) externalCacheSize() (int, error) {
	args := []string{
		"-d",
		"0",
		s.cacheDir,
	}

	cmd := exec.Command("du", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	parts := strings.SplitN(string(out), "\t", 2)

	i, err := strconv.Atoi(parts[0])
	if err != nil {
		s.log.Error().Msgf("unable to convert to int: %#v", err)
		return 0, err
	}

	return i, nil
}

func (s *Service) startCacheWorker() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelWorker = cancel

	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for {
			percentageInUse, err := s.cachePercentageInUse()
			if err != nil {
				s.log.Error().Err(err).Msgf("unable to get cache stats: %s", err)
				continue
			}

			// only run when at 95%
			if percentageInUse > 95 {
				s.log.Info().Int("cacheInUse", percentageInUse).Msg("reached threshold; start running cache cleaner")

				for {
					if removeErr := s.removeOldestFiles(1000); removeErr != nil {
						s.log.Error().Err(removeErr).Msgf("unable to remove files from cache: %s", removeErr)
						break
					}

					percentageInUse, err = s.cachePercentageInUse()
					if err != nil {
						s.log.Error().Err(err).Msgf("unable to get cache stats: %s", err)
						break
					}

					if percentageInUse < 95 {
						s.log.Info().Msg("finished running cache cleaner")
						break
					}
				}
			}

			s.log.Debug().Int("cacheInUse", percentageInUse).Msg("cache cleaner found percentage in use")

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				continue
			}
		}
	}()
}
