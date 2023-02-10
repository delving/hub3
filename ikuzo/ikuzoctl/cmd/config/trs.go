package config

import (
	"os"

	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/trs"
)

type TRS struct {
	Path string
}

func (t *TRS) newService(cfg *Config) (*trs.Service, error) {
	dataBucket, err := createBucket(t.Path)
	if err != nil {
		cfg.log.Fatal().Err(err).Stack().Msg("unable to create databucket")
	}
	defer dataBucket.Close()

	svc, err := trs.NewService(
		trs.SetBlobBucket(dataBucket),
	)
	if err != nil {
		cfg.log.Fatal().Err(err).Stack().Msg("unable to create sync service")
	}

	return svc, nil
}

func (r *TRS) AddOptions(cfg *Config) error {
	if cfg.trs != nil {
		return nil
	}

	svc, err := r.newService(cfg)
	if err != nil {
		return err
	}

	cfg.trs = svc

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
	)

	return nil
}

func createBucket(path string) (*blob.Bucket, error) {
	// The directory you pass to fileblob.OpenBucket must exist first.
	if err := os.MkdirAll(path, 0o777); err != nil {
		return nil, err
	}

	// Create a file-based bucket.
	bucket, err := fileblob.OpenBucket(path, nil)
	if err != nil {
		return nil, err
	}
	// defer bucket.Close()
	return bucket, nil
}
