package trs

import "gocloud.dev/blob"

type Option func(*Service) error

//	func SetStore(store Store) Option {
//		return func(s *Service) error {
//			s.store = store
//			return nil
//		}
//	}

func SetBlobBucket(bucket *blob.Bucket) Option {
	return func(s *Service) error {
		s.bucket = blob.PrefixedBucket(bucket, "trs/")
		return nil
	}
}
