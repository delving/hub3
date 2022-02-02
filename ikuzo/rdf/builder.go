package rdf

type Builder struct{}

func (b *Builder) IRI(str string) *IRI {
	return &IRI{str: str}
}

func (b *Builder) Literal(str string) *Literal {
	return &Literal{str: str, DataType: rdfLangString}
}

func (b *Builder) LiteralWithLang(str, lang string) *Literal {
	return &Literal{str: str, lang: lang, DataType: rdfLangString}
}

func (b *Builder) LiteralWithDataType(str string, dataType *IRI) *Literal {
	return &Literal{str: str, DataType: dataType}
}
