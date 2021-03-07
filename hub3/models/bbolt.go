package models

// import (
// "os"

// "github.com/asdine/storm"
// "github.com/delving/hub3/config"
// "github.com/rs/zerolog/log"
// )

// var (
// orm *storm.DB
// )

// func ORM() *storm.DB {
// if orm == nil {
// ResetStorm()
// }

// return orm
// }

// func newDB(path string) (*storm.DB, error) {
// if path == "" {
// path = "hub3.db"
// }

// return storm.Open(path)
// }

// func ResetStorm() {
// var err error

// orm, err = newDB("")
// if err != nil {
// log.Fatal().Err(err).Msg("unable to open bbolt storm ORM.")
// }
// }

// func ResetEADCache() {
// cacheDir := config.Config.EAD.CacheDir
// os.RemoveAll(cacheDir)
// os.MkdirAll(cacheDir, os.ModePerm)
// }
