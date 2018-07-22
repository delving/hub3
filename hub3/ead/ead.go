package ead

// NewNodeList converts the Archival Description Level to a Nodelist
// Nodelist is an optimized lossless Protocol Buffer container.
func (dsc *Cdsc) NewNodeList() (*NodeList, error) {
	nl := &NodeList{}
	nl.Type = dsc.Attrtype
	for _, label := range dsc.Chead {
		nl.Label = append(nl.Label, label.Head)
	}
	for _, c1 := range dsc.Cc01 {
		node, err := c1.NewNode()
		if err != nil {
			return nil, err
		}
		nl.Nodes = append(nl.Nodes, node)
	}
	return nl, nil
}

// NewNodeID converts a unitid field from the EAD did to a NodeID
func (ui *Cunitid) NewNodeID() (*NodeID, error) {
	id := &NodeID{
		ID:       ui.ID,
		TypeID:   ui.Attridentifier,
		Type:     ui.Attrtype,
		Audience: ui.Attraudience,
	}
	return id, nil
}

// NewNodeIDs extract Unit Identifiers from the EAD did
func (cdid *Cdid) NewNodeIDs() ([]*NodeID, string, error) {
	ids := make([]*NodeID, len(cdid.Cunitid))
	var invertoryNumber string
	for _, unitid := range cdid.Cunitid {
		id, err := unitid.NewNodeID()
		if err != nil {
			return nil, "", err
		}
		if id.GetType() == "ABS" {
			invertoryNumber = id.GetID()
		}
		ids = append(ids, id)
	}
	return ids, invertoryNumber, nil
}

// NewNodeDate extract date infomation frme the EAD unitdate
func (date *Cunitdate) NewNodeDate() (*NodeDate, error) {
	nDate := &NodeDate{
		Calendar: date.Attrcalendar,
		Era:      date.Attrera,
		Normal:   date.Attrnormal,
		Label:    date.Date,
	}
	return nDate, nil
}

// NewHeader creates an Archival Header
func (cdid *Cdid) NewHeader() (*Header, error) {
	header := &Header{}
	if cdid.Cphysdesc != nil {
		header.Physdesc = cdid.Cphysdesc.PhyscDesc
	}

	for _, label := range cdid.Cunittitle {
		// TODO add check for embedded date field
		header.Label = append(header.Label, label.Title)
	}

	for _, date := range cdid.Cunitdate {
		nodeDate, err := date.NewNodeDate()
		if err != nil {
			return nil, err
		}
		header.Date = append(header.Date, nodeDate)
	}

	nodeIDs, inventoryID, err := cdid.NewNodeIDs()
	if err != nil {
		return nil, err
	}
	if inventoryID != "" {
		header.InventoryNumber = inventoryID
	}
	header.ID = append(header.ID, nodeIDs...)

	return header, nil
}

// NewNode converts EAD c01 to a Archival Node
func (c *Cc01) NewNode() (*Node, error) {
	node := &Node{
		CTag:  c.XMLName.Local,
		Depth: int32(1),
		Type:  c.Attrlevel,
	}
	header, err := c.Cdid.NewHeader()
	if err != nil {
		return nil, err
	}
	node.Header = header

	parentIDS := []string{header.GetInventoryNumber()}

	if len(c.Cc02) != 0 {
		for _, c02 := range c.Cc02 {
			n, err := c02.NewNode(parentIDS)
			if err != nil {
				return nil, err
			}
			node.Nodes = append(node.Nodes, n)
		}
	}
	return node, nil
}

func (c *Cc02) NewNode(parentIDS []string) (*Node, error) {
	node := &Node{
		CTag:      c.XMLName.Local,
		Depth:     int32(2),
		Type:      c.Attrlevel,
		ParentIDS: parentIDS,
	}
	header, err := c.Cdid.NewHeader()
	if err != nil {
		return nil, err
	}
	node.Header = header

	parentIDS = append(parentIDS, header.GetInventoryNumber())

	if len(c.Cc03) != 0 {
		for _, nested := range c.Cc03 {
			n, err := nested.NewNode(parentIDS)
			if err != nil {
				return nil, err
			}
			node.Nodes = append(node.Nodes, n)
		}
	}
	return node, nil
}

func (c *Cc03) NewNode(parentIDS []string) (*Node, error) {
	node := &Node{
		CTag:      c.XMLName.Local,
		Depth:     int32(3),
		Type:      c.Attrlevel,
		ParentIDS: parentIDS,
	}
	header, err := c.Cdid.NewHeader()
	if err != nil {
		return nil, err
	}
	node.Header = header

	parentIDS = append(parentIDS, header.GetInventoryNumber())

	for _, nested := range c.Cc04 {
		n, err := nested.NewNode(parentIDS)
		if err != nil {
			return nil, err
		}
		node.Nodes = append(node.Nodes, n)
	}

	//for _, content := range c.Cscopecontent {

	//}

	return node, nil
}

func (c *Cc04) NewNode(parentIDS []string) (*Node, error) {
	node := &Node{
		CTag:      c.XMLName.Local,
		Depth:     int32(4),
		Type:      c.Attrlevel,
		ParentIDS: parentIDS,
	}
	header, err := c.Cdid.NewHeader()
	if err != nil {
		return nil, err
	}
	node.Header = header

	parentIDS = append(parentIDS, header.GetInventoryNumber())

	//if len(c.Cc05) != 0 {
	//for _, nested := range c.Cc05 {
	//n, err := nested.NewNode(parentIDS)
	//if err != nil {
	//return nil, err
	//}
	//node.Nodes = append(node.Nodes, n)
	//}
	//}
	return node, nil
}
