package alto

import (
	"encoding/xml"
	"fmt"
	"io"
)

type Description struct {
	XMLName                xml.Name                `xml:"Description,omitempty" json:"Description,omitempty"`
	MeasurementUnit        *MeasurementUnit        `xml:"MeasurementUnit,omitempty" json:"MeasurementUnit,omitempty"`
	OCRProcessing          *OCRProcessing          `xml:"OCRProcessing,omitempty" json:"OCRProcessing,omitempty"`
	SourceImageInformation *SourceImageInformation `xml:"sourceImageInformation,omitempty" json:"sourceImageInformation,omitempty"`
}

type Layout struct {
	XMLName xml.Name `xml:"Layout,omitempty" json:"Layout,omitempty"`
	Page    *Page    `xml:"Page,omitempty" json:"Page,omitempty"`
}

type MeasurementUnit struct {
	XMLName xml.Name `xml:"MeasurementUnit,omitempty" json:"MeasurementUnit,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type OCRProcessing struct {
	XMLName           xml.Name           `xml:"OCRProcessing,omitempty" json:"OCRProcessing,omitempty"`
	OcrProcessingStep *OcrProcessingStep `xml:"ocrProcessingStep,omitempty" json:"ocrProcessingStep,omitempty"`
}

type Page struct {
	XMLName    xml.Name    `xml:"Page,omitempty" json:"Page,omitempty"`
	AttrHEIGHT string      `xml:"HEIGHT,attr"  json:",omitempty"`
	AttrWIDTH  string      `xml:"WIDTH,attr"  json:",omitempty"`
	PrintSpace *PrintSpace `xml:"PrintSpace,omitempty" json:"PrintSpace,omitempty"`
}

type PrintSpace struct {
	XMLName    xml.Name     `xml:"PrintSpace,omitempty" json:"PrintSpace,omitempty"`
	AttrHEIGHT string       `xml:"HEIGHT,attr"  json:",omitempty"`
	AttrHPOS   string       `xml:"HPOS,attr"  json:",omitempty"`
	AttrVPOS   string       `xml:"VPOS,attr"  json:",omitempty"`
	AttrWIDTH  string       `xml:"WIDTH,attr"  json:",omitempty"`
	TextBlock  []*TextBlock `xml:"TextBlock,omitempty" json:"TextBlock,omitempty"`
}

type String struct {
	XMLName     xml.Name `xml:"String,omitempty" json:"String,omitempty"`
	AttrCONTENT string   `xml:"CONTENT,attr"  json:",omitempty"`
	AttrHEIGHT  string   `xml:"HEIGHT,attr"  json:",omitempty"`
	AttrHPOS    string   `xml:"HPOS,attr"  json:",omitempty"`
	AttrVPOS    string   `xml:"VPOS,attr"  json:",omitempty"`
	AttrWIDTH   string   `xml:"WIDTH,attr"  json:",omitempty"`
}

type TextBlock struct {
	XMLName    xml.Name  `xml:"TextBlock,omitempty" json:"TextBlock,omitempty"`
	AttrHEIGHT string    `xml:"HEIGHT,attr"  json:",omitempty"`
	AttrHPOS   string    `xml:"HPOS,attr"  json:",omitempty"`
	AttrID     string    `xml:"ID,attr"  json:",omitempty"`
	AttrVPOS   string    `xml:"VPOS,attr"  json:",omitempty"`
	AttrWIDTH  string    `xml:"WIDTH,attr"  json:",omitempty"`
	TextLine   *TextLine `xml:"TextLine,omitempty" json:"TextLine,omitempty"`
}

type TextLine struct {
	XMLName    xml.Name `xml:"TextLine,omitempty" json:"TextLine,omitempty"`
	AttrHEIGHT string   `xml:"HEIGHT,attr"  json:",omitempty"`
	AttrHPOS   string   `xml:"HPOS,attr"  json:",omitempty"`
	AttrVPOS   string   `xml:"VPOS,attr"  json:",omitempty"`
	AttrWIDTH  string   `xml:"WIDTH,attr"  json:",omitempty"`
	String     *String  `xml:"String,omitempty" json:"String,omitempty"`
}

type Alto struct {
	XMLName     xml.Name     `xml:"alto,omitempty" json:"alto,omitempty"`
	Description *Description `xml:"Description,omitempty" json:"Description,omitempty"`
	Layout      *Layout      `xml:"Layout,omitempty" json:"Layout,omitempty"`
}

type FileName struct {
	XMLName xml.Name `xml:"fileName,omitempty" json:"fileName,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type OcrProcessingStep struct {
	XMLName                xml.Name                `xml:"ocrProcessingStep,omitempty" json:"ocrProcessingStep,omitempty"`
	ProcessingDateTime     *ProcessingDateTime     `xml:"processingDateTime,omitempty" json:"processingDateTime,omitempty"`
	ProcessingSoftware     *ProcessingSoftware     `xml:"processingSoftware,omitempty" json:"processingSoftware,omitempty"`
	ProcessingStepSettings *ProcessingStepSettings `xml:"processingStepSettings,omitempty" json:"processingStepSettings,omitempty"`
}

type ProcessingDateTime struct {
	XMLName xml.Name `xml:"processingDateTime,omitempty" json:"processingDateTime,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type ProcessingSoftware struct {
	XMLName xml.Name `xml:"processingSoftware,omitempty" json:"processingSoftware,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type ProcessingStepSettings struct {
	XMLName xml.Name `xml:"processingStepSettings,omitempty" json:"processingStepSettings,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type SourceImageInformation struct {
	XMLName  xml.Name  `xml:"sourceImageInformation,omitempty" json:"sourceImageInformation,omitempty"`
	FileName *FileName `xml:"fileName,omitempty" json:"fileName,omitempty"`
}

func (a *Alto) extractStrings() ([]string, error) {
	var content []string
	space := a.Layout.Page.PrintSpace
	if space == nil {
		return content, fmt.Errorf("empty page")
	}

	for _, block := range space.TextBlock {
		text := block.TextLine.String
		if text != nil && text.AttrCONTENT != "" {
			content = append(content, text.AttrCONTENT)
		}
	}

	return content, nil
}

func (a *Alto) WriteTo(w io.Writer) (int64, error) {
	content, err := a.extractStrings()
	if err != nil {
		return 0, err
	}

	var written int64
	for _, text := range content {
		n, err := fmt.Fprintln(w, text)
		if err != nil {
			return 0, err
		}

		written += int64(n)
	}

	return written, nil
}
