package ead

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
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
	rdf "github.com/kiivihal/rdf2go"
	"github.com/rs/xid"
	"google.golang.org/protobuf/proto"
)

const (
	daoPath = "dao"
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

func (c *DaoClient) dropOrphans(cfg DaoConfig) error {
	m := &domainpb.IndexMessage{
		OrganisationID: cfg.OrgID,
		DatasetID:      cfg.ArchiveID,
		Revision: &domainpb.Revision{
			SHA:     cfg.UUID,
			Path:    cfg.RevisionKey,
			GroupID: cfg.InventoryID,
		},
		ActionType: domainpb.ActionType_DROP_ORPHANS,
	}

	// publish message
	if err := c.bi.Publish(context.Background(), m); err != nil {
		return err
	}

	return nil
}

func (c *DaoClient) PublishFindingAid(cfg DaoConfig) error {
	fa, err := cfg.FindingAid(c)
	if err != nil {
		return err
	}

	cfg.updateRevisionKey()

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

		if c.bi != nil {
			c.bi.Publish(context.Background(), m)
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

	if c.bi != nil {
		c.bi.Publish(context.Background(), m)
	}

	cfg.ObjectCount = int(fa.GetFileCount())
	cfg.MimeTypes = getMimeTypes(&fa)

	return c.dropOrphans(cfg)
}

func (c *DaoClient) StoreMets(cfg *DaoConfig) error {
	config.Config.Logger.Debug().
		Str("link", cfg.Link).
		Str("orgID", cfg.OrgID).
		Str("datasetID", cfg.ArchiveID).
		Str("InventoryID", cfg.InventoryID).
		Msg("storing remotes mets file")

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

	return ioutil.WriteFile(
		cfg.getMetsFilePath(),
		buf.Bytes(),
		os.ModePerm,
	)
}

func (c *DaoClient) DefaultDaoFn(cfg DaoConfig) error {
	if !strings.Contains(cfg.Link, "/gaf/api/mets/v1/") {
		return fmt.Errorf("invalid daolink to GAF: %s", cfg.Link)
	}

	return c.PublishFindingAid(cfg)
}

type DaoConfig struct {
	OrgID          string
	HubID          string
	ArchiveID      string // same as DatasetID
	ArchiveTitle   []string
	InventoryID    string
	InventoryPath  string
	InventoryTitle string
	UUID           string
	Link           string
	ObjectCount    int
	MimeTypes      []string
	RevisionKey    string
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
		Link:           tree.DaoLink,
		UUID:           getUUID(tree.DaoLink),
	}
}

func (cfg *DaoConfig) updateRevisionKey() {
	cfg.RevisionKey = xid.New().String()
}

func (cfg *DaoConfig) getMetsFilePath() string {
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
		log.Printf("mkdir error: %#v", err)
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

func (cfg *DaoConfig) metsFileExists() bool {
	_, err := os.Stat(cfg.getMetsFilePath())

	return !os.IsNotExist(err)
}

func (cfg *DaoConfig) Mets() (*Cmets, error) {
	return readMETS(cfg.getMetsFilePath())
}

func (cfg *DaoConfig) FindingAid(c *DaoClient) (eadpb.FindingAid, error) {
	if c.HttpFallback && !cfg.metsFileExists() {
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
		GroupID:       cfg.InventoryID,
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
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
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

	return triples
}

func (cfg *DaoConfig) fragmentGraph(file *eadpb.File) (*fragments.FragmentGraph, error) {
	subjectBase := fmt.Sprintf("%s/%s/archive/%s/%s", config.Config.RDF.BaseURL, cfg.OrgID, cfg.ArchiveID, cfg.InventoryID)
	id := fmt.Sprintf("%s-%s", cfg.InventoryID, file.Filename)
	header := createHeader(cfg, id, subjectBase, "mets")
	rm := fragments.NewEmptyResourceMap()

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
	// remove files because we don't want them to be stored
	fa.Files = []*eadpb.File{}

	subjectBase := fmt.Sprintf("%s/%s/archive/%s/%s", config.Config.RDF.BaseURL, cfg.OrgID, fa.ArchiveID, fa.InventoryID)
	id := fmt.Sprintf("%s-findingaid", fa.GetInventoryID())
	header := createHeader(cfg, id, subjectBase, "findingaid")

	rm := fragments.NewEmptyResourceMap()

	for idx, t := range findingAidTriples(header.EntryURI, fa) {
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

	metsPath := getMetsFilePath(archiveID, uuid)
	if _, err := os.Stat(metsPath); os.IsNotExist(err) {
		return cfg, ErrNoFileNotFound
	}

	r, err := os.Open(metsPath)
	if err != nil {
		return cfg, err
	}

	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
