// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

var xsdLabel2ObjectXSDType = make(map[string]int32)

var int2ObjectXSDType = map[int32]ObjectXSDType{
	0:  ObjectXSDType_STRING,
	1:  ObjectXSDType_BOOLEAN,
	2:  ObjectXSDType_DECIMAL,
	3:  ObjectXSDType_FLOAT,
	4:  ObjectXSDType_DOUBLE,
	5:  ObjectXSDType_DATETIME,
	6:  ObjectXSDType_TIME,
	7:  ObjectXSDType_DATE,
	8:  ObjectXSDType_GYEARMONTH,
	9:  ObjectXSDType_GYEAR,
	10: ObjectXSDType_GMONTHDAY,
	11: ObjectXSDType_GDAY,
	12: ObjectXSDType_GMONTH,
	13: ObjectXSDType_HEXBINARY,
	14: ObjectXSDType_BASE64BINARY,
	15: ObjectXSDType_ANYURI,
	16: ObjectXSDType_NORMALIZEDSTRING,
	17: ObjectXSDType_TOKEN,
	18: ObjectXSDType_LANGUAGE,
	19: ObjectXSDType_NMTOKEN,
	20: ObjectXSDType_NAME,
	21: ObjectXSDType_NCNAME,
	22: ObjectXSDType_INTEGER,
	23: ObjectXSDType_NONPOSITIVEINTEGER,
	24: ObjectXSDType_NEGATIVEINTEGER,
	25: ObjectXSDType_LONG,
	26: ObjectXSDType_INT,
	27: ObjectXSDType_SHORT,
	28: ObjectXSDType_BYTE,
	29: ObjectXSDType_NONNEGATIVEINTEGER,
	30: ObjectXSDType_UNSIGNEDLONG,
	31: ObjectXSDType_UNSIGNEDINT,
	32: ObjectXSDType_UNSIGNEDSHORT,
	33: ObjectXSDType_UNSIGNEDBYTE,
	34: ObjectXSDType_POSITIVEINTEGER,
}

var objectXSDType2XSDLabel = map[int32]string{
	0:  "xsd:string",
	1:  "xsd:boolean",
	2:  "xsd:decimal",
	3:  "xsd:float",
	4:  "xsd:double",
	5:  "xsd:dateTime",
	6:  "xsd:time",
	7:  "xsd:date",
	8:  "xsd:gYearMonth",
	9:  "xsd:gYear",
	10: "xsd:gMonthDay",
	11: "xsd:gDay",
	12: "xsd:gMonth",
	13: "xsd:hexBinary",
	14: "xsd:base64Binary",
	15: "xsd:anyURI",
	16: "xsd:normalizedString",
	17: "xsd:token",
	18: "xsd:language",
	19: "xsd:NMTOKEN",
	20: "xsd:Name",
	21: "xsd:NCName",
	22: "xsd:integer",
	23: "xsd:nonPositiveInteger",
	24: "xsd:negativeInteger",
	25: "xsd:long",
	26: "xsd:int",
	27: "xsd:short",
	28: "xsd:byte",
	29: "xsd:nonNegativeInteger",
	30: "xsd:unsignedLong",
	31: "xsd:unsignedInt",
	32: "xsd:unsignedShort",
	33: "xsd:unsignedByte",
	34: "xsd:positiveInteger",
}
