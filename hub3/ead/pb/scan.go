package pb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// NewPager creates a new Pager with all navigation settings.
func NewPager(total int64, request *ViewRequest) (Pager, error) {
	pager := Pager{
		HasNext:        false,
		HasPrevious:    false,
		TotalCount:     int32(total),
		NrPages:        0,
		PageCurrent:    request.GetPage(),
		PageNext:       0,
		PagePrevious:   0,
		PageSize:       request.GetPageSize(),
		ActiveFilename: request.GetFilename(),
		ActiveSortKey:  int32(request.GetSortKey()),
		Paging:         request.GetPaging(),
	}
	if pager.TotalCount > 0 && pager.PageSize != 0 {
		pager.NrPages = (pager.TotalCount / pager.PageSize)
		if pager.TotalCount%pager.PageSize != 0 {
			pager.NrPages++
		}
	}

	pager.setPaging()

	return pager, nil
}

func (p *Pager) setPaging() {
	// reset paging values
	if p.HasNext || p.PageNext != int32(0) {
		p.HasNext = false
		p.PageNext = int32(0)
	}

	if p.HasPrevious || p.PagePrevious != int32(0) {
		p.HasPrevious = false
		p.PagePrevious = int32(0)
	}

	// calculate next
	if p.NrPages > p.PageCurrent {
		p.HasNext = true
		p.PageNext = p.PageCurrent + 1
	}

	// calculate previous
	if p.PageCurrent > 1 {
		p.HasPrevious = true
		p.PagePrevious = p.PageCurrent - 1
	}
}

// SetPage takes a sortKey and gives back on which page this can be found.
func (p *Pager) SetPage(sortKey int32) {
	p.ActiveSortKey = sortKey
	p.PageCurrent = sortKey / p.GetPageSize()

	if sortKey%p.GetPageSize() != 0 {
		p.PageCurrent++
	}

	p.setPaging()
}

type pbWrapper struct {
	Protobuf *fragments.ProtoBuf `json:"protobuf"`
}

func unmarshalWrapper(hit json.RawMessage) (*pbWrapper, error) {
	fw := new(pbWrapper)
	err := json.Unmarshal(hit, &fw)

	return fw, err
}

// DecodePBFile decodes a pb.File from an json.RawMessage.
// This is used to parse the protobuf from an ElasticSearch response.
func DecodePBFile(hit json.RawMessage) (*File, error) {
	fw, err := unmarshalWrapper(hit)
	if err != nil {
		return nil, err
	}

	if fw.Protobuf.MessageType != "pb.File" {
		return nil, fmt.Errorf("hit from ElasticSearch contains wrong messageType: %s", fw.Protobuf.MessageType)
	}

	var file File

	b, err := hex.DecodeString(fw.Protobuf.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode hex string")
	}

	err = proto.Unmarshal(b, &file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal proto data")
	}

	return &file, nil
}

// DecodePBFindingAid decodes a pb.FindingAid from an json.RawMessage.
// This is used to parse the protobuf from an ElasticSearch response.
func DecodePBFindingAid(hit json.RawMessage) (*FindingAid, error) {
	fw, err := unmarshalWrapper(hit)
	if err != nil {
		return nil, err
	}

	if fw.Protobuf.MessageType != "pb.FindingAid" {
		return nil, fmt.Errorf("hit from ElasticSearch contains wrong messageType: %s", fw.Protobuf.MessageType)
	}

	b, err := hex.DecodeString(fw.Protobuf.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode hex string")
	}

	var fa FindingAid
	if err := proto.Unmarshal(b, &fa); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal proto data")
	}

	return &fa, nil
}
