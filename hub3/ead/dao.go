package ead

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/ead/eadpb"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/go-chi/chi"
	rdf "github.com/kiivihal/rdf2go"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type DaoClient struct {
	bi           *index.Service
	client       *http.Client
	HttpFallback bool // retrieve DAO url if not present locally
}

func NewDaoClient(bi *index.Service) DaoClient {
	return DaoClient{
		client: &http.Client{Timeout: 10 * time.Second},
		bi:     bi,
	}
}

// GetDaoConfig convenience function to retrieve the DaoConfig
func (c *DaoClient) GetDaoConfig(archiveID, uuid string) (DaoConfig, error) {
	return GetDaoConfig(archiveID, uuid)
}

func (c *DaoClient) GetDigitalObjectCount(archiveID string) (int, error) {
	var digitalObjects int

	files, err := ioutil.ReadDir(getMetsDirPath(archiveID))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0, err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		uuid := strings.TrimSuffix(f.Name(), ".json")

		cfg, err := c.GetDaoConfig(archiveID, uuid)
		if err != nil {
			return 0, err
		}

		mets, err := cfg.Mets()
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return 0, err
		}

		metsRef, err := mets.extractFiles()
		if err != nil {
			return 0, err
		}

		digitalObjects += len(metsRef)
	}

	return digitalObjects, nil
}

func (c *DaoClient) dropOrphans(cfg *DaoConfig) error {
	m := &domainpb.IndexMessage{
		OrganisationID: cfg.OrgID,
		DatasetID:      cfg.ArchiveID,
		Revision: &domainpb.Revision{
			SHA:     cfg.RevisionKey,
			Path:    fmt.Sprintf("mets/%s", cfg.UUID),
			GroupID: cfg.InventoryPath,
		},
		ActionType: domainpb.ActionType_DROP_ORPHANS,
	}

	// publish message
	if c.bi != nil {
		if err := c.bi.Publish(context.Background(), m); err != nil {
			return err
		}
	}

	return nil
}

func (c *DaoClient) PublishFiles(cfg *DaoConfig, fa *eadpb.FindingAid) error {
	cfg.updateRevisionKey()

	err := validateFindingAid(fa)
	if err != nil {
		return err
	}

	for _, file := range fa.Files {
		fg, err := cfg.fragmentGraph(file)
		if err != nil {
			return err
		}
		fg.Meta.SourceID = cfg.RevisionKey

		m, err := fg.IndexMessage()
		if err != nil {
			return err
		}

		m.IndexType = domainpb.IndexType_DIGITAL_OBJECTS

		if c.bi != nil {
			c.bi.Publish(context.Background(), m)
		}
	}
	return nil
}

func (c *DaoClient) PublishFindingAid(cfg *DaoConfig, excludeMetsFiles ...bool) error {
	fa, err := cfg.FindingAid(c)
	if err != nil {
		return err
	}

	var excludeFiles bool
	if len(excludeMetsFiles) > 0 && excludeMetsFiles[0] {
		excludeFiles = true
	}

	if !excludeFiles {
		err = c.PublishFiles(cfg, &fa)
		if err != nil {
			return err
		}
	}

	fg, err := cfg.findingAidFragmentGraph(&fa)
	if err != nil {
		return err
	}

	fg.Meta.SourceID = cfg.RevisionKey

	m, err := fg.IndexMessage()
	if err != nil {
		return err
	}

	m.IndexType = domainpb.IndexType_DIGITAL_OBJECTS

	if c.bi != nil {
		c.bi.Publish(context.Background(), m)
	}

	if !excludeFiles {
		cfg.ObjectCount = int(fa.GetFileCount())
		cfg.MimeTypes = getMimeTypes(&fa)
		return c.dropOrphans(cfg)
	}

	return nil
}

func validateFindingAid(fa *eadpb.FindingAid) error {
	duplicateFilenamesErr := assertUniqueFilenames(fa.Files)
	if duplicateFilenamesErr != nil {
		return duplicateFilenamesErr
	}

	return nil
}

func assertUniqueFilenames(files []*eadpb.File) error {
	fileNames := make(map[string]int32)

	for _, file := range files {
		_, exists := fileNames[file.Filename]
		if exists {
			return errors.New(fmt.Sprintf("duplicate filename found: %s", file.Filename))
		}

		fileNames[file.Filename] = file.SortKey
	}

	return nil
}

func (c *DaoClient) StoreMets(cfg *DaoConfig) error {
	log.Debug().
		Str("link", cfg.Link).
		Str("orgID", cfg.OrgID).
		Str("datasetID", cfg.ArchiveID).
		Str("InventoryID", cfg.InventoryID).
		Msg("storing remote mets file")

	resp, err := c.client.Get(cfg.Link)
	if err != nil {
		metsRetrieveErr := fmt.Errorf("unable to retrieve METS %s client error: %s", cfg.Link, err)
		logMETSError(cfg.ArchiveID, cfg.InventoryID, metsRetrieveErr.Error())

		return metsRetrieveErr
	}

	if resp.StatusCode != http.StatusOK {
		metsStatusErr := fmt.Errorf("unable to retrieve METS %s HTTP status error: %d", cfg.Link, resp.StatusCode)
		logMETSError(cfg.ArchiveID, cfg.InventoryID, metsStatusErr.Error())

		return metsStatusErr
	}

	defer resp.Body.Close()

	mets, err := metsParse(resp.Body)
	if err != nil {
		return err
	}

	mets.CmetsHdr.AttrCREATEDATE = ""
	mets.CmetsHdr.AttrLASTMODDATE = ""

	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "\t")

	if err := enc.Encode(mets); err != nil {
		return err
	}

	if _, err := os.Stat(cfg.GetMetsFilePath()); os.IsNotExist(err) {
		if mkDirErr := os.MkdirAll(cfg.getDirPath(), os.ModePerm); mkDirErr != nil {
			return mkDirErr
		}
	}

	return ioutil.WriteFile(
		cfg.GetMetsFilePath(),
		buf.Bytes(),
		os.ModePerm,
	)
}

func (c *DaoClient) DefaultDaoFn(cfg *DaoConfig) error {
	if !strings.Contains(cfg.Link, "/gaf/api/mets/v1/") {
		return fmt.Errorf("invalid daolink to GAF: ' %s '", cfg.Link)
	}

	return c.PublishFindingAid(cfg)
}

func validateMetsRequest(r *http.Request) (string, string, error) {
	spec := chi.URLParam(r, "spec")
	err := ValidateSpec(spec)
	if err != nil {
		return "", "", err
	}

	uuid := chi.URLParam(r, "UUID")
	if uuid == "" {
		return "", "", fmt.Errorf("UUID cannot be empty")
	}

	return spec, uuid, nil
}

// Index published records from stored METS files identified by their UUID
func (c *DaoClient) Index(w http.ResponseWriter, r *http.Request) {
	spec, uuid, err := validateMetsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := c.GetDaoConfig(spec, uuid)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			http.Error(w, "unknown UUID", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}

	if err := c.PublishFindingAid(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete indexes the stored METS files identified by their UUID
func (c *DaoClient) Delete(archiveID, uuid string) error {
	cfg, err := c.GetDaoConfig(archiveID, uuid)
	if errors.Is(err, ErrFileNotFound) {
		return nil
	}

	cfg.RevisionKey = "1"

	if err := c.dropOrphans(&cfg); err != nil {
		return err
	}

	if err := cfg.Delete(); err != nil {
		return err
	}

	return nil
}

func (c *DaoClient) HandleDelete(w http.ResponseWriter, r *http.Request) {
	spec, uuid, err := validateMetsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deleteErr := c.Delete(spec, uuid)
	if deleteErr != nil {
		http.Error(w, deleteErr.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// DownloadConfig is a handler that returns a stored METS XML for an inventory.
func (c *DaoClient) DownloadConfig(w http.ResponseWriter, r *http.Request) {
	spec, uuid, err := validateMetsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cfgPath := path.Join(config.Config.EAD.CacheDir, spec, "mets", uuid+".json")
	http.ServeFile(w, r, cfgPath)
}

// DownloadXML is a handler that returns a stored METS XML for an inventory.
func (c *DaoClient) DownloadXML(w http.ResponseWriter, r *http.Request) {
	spec, uuid, err := validateMetsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	metsPath := path.Join(config.Config.EAD.CacheDir, spec, "mets", uuid)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.xml", spec, uuid))
	w.Header().Set("Content-Type", "application/xml")
	http.ServeFile(w, r, metsPath)
}

const (
	SourceEad    = "EAD"
	SourceOaiPmh = "OAI-PMH"
)

type DaoConfig struct {
	OrgID          string
	HubID          string
	ArchiveID      string // same as DatasetID
	ArchiveTitle   []string
	InventoryID    string
	InventoryPath  string
	InventoryTitle string
	UUID           string // dUUID
	Link           string
	ObjectCount    int
	MimeTypes      []string
	RevisionKey    string
	FilterTypes    []string
	PeriodDesc     []string
	Filenames      []string // names of the files within the digital object
	FileUUIDs      []string // uuids of the files within the digital object
	Source         string   // source of the METS file (i.e. 'EAD' or 'OAI-PMH')
}

func getUUID(daoLink string) string {
	parts := strings.Split(daoLink, "/")
	return parts[len(parts)-1]
}

func newDaoConfig(cfg *NodeConfig, tree *fragments.Tree) DaoConfig {
	return DaoConfig{
		OrgID:          cfg.OrgID,
		HubID:          tree.HubID,
		ArchiveID:      cfg.Spec,
		ArchiveTitle:   cfg.Title,
		InventoryID:    tree.UnitID,
		InventoryPath:  tree.CLevel,
		InventoryTitle: tree.Label,
		UUID:           getUUID(tree.DaoLink),
		PeriodDesc:     cfg.PeriodDesc,
		Link:           tree.DaoLink,
		FileUUIDs:      make([]string, 0),
		Filenames:      make([]string, 0),
	}
}

func (cfg *DaoConfig) updateRevisionKey() {
	cfg.RevisionKey = xid.New().String()
}

func (cfg *DaoConfig) GetMetsFilePath() string {
	return getMetsFilePath(cfg.ArchiveID, cfg.UUID)
}

func (cfg *DaoConfig) getConfigPath() string {
	return getDaoConfigPath(cfg.ArchiveID, cfg.UUID)
}

func (cfg *DaoConfig) getDirPath() string {
	return getMetsDirPath(cfg.ArchiveID)
}

func (cfg *DaoConfig) Write() error {
	err := os.MkdirAll(cfg.getDirPath(), os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("mkdir error")
		return err
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		cfg.getConfigPath(),
		b,
		os.ModePerm,
	)
}

// Delete removes the DaoConfig and METS file
func (cfg *DaoConfig) Delete() error {
	files := []string{
		cfg.GetMetsFilePath(),
		getDaoConfigPath(cfg.ArchiveID, cfg.UUID),
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return err
		}
	}

	return nil
}

func (cfg *DaoConfig) hasOrphanedMetsFile() bool {
	if cfg.MetsFileExists() && !cfg.daoConfigExists() {
		return true
	}

	return false
}

func (cfg *DaoConfig) daoConfigExists() bool {
	_, err := os.Stat(cfg.getConfigPath())

	return !os.IsNotExist(err)
}

func (cfg *DaoConfig) MetsFileExists() bool {
	_, err := os.Stat(cfg.GetMetsFilePath())

	return !os.IsNotExist(err)
}

func (cfg *DaoConfig) Mets() (*Cmets, error) {
	return readMETS(cfg.GetMetsFilePath())
}

func (cfg *DaoConfig) FindingAid(c *DaoClient) (eadpb.FindingAid, error) {
	if c.HttpFallback && !cfg.MetsFileExists() {
		if err := c.StoreMets(cfg); err != nil {
			return eadpb.FindingAid{}, err
		}
	}

	mets, err := cfg.Mets()
	if err != nil {
		return eadpb.FindingAid{}, err
	}

	return mets.newFindingAid(cfg)
}

func createHeader(cfg *DaoConfig, id, subjectBase, tag string) *fragments.Header {
	subject := fmt.Sprintf("%s/%s", subjectBase, id)
	header := &fragments.Header{
		OrgID: cfg.OrgID,
		Spec:  cfg.ArchiveID,
		HubID: fmt.Sprintf(
			"%s_%s_%s",
			cfg.OrgID,
			cfg.ArchiveID,
			strings.ReplaceAll(id, "/", "-"),
		),
		DocType:       fragments.FragmentGraphDocType,
		EntryURI:      subject,
		NamedGraphURI: fmt.Sprintf("%s/graph", subject),
		SourcePath:    "mets/" + cfg.UUID,
		GroupID:       cfg.InventoryPath,
		Modified:      fragments.NowInMillis(),
		Tags:          []string{tag},
	}

	return header
}

func (cfg *DaoConfig) getArchiveTitle() string {
	if len(cfg.ArchiveTitle) == 0 || cfg.ArchiveTitle[0] == "" {
		return ""
	}

	return cfg.ArchiveTitle[0]
}

func (cfg *DaoConfig) fileTriples(subject string, file *eadpb.File) []*rdf.Triple {
	s := rdf.NewResource(subject)
	triples := []*rdf.Triple{
		rdf.NewTriple(
			s,
			rdf.NewResource(fragments.RDFType),
			rdf.NewResource("https://archief.nl/def/ead/mets/File"),
		),
		rdf.NewTriple(
			s,
			rdf.NewResource("http://www.w3.org/2000/01/rdf-schema#label"),
			rdf.NewLiteral(file.Filename),
		),
	}
	t := func(s rdf.Term, p, o string, oType convert) {
		triple := addNonEmptyTriple(s, p, o, oType)
		if triple != nil {
			triples = append(triples, triple)
		}
	}

	t(s, "fileName", file.Filename, rdf.NewLiteral)
	t(s, "fileSize", string(file.FileSize), rdf.NewLiteral)
	t(s, "mimeType", file.MimeType, rdf.NewLiteral)
	t(s, "file-uuid", file.Fileuuid, rdf.NewLiteral)
	t(s, "order", string(file.SortKey), rdf.NewLiteral)
	t(s, "duuid", cfg.UUID, rdf.NewLiteral)
	t(s, "archiveID", cfg.ArchiveID, rdf.NewLiteral)
	t(s, "archiveTitle", cfg.getArchiveTitle(), rdf.NewLiteral)
	t(s, "inventoryID", cfg.InventoryID, rdf.NewLiteral)
	t(s, "inventoryTitle", cfg.InventoryTitle, rdf.NewLiteral)

	// genreform based filtering
	for _, filter := range cfg.FilterTypes {
		t(s, "filterType", filter, rdf.NewLiteral)
	}

	return triples
}

func (cfg *DaoConfig) fragmentGraph(file *eadpb.File) (*fragments.FragmentGraph, error) {
	subjectBase := fmt.Sprintf("%s/%s/archive/%s/%s", config.Config.RDF.BaseURL, cfg.OrgID, cfg.ArchiveID, cfg.InventoryID)
	id := fmt.Sprintf("%s-%s", cfg.InventoryID, file.Filename)
	header := createHeader(cfg, id, subjectBase, "mets")
	rm := fragments.NewEmptyResourceMap(cfg.OrgID)

	for idx, t := range cfg.fileTriples(header.EntryURI, file) {
		if err := rm.AppendOrderedTriple(t, false, idx); err != nil {
			return nil, err
		}
	}

	fg := fragments.NewFragmentGraph()
	fg.Meta = header
	fg.Tree = &fragments.Tree{
		CLevel:      cfg.InventoryPath,
		UnitID:      cfg.InventoryID,
		MimeTypes:   []string{file.MimeType},
		SortKey:     uint64(file.SortKey),
		Title:       cfg.InventoryTitle,
		InventoryID: cfg.ArchiveID,
		Label:       file.Filename,
	}

	b, err := proto.Marshal(file)
	if err != nil {
		return fg, fmt.Errorf("unable to marshal protobuf message: %#v", err)
	}

	fg.ProtoBuf = &fragments.ProtoBuf{
		MessageType: "eadpb.File",
		Data:        fmt.Sprintf("%x", b),
	}
	fg.SetResources(rm)

	return fg, nil
}

func (cfg *DaoConfig) findingAidFragmentGraph(fa *eadpb.FindingAid) (*fragments.FragmentGraph, error) {
	subjectBase := fmt.Sprintf("%s/%s/archive/%s/%s", config.Config.RDF.BaseURL, cfg.OrgID, fa.ArchiveID, fa.InventoryID)
	id := fmt.Sprintf("%s-findingaid", fa.GetInventoryID())
	header := createHeader(cfg, id, subjectBase, "findingaid")

	rm := fragments.NewEmptyResourceMap(cfg.OrgID)

	for idx, t := range findingAidTriples(header.EntryURI, fa, cfg) {
		if err := rm.AppendOrderedTriple(t, false, idx); err != nil {
			return nil, err
		}
	}

	fg := fragments.NewFragmentGraph()
	fg.Meta = header
	fg.Tree = &fragments.Tree{
		CLevel:      fa.InventoryPath,
		UnitID:      fa.InventoryID,
		Title:       fa.InventoryTitle,
		InventoryID: fa.ArchiveID,
		Label:       "findingaid",
	}

	// remove files because we don't want them to be stored
	fa.Files = []*eadpb.File{}

	b, err := proto.Marshal(fa)
	if err != nil {
		return fg, fmt.Errorf("unable to marshal protobuf message: %#v", err)
	}

	fg.ProtoBuf = &fragments.ProtoBuf{
		MessageType: "eadpb.FindingAid",
		Data:        fmt.Sprintf("%x", b),
	}

	fg.SetResources(rm)

	return fg, nil
}

func getMetsFilePath(archiveID, uuid string) string {
	return path.Join(config.Config.EAD.CacheDir, archiveID, "mets", uuid)
}

func getDaoConfigPath(archiveID, uuid string) string {
	return getMetsFilePath(archiveID, uuid) + ".json"
}

func getMetsDirPath(archiveID string) string {
	return path.Join(config.Config.EAD.CacheDir, archiveID, "mets")
}

func GetDaoConfig(archiveID, uuid string) (DaoConfig, error) {
	var cfg DaoConfig

	daoPath := getDaoConfigPath(archiveID, uuid)
	if _, err := os.Stat(daoPath); os.IsNotExist(err) {
		return cfg, ErrFileNotFound
	}

	f, err := os.Open(daoPath)
	if err != nil {
		return cfg, err
	}

	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
