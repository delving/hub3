package oaipmh

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const tokenSep = "::"

var ErrResumptionTokenMalformed = errors.New("resumptionToken is not well-formed")

type RawToken struct {
	HarvestID    string
	Cursor       int
	StorePayload string // payload set by the store
}

func parseToken(input string) (RawToken, error) {
	parts := strings.SplitN(input, tokenSep, 3)
	if len(parts) != 3 {
		return RawToken{}, ErrBadResumptionToken
	}

	seen, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return RawToken{}, err
	}

	return RawToken{
		HarvestID:    parts[0],
		Cursor:       int(seen),
		StorePayload: parts[2],
	}, nil
}

func (tr *RawToken) String() string {
	return fmt.Sprintf("%s::%d::%s", tr.HarvestID, tr.Cursor, tr.StorePayload)
}
