package ead

// Description is simplified version of the 'eadheader', 'archdesc/did' and
// 'archdesc/descgroups'.
//
// The goal of the simplification is reduce the complexity of the Archival Description
// for searching and rendering without loosing semantic meaning.
// proteus:generate
type Description struct {
	Summary Summary `json:"summary,omitempty"`
}

// Summary holds the essential metadata information to describe an Archive.
type Summary struct {
	FindingAid *FindingAid `json:"findingAid,omitempty"`
	File       *File       `json:"file,omitempty"`
	Profile    *Profile    `json:"profile,omitempty"`
}

// FindingAid holds the core information about the Archival Record
type FindingAid struct {
	ID         string   `json:"id,omitempty"`
	Country    string   `json:"country,omitempty"`
	AgencyCode string   `json:"agencyCode,omitempty"`
	Title      []string `json:"title,omitempty"`
	ShortTitle string   `json:"shortTitle,omitempty"`
	Unit       *Unit    `json:"unit,omitempty"`
}

// Unit holds the meta information of the Archival Record
type Unit struct {
	Date             []string `json:"date,omitempty"`
	DateBulk         string   `json:"dateBulk"`
	ID               string   `json:"id,omitempty"`
	Physical         string   `json:"physical,omitempty"`
	Files            string   `json:"files,omitempty"`
	Length           string   `json:"length,omitempty"`
	Language         string   `json:"language,omitempty"`
	Material         string   `json:"material,omitempty"`
	Repository       string   `json:"repository,omitempty"`
	PhysicalLocation string   `json:"physicalLocation,omitempty"`
	Origin           string   `json:"origin,omitempty"`
	Abstract         []string `json:"abstract,omitempty"`
}

// File holds the meta-information about the EAD file
type File struct {
	Title           string   `json:"title,omitempty"`
	Author          string   `json:"author,omitempty"`
	Edition         []string `json:"edition,omitempty"`
	Publisher       string   `json:"publisher,omitempty"`
	PublicationDate string   `json:"publicationDate,omitempty"`
	Copyright       string   `json:"copyright,omitempty"`
	CopyrightURI    string   `json:"copyrightURI"`
}

// Profile details information about the creation of the Archival Record
type Profile struct {
	Creation string `json:"creation,omitempty"`
	Language string `json:"language,omitempty"`
}

// NewDescription creates an Description from a Cead object.
func NewDescription(ead *Cead) (*Description, error) {
	desc := new(Description)
	if ead.Ceadheader != nil {
		desc.Summary.Profile = newProfile(ead.Ceadheader)
		desc.Summary.File = newFile(ead.Ceadheader)
		desc.Summary.FindingAid = newFindingAid(ead.Ceadheader)
		err := desc.Summary.FindingAid.AddUnit(ead.Carchdesc)
		if err != nil {
			return nil, err
		}

	}
	return desc, nil
}

// newProfile creates a new *Profile from the eadheader profilestmt.
func newProfile(header *Ceadheader) *Profile {
	if header.Cprofiledesc != nil {
		profile := new(Profile)
		if header.Cprofiledesc.Clangusage != nil {
			profile.Language = sanitizeXMLAsString(header.Cprofiledesc.Clangusage.LangUsage)
		}
		if header.Cprofiledesc.Ccreation != nil {
			profile.Creation = sanitizeXMLAsString(header.Cprofiledesc.Ccreation.Creation)
		}
		return profile
	}
	return nil
}

// newFile creates a new *File from the eadheader filestmt.
func newFile(header *Ceadheader) *File {
	fileDesc := header.Cfiledesc
	if fileDesc != nil {
		file := new(File)
		if fileDesc.Ctitlestmt != nil {
			if fileDesc.Ctitlestmt.Ctitleproper != nil {
				file.Title = sanitizeXMLAsString(
					fileDesc.Ctitlestmt.Ctitleproper.TitleProper,
				)
			}
			if fileDesc.Ctitlestmt.Cauthor != nil {
				file.Author = sanitizeXMLAsString(
					fileDesc.Ctitlestmt.Cauthor.Author,
				)
			}
		}
		if fileDesc.Ceditionstmt != nil {
			for _, edition := range fileDesc.Ceditionstmt.Cedition {
				file.Edition = append(
					file.Edition,
					sanitizeXMLAsString(edition.Edition),
				)
			}
		}
		if fileDesc.Cpublicationstmt != nil {
			if fileDesc.Cpublicationstmt.Cpublisher != nil {
				file.Publisher = fileDesc.Cpublicationstmt.Cpublisher.Publisher
			}
			if fileDesc.Cpublicationstmt.Cdate != nil {
				file.PublicationDate = fileDesc.Cpublicationstmt.Cdate.Date
			}
			if len(fileDesc.Cpublicationstmt.Cp) > 0 {
				for _, p := range fileDesc.Cpublicationstmt.Cp {
					if p.Attrid == "copyright" && p.Cextref != nil {
						file.Copyright = p.Cextref.ExtRef
						file.CopyrightURI = p.Cextref.Attrhref
					}
				}
			}
		}
		return file
	}
	return nil
}

// newFindingAid creates a new FindingAid with information from the EadHeader.
// You must call AddUnit to populate the *Unit.
func newFindingAid(header *Ceadheader) *FindingAid {
	if header.Ceadid != nil {
		aid := new(FindingAid)
		aid.ID = header.Ceadid.EadID
		aid.Country = header.Ceadid.Attrcountrycode
		aid.AgencyCode = header.Ceadid.Attrmainagencycode
		return aid
	}
	return nil
}

// AddUnit adds the DID information from the ArchDesc to the FindingAid.
func (fa *FindingAid) AddUnit(archdesc *Carchdesc) error {
	if archdesc.Cdid != nil && fa != nil {
		did := archdesc.Cdid
		for _, title := range did.Cunittitle {
			if title.Attrtype == "short" {
				fa.ShortTitle = sanitizeXMLAsString(title.RawTitle)
				continue
			}
			fa.Title = append(
				fa.Title,
				sanitizeXMLAsString(title.RawTitle),
			)
		}

		unit := new(Unit)

		// only write one ID, only clevel unitids have more than one
		for _, unitid := range did.Cunitid {
			unit.ID = unitid.ID
		}

		for _, date := range did.Cunitdate {
			if date != nil {
				switch date.Attrtype {
				case "bulk":
					unit.DateBulk = date.Date
				default:
					unit.Date = append(unit.Date, date.Date)
				}
			}

		}

		if did.Cphysdesc != nil {
			for _, extent := range did.Cphysdesc.Cextent {
				switch extent.Attrunit {
				case "files":
					unit.Files = extent.Extent
				case "meter", "metre", "metres":
					unit.Length = extent.Extent
				}
			}
			unit.Physical = sanitizeXMLAsString(did.Cphysdesc.Raw)
		}

		if did.Clangmaterial != nil {
			unit.Language = sanitizeXMLAsString(did.Clangmaterial.Raw)
		}

		if did.Cmaterialspec != nil {
			unit.Material = sanitizeXMLAsString(did.Cmaterialspec.Raw)
		}

		if did.Crepository != nil {
			unit.Repository = sanitizeXMLAsString(did.Crepository.Raw)
		}

		if did.Cphysloc != nil {
			unit.PhysicalLocation = sanitizeXMLAsString(did.Cphysloc.Raw)
		}

		if did.Corigination != nil {
			unit.Origin = sanitizeXMLAsString(did.Corigination.Raw)
		}

		if did.Cabstract != nil {
			unit.Abstract = did.Cabstract.Abstract()
		}

		fa.Unit = unit
	}

	return nil
}
