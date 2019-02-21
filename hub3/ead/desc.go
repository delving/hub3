package ead

// Description holds the information from the EAD ArchDesc in a structured form
// that can be easily digested by webservices. It aims to hide the structural
// diversity at the lowest level in the ArchDesc.
type Description struct {
	Changes []struct {
		Author       string `json:"author"`
		ChangeNumber string `json:"changeNumber"`
		Date         string `json:"date"`
		Description  string `json:"description"`
	} `json:"changes"`
	DescriptionGroups []struct {
		Abstract struct {
			Label string `json:"label"`
			Text  string `json:"text"`
		} `json:"abstract"`
		ArchiveTitle string `json:"archiveTitle"`
		Author       string `json:"author"`
		DateBulk     struct {
			Normal string `json:"normal"`
			Text   string `json:"text"`
		} `json:"dateBulk"`
		DateInclusive struct {
			Normal string `json:"normal"`
			Text   string `json:"text"`
		} `json:"dateInclusive"`
		EadID        string `json:"eadId"`
		Head         string `json:"head"`
		LangMaterial struct {
			Label      string `json:"label"`
			LangCode   string `json:"langCode"`
			ScriptCode string `json:"scriptCode"`
			Text       string `json:"text"`
		} `json:"langMaterial"`
		Language string `json:"language"`
		Licence  struct {
			Link string `json:"link"`
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"licence"`
		MaterialSpec struct {
			Label string `json:"label"`
			Text  string `json:"text"`
		} `json:"materialSpec"`
		Origination struct {
			Corpname string `json:"corpname"`
			Label    string `json:"label"`
		} `json:"origination"`
		Periods  []string `json:"periods"`
		PhysDesc struct {
			Extent []struct {
				Number string `json:"number"`
				Text   string `json:"text"`
				Units  string `json:"units"`
			} `json:"extent"`
			Label string `json:"label"`
		} `json:"physDesc"`
		Publisher  string `json:"publisher"`
		Repository struct {
			Label string `json:"label"`
			Text  string `json:"text"`
		} `json:"repository"`
		Type   string `json:"type"`
		UnitID struct {
			CountryCode    string `json:"countryCode"`
			Label          string `json:"label"`
			RepositoryCode string `json:"repositoryCode"`
			Text           string `json:"text"`
		} `json:"unitId"`
		UnitShortTitle string `json:"unitShortTitle"`
		UnitTitle      string `json:"unitTitle"`
		Version        string `json:"version"`
	} `json:"descriptionGroups"`
}
