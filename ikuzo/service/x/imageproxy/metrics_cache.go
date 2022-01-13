package imageproxy

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

type CacheSize struct {
	Count       uint64
	SizeInBytes uint64
}

type CacheMetrics struct {
	SourceFiles CacheSize `json:"sourceFiles"`
	Thumbnails  CacheSize `json:"thumbnails"`
	DeepZoom    struct {
		Descriptions CacheSize `json:"descriptions"`
		Tiles        CacheSize `json:"tiles"`
	} `json:"deepZoom"`
	Total CacheSize `json:"total,omitempty"`
}

func (s *Service) CacheMetrics() CacheMetrics {
	cm := s.cm
	cm.Total.Count = cm.TotalFiles()
	cm.Total.SizeInBytes = cm.TotalSizeInBytes()

	return cm
}

func (cm CacheMetrics) TotalFiles() uint64 {
	return cm.SourceFiles.Count +
		cm.Thumbnails.Count +
		cm.DeepZoom.Descriptions.Count +
		cm.DeepZoom.Tiles.Count
}

func (cm CacheMetrics) TotalSizeInBytes() uint64 {
	return cm.SourceFiles.SizeInBytes +
		cm.Thumbnails.SizeInBytes +
		cm.DeepZoom.Descriptions.SizeInBytes +
		cm.DeepZoom.Tiles.SizeInBytes
}

func newCacheMetrics() CacheMetrics {
	cm := CacheMetrics{
		SourceFiles: CacheSize{},
		Thumbnails:  CacheSize{},
	}
	cm.DeepZoom.Descriptions = CacheSize{}
	cm.DeepZoom.Tiles = CacheSize{}

	return cm
}

func (cm *CacheMetrics) addSourceFile(size int64) {
	atomic.AddUint64(&cm.SourceFiles.Count, 1)
	atomic.AddUint64(&cm.SourceFiles.SizeInBytes, uint64(size))
}

func (cm *CacheMetrics) removeSourceFile(size int64) {
	atomic.AddUint64(&cm.SourceFiles.Count, -uint64(0))
	atomic.AddUint64(&cm.SourceFiles.SizeInBytes, -uint64(size))
}

func (cm *CacheMetrics) addThumbnail(size int64) {
	atomic.AddUint64(&cm.Thumbnails.Count, 1)
	atomic.AddUint64(&cm.Thumbnails.SizeInBytes, uint64(size))
}

func (cm *CacheMetrics) removeThumbnail(size int64) {
	atomic.AddUint64(&cm.Thumbnails.Count, -uint64(0))
	atomic.AddUint64(&cm.Thumbnails.SizeInBytes, -uint64(size))
}

func (cm *CacheMetrics) addDeepZoom(size int64) {
	atomic.AddUint64(&cm.DeepZoom.Descriptions.Count, 1)
	atomic.AddUint64(&cm.DeepZoom.Descriptions.SizeInBytes, uint64(size))
}

func (cm *CacheMetrics) removeDeepZoom(size int64) {
	atomic.AddUint64(&cm.DeepZoom.Descriptions.Count, -uint64(0))
	atomic.AddUint64(&cm.DeepZoom.Descriptions.SizeInBytes, -uint64(size))
}

func (cm *CacheMetrics) addDeepZoomTiles(tiles int, size int64) {
	atomic.AddUint64(&cm.DeepZoom.Tiles.Count, uint64(tiles))
	atomic.AddUint64(&cm.DeepZoom.Tiles.SizeInBytes, uint64(size))
}

func (cm *CacheMetrics) removeDeepZoomTiles(tiles int, size int64) {
	atomic.AddUint64(&cm.DeepZoom.Descriptions.Count, -uint64(tiles))
	atomic.AddUint64(&cm.DeepZoom.Descriptions.SizeInBytes, -uint64(size))
}

func (s *Service) buildCacheMetrics() error {
	sizes := make(chan int64)

	s.cm = newCacheMetrics()

	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil || file == nil {
			return nil // Ignore errors
		}

		if err := s.updateCacheMetrics(path, file, false); err != nil {
			return err
		}

		if file.IsDir() && strings.HasSuffix(file.Name(), "_files") {
			return filepath.SkipDir
		}

		return nil
	}

	go func() {
		if err := filepath.Walk(s.cacheDir, readSize); err != nil {
			s.log.Error().Err(err).Msgf("unable to build cache: %s", err)
		}

		close(sizes)
	}()

	size := int64(0)
	for s := range sizes {
		size += s
	}

	return nil
}

// updateCacheMetrics add or decrement all source and derivatives for the given path.
func (s *Service) updateCacheMetrics(path string, file os.FileInfo, removed bool) error {
	sourcePath := file.Name()

	var fn func(size int64)

	switch {
	case strings.HasSuffix(sourcePath, "="):
		if !removed {
			fn = s.cm.addSourceFile
		} else {
			fn = s.cm.removeSourceFile
		}
	case strings.HasSuffix(sourcePath, ".dzi"):
		if !removed {
			fn = s.cm.addDeepZoom
		} else {
			fn = s.cm.removeDeepZoom
		}
	case strings.HasSuffix(sourcePath, "_tn.jpg"):
		if !removed {
			fn = s.cm.addThumbnail
		} else {
			fn = s.cm.removeThumbnail
		}
	case strings.HasSuffix(sourcePath, "_files"):
		tiles, size := s.countTiles(path)
		if !removed {
			s.cm.addDeepZoomTiles(tiles, size)
		} else {
			s.cm.removeDeepZoomTiles(tiles, size)
		}

		return nil
	default:
		return nil
	}

	if !file.IsDir() {
		fn(file.Size())
	}

	return nil
}

func (s *Service) countTiles(sourcePath string) (int, int64) {
	sizes := make(chan int64)

	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil || file == nil {
			return nil // Ignore errors
		}

		if !file.IsDir() && strings.HasSuffix(file.Name(), ".jpeg") {
			sizes <- file.Size()
		}

		return nil
	}

	go func() {
		if err := filepath.Walk(sourcePath, readSize); err != nil {
			s.log.Error().Err(err).Msgf("unable to build cache: %s", err)
		}
		close(sizes)
	}()

	var files int

	size := int64(0)

	for s := range sizes {
		size += s
		files++
	}

	return files, size
}
