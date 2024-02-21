package bulk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pelletier/go-toml"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/service/x/sparql"
)

func (s *Service) scheduleTasks() error {
	slog.Info("setting up scheduled tasks for bulk")
	s.scheduler = gocron.NewScheduler(time.UTC)
	s.scheduler.SingletonModeAll()

	s.scheduler.Every(60).Seconds().Do(s.harvestSparqlDatasets)

	s.scheduler.StartAsync()
	return nil
}

func (s *Service) harvestSparqlDatasets() {
	configPath := strings.TrimSuffix(s.harvestConfigPath, "/")
	slog.Info("harvesting datasets", "configPath", configPath)

	matches, err := filepath.Glob(configPath + "/*.toml")
	if err != nil {
		slog.Error("unable to find configuration files", "path", configPath)
		return
	}

	for _, path := range matches {
		slog.Info("dataset to harvest", "path", path)
		if processErr := s.processSparqlDataset(path); processErr != nil {
			slog.Error("unable to process sparql dataset", "error", processErr)
		}
		time.Sleep(3 * time.Second)
	}
}

func (s *Service) processSparqlDataset(path string) error {
	cfg, err := decodeConfig(path)
	if err != nil {
		return err
	}

	if cfg.OrgID == "" || cfg.Spec == "" {
		return errors.New("orgID and Spec must be configured in the toml configuration")
	}

	ds, _, err := models.GetOrCreateDataSet(cfg.OrgID, cfg.Spec)
	if err != nil {
		return err
	}

	indexer := &graphIndexer{
		s:         s,
		ds:        ds,
		cfg:       cfg,
		startTime: time.Now(),
	}

	defer func() {
		cfg.LastCheck = time.Now()
		writeErr := writeConfig(path, cfg)
		if writeErr != nil {
			slog.Error("unable to write harvestConfig file", "path", path, "error", writeErr)
			return
		}
	}()

	if !indexer.hasChanges() {
		slog.Info("no changes for sparql dataset", "datasetID", cfg.Spec, "TargetDataSets", cfg.TargetDatasets)
		return nil
	}

	if incrementErr := indexer.incrementRevision(); incrementErr != nil {
		slog.Info("unable to increment dataset revision", "datasetID", cfg.Spec, "error", incrementErr)
		return incrementErr
	}

	harvestErr := sparql.HarvestGraphs(context.Background(), cfg, indexer.IndexGraph)
	if harvestErr != nil {
		slog.Info("unable to harvest sparql graphs", "datasetID", cfg.Spec, "error", harvestErr)
		return harvestErr
	}

	slog.Info(
		"started processing harvest dataset",
		"seen", indexer.seen, "total", indexer.cfg.TotalSizeSubjects,
		"max", indexer.cfg.MaxSubjects, "errors", len(indexer.cfg.HarvestErrors),
		"duration", prettyDuration(time.Since(indexer.startTime)),
		"spec", indexer.cfg.Spec,
	)

	if dropOrphansErr := indexer.dropOrphans(); dropOrphansErr != nil {
		slog.Info("unable to drop orphans sparql dataset", "datasetID", cfg.Spec, "error", dropOrphansErr)
		return dropOrphansErr
	}

	slog.Info(
		"finished processing harvest dataset",
		"seen", indexer.seen, "total", indexer.cfg.TotalSizeSubjects,
		"max", indexer.cfg.MaxSubjects, "errors", len(indexer.cfg.HarvestErrors),
		"duration", prettyDuration(time.Since(indexer.startTime)),
		"spec", indexer.cfg.Spec,
	)

	return nil
}

type graphIndexer struct {
	s         *Service
	ds        *models.DataSet
	cfg       *sparql.HarvestConfig
	seen      uint64
	startTime time.Time
}

func (gi *graphIndexer) incrementRevision() error {
	newDS, err := gi.ds.IncrementRevision()
	if err != nil {
		return err
	}
	gi.ds = newDS
	return nil
}

func (gi *graphIndexer) dropOrphans() error {
	m := &domainpb.IndexMessage{
		OrganisationID: gi.cfg.OrgID,
		DatasetID:      gi.cfg.Spec,
		Revision:       &domainpb.Revision{Number: int32(gi.ds.Revision)},
		ActionType:     domainpb.ActionType_DROP_ORPHANS,
	}

	if err := gi.s.index.Publish(context.Background(), m); err != nil {
		return err
	}

	return nil
}

func (gi *graphIndexer) hasChanges() (hasChange bool) {
	updated := make(map[string]int, len(gi.cfg.TargetDatasets))
	for targetSpec, revision := range gi.cfg.TargetDatasets {
		ds, err := models.GetDataSet(gi.cfg.OrgID, targetSpec)
		if err != nil {
			hasChange = true
			updated[targetSpec] = 0
			continue
		}

		if ds.InProgress {
			// when any of the datasets is InProgress we cannot harvest
			return false
		}

		if ds.Revision != revision {
			hasChange = true
		}
		updated[targetSpec] = ds.Revision
	}

	if hasChange {
		gi.cfg.TargetDatasets = updated
		return true
	}

	return false
}

func getLocalID(subject string) string {
	parts := strings.Split(subject, "/")
	return parts[len(parts)-1]
}

func (gi *graphIndexer) IndexGraph(g *rdf.Graph) error {
	if g == nil {
		return fmt.Errorf("cannot process nil *rdf.Graph")
	}

	// slog.Info("processing subject", "subject", g.Subject)

	fg := fragments.NewFragmentGraph()
	fg.Meta.OrgID = gi.cfg.OrgID
	fg.Meta.HubID = fmt.Sprintf(
		"%s_%s_%s",
		gi.cfg.OrgID, gi.cfg.Spec, getLocalID(g.Subject.RawValue()),
	)
	fg.Meta.Spec = gi.cfg.Spec
	fg.Meta.Revision = int32(gi.ds.Revision)
	fg.Meta.Modified = fragments.NowInMillis()
	fg.Meta.NamedGraphURI = fmt.Sprintf("%s/graph", g.Subject.RawValue())
	fg.Meta.EntryURI = fg.GetAboutURI()
	fg.Meta.Tags = gi.cfg.Tags

	if strings.HasSuffix(fg.Meta.HubID, "_") {
		return fmt.Errorf("invalid hubID %s extracted from subject %s", fg.Meta.HubID, g.Subject.String())
	}

	if fg.Meta.HasTag("nk") {
		// TODO: manipulate the graph for
		_ = fg
	}

	fb := fragments.NewFragmentBuilder(fg)

	var err error
	fb.Graph, err = g.AsLegacyGraph()
	if err != nil {
		return err
	}

	fb.ResourceMap()
	_ = fb.Doc()

	processErr := processV2(context.Background(), fb, gi.s.index)
	if processErr != nil {
		return processErr
	}

	atomic.AddUint64(&gi.seen, 1)
	if gi.seen%100 == 0 {
		slog.Info(
			"harvesting progress",
			"seen", gi.seen, "total", gi.cfg.TotalSizeSubjects,
			"max", gi.cfg.MaxSubjects, "errors", len(gi.cfg.HarvestErrors),
			"duration", prettyDuration(time.Since(gi.startTime)),
			"spec", gi.cfg.Spec,
		)
	}

	return nil
}

func writeConfig(path string, cfg *sparql.HarvestConfig) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to find configuration; %w", err)
	}

	defer f.Close()

	if encodeErr := toml.NewEncoder(f).Encode(cfg); encodeErr != nil {
		return fmt.Errorf("unable to write configuration to path %q; %w", path, encodeErr)
	}

	return nil
}

func decodeConfig(path string) (cfg *sparql.HarvestConfig, err error) {
	f, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("unable to find configuration; %w", err)
	}
	defer f.Close()

	var config sparql.HarvestConfig
	decodeErr := toml.NewDecoder(f).Decode(&config)
	if decodeErr != nil {
		return cfg, fmt.Errorf("unable to decode %s; %w", path, decodeErr)
	}

	return &config, nil
}

func prettyDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	var result string
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 || result == "" {
		result += fmt.Sprintf("%ds", seconds)
	}

	return result
}
