package imageproxy

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
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

	cmd := exec.Command("find", s.cacheDir, "-type", "f", "-printf", "%T+ %p\n")

	log.Printf("%s", cmd.String())

	out, err := cmd.Output()
	if err != nil {
		log.Printf("%s\n%s", cmd.String(), out)
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	sort.Strings(lines)

	if len(lines) > nr {
		lines = lines[:nr]
	}

	return s.removeCacheFiles(lines)
}

// removeCacheFiles removes the files listed as "<mtime> <path>" lines.
// Deleting a deepzoom tile on its own leaves a pyramid that is missing
// tiles but still looks cached, so a tile takes its whole pyramid
// (the "<base>_files" directory and the "<base>.dzi" marker) with it.
func (s *Service) removeCacheFiles(lines []string) error {
	removedPyramids := map[string]bool{}

	for _, line := range lines {
		_, p, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}

		base, _, isTile := strings.Cut(p, "_files/")
		if isTile {
			if !removedPyramids[base] {
				removedPyramids[base] = true

				if err := os.RemoveAll(base + "_files"); err != nil {
					log.Printf("unable to remove pyramid %s: %s", base, err)
				}

				if err := os.Remove(base + deepZoomSuffix); err != nil && !errors.Is(err, os.ErrNotExist) {
					log.Printf("unable to remove dzi %s: %s", base, err)
				}
			}
			// the pyramid removal normally covers p; remove directly if not,
			// so the eviction loop always makes progress
			if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Printf("unable to remove cache file %s: %s", p, err)
			}

			continue
		}

		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("unable to remove cache file %s: %s", p, err)
		}
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
