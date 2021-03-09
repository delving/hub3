// Code generated by oto; DO NOT EDIT.

'use strict';

const jsonDef = `{"packageName":"generated","services":[{"name":"NamespaceService","methods":[{"name":"DeleteNamespace","nameLowerCamel":"deleteNamespace","inputObject":{"typeID":"command-line-arguments.DeleteNamespaceRequest","typeName":"DeleteNamespaceRequest","objectName":"DeleteNamespaceRequest","objectNameLowerCamel":"deleteNamespaceRequest","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"outputObject":{"typeID":"command-line-arguments.DeleteNamespaceResponse","typeName":"DeleteNamespaceResponse","objectName":"DeleteNamespaceResponse","objectNameLowerCamel":"deleteNamespaceResponse","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"comment":"DeletetNamespace deletes a Namespace","metadata":{"CAUTION":"You may lose data"}},{"name":"GetNamespace","nameLowerCamel":"getNamespace","inputObject":{"typeID":"command-line-arguments.GetNamespaceRequest","typeName":"GetNamespaceRequest","objectName":"GetNamespaceRequest","objectNameLowerCamel":"getNamespaceRequest","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"outputObject":{"typeID":"command-line-arguments.GetNamespaceResponse","typeName":"GetNamespaceResponse","objectName":"GetNamespaceResponse","objectNameLowerCamel":"getNamespaceResponse","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"comment":"GetNamespace gets a Namespace","metadata":{}},{"name":"PutNamespace","nameLowerCamel":"putNamespace","inputObject":{"typeID":"command-line-arguments.PutNamespaceRequest","typeName":"PutNamespaceRequest","objectName":"PutNamespaceRequest","objectNameLowerCamel":"putNamespaceRequest","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"outputObject":{"typeID":"command-line-arguments.PutNamespaceResponse","typeName":"PutNamespaceResponse","objectName":"PutNamespaceResponse","objectNameLowerCamel":"putNamespaceResponse","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"comment":"PutNamespace stores a Namespace","metadata":{}},{"name":"Search","nameLowerCamel":"search","inputObject":{"typeID":"command-line-arguments.SearchNamespaceRequest","typeName":"SearchNamespaceRequest","objectName":"SearchNamespaceRequest","objectNameLowerCamel":"searchNamespaceRequest","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"outputObject":{"typeID":"command-line-arguments.SearchNamespaceResponse","typeName":"SearchNamespaceResponse","objectName":"SearchNamespaceResponse","objectNameLowerCamel":"searchNamespaceResponse","multiple":false,"package":"","isObject":true,"jsType":"object","swiftType":"Any"},"comment":"Search returns a filtered list of Namespaces","metadata":{}}],"comment":"NamespaceService allows you to programmatically manage namespaces","metadata":{}}],"objects":[{"typeID":"command-line-arguments.DeleteNamespaceRequest","name":"DeleteNamespaceRequest","imported":false,"fields":[{"name":"ID","nameLowerCamel":"id","type":{"typeID":"command-line-arguments.string","typeName":"string","objectName":"string","objectNameLowerCamel":"string","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":false,"comment":"ID is the unique identifier of a Namespace","tag":"","parsedTags":{},"example":null,"metadata":{}}],"comment":"DeleteNamespaceRequest is the input object for NamespaceService.DeleteNamespace","metadata":{}},{"typeID":"command-line-arguments.DeleteNamespaceResponse","name":"DeleteNamespaceResponse","imported":false,"fields":[{"name":"Error","nameLowerCamel":"error","type":{"typeID":"","typeName":"string","objectName":"","objectNameLowerCamel":"","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":true,"comment":"Error is string explaining what went wrong. Empty if everything was fine.","tag":"","parsedTags":null,"example":"something went wrong","metadata":{}}],"comment":"DeleteNamespaceRequest is the output object for NamespaceService.DeleteNamespace","metadata":{}},{"typeID":"command-line-arguments.GetNamespaceRequest","name":"GetNamespaceRequest","imported":false,"fields":[{"name":"ID","nameLowerCamel":"id","type":{"typeID":"command-line-arguments.string","typeName":"string","objectName":"string","objectNameLowerCamel":"string","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":false,"comment":"ID is the unique identifier of the namespace","tag":"","parsedTags":{},"example":"123","metadata":{"example":"123"}}],"comment":"GetNamespaceRequest is the input object for GetNamespaceService.GetNamespace","metadata":{}},{"typeID":"command-line-arguments.GetNamespaceResponse","name":"GetNamespaceResponse","imported":false,"fields":[{"name":"Namespace","nameLowerCamel":"namespace","type":{"typeID":"github.com/delving/hub3/ikuzo/domain.*Namespace","typeName":"*domain.Namespace","objectName":"*Namespace","objectNameLowerCamel":"*Namespace","multiple":false,"package":"github.com/delving/hub3/ikuzo/domain","isObject":false,"jsType":"","swiftType":""},"omitEmpty":false,"comment":"Namespace is the Namespace","tag":"","parsedTags":{},"example":null,"metadata":{}},{"name":"Error","nameLowerCamel":"error","type":{"typeID":"","typeName":"string","objectName":"","objectNameLowerCamel":"","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":true,"comment":"Error is string explaining what went wrong. Empty if everything was fine.","tag":"","parsedTags":null,"example":"something went wrong","metadata":{}}],"comment":"GetNamespaceResponse is the output object for GetNamespaceService.GetNamespace","metadata":{}},{"typeID":"command-line-arguments.PutNamespaceRequest","name":"PutNamespaceRequest","imported":false,"fields":[{"name":"Namespace","nameLowerCamel":"namespace","type":{"typeID":"github.com/delving/hub3/ikuzo/domain.*Namespace","typeName":"*domain.Namespace","objectName":"*Namespace","objectNameLowerCamel":"*Namespace","multiple":false,"package":"github.com/delving/hub3/ikuzo/domain","isObject":false,"jsType":"","swiftType":""},"omitEmpty":false,"comment":"","tag":"","parsedTags":{},"example":null,"metadata":{}}],"comment":"PutNamespaceRequest is the input object for NamespaceService.PutNamespace","metadata":{}},{"typeID":"command-line-arguments.PutNamespaceResponse","name":"PutNamespaceResponse","imported":false,"fields":[{"name":"Error","nameLowerCamel":"error","type":{"typeID":"","typeName":"string","objectName":"","objectNameLowerCamel":"","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":true,"comment":"Error is string explaining what went wrong. Empty if everything was fine.","tag":"","parsedTags":null,"example":"something went wrong","metadata":{}}],"comment":"PutNamespaceResponse is the output object for NamespaceService.PutNamespace","metadata":{}},{"typeID":"command-line-arguments.SearchNamespaceRequest","name":"SearchNamespaceRequest","imported":false,"fields":[{"name":"Prefix","nameLowerCamel":"prefix","type":{"typeID":"command-line-arguments.string","typeName":"string","objectName":"string","objectNameLowerCamel":"string","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":false,"comment":"Prefix for a Namespace","tag":"","parsedTags":{},"example":null,"metadata":{}},{"name":"BaseURI","nameLowerCamel":"baseURI","type":{"typeID":"command-line-arguments.string","typeName":"string","objectName":"string","objectNameLowerCamel":"string","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":false,"comment":"BaseURI for a Namespace","tag":"","parsedTags":{},"example":null,"metadata":{}}],"comment":"SearchNamespaceRequest is the input object for NamespaceService.Search","metadata":{}},{"typeID":"command-line-arguments.SearchNamespaceResponse","name":"SearchNamespaceResponse","imported":false,"fields":[{"name":"Hits","nameLowerCamel":"hits","type":{"typeID":"github.com/delving/hub3/ikuzo/domain.*Namespace","typeName":"*domain.Namespace","objectName":"*Namespace","objectNameLowerCamel":"*Namespace","multiple":true,"package":"github.com/delving/hub3/ikuzo/domain","isObject":false,"jsType":"","swiftType":""},"omitEmpty":false,"comment":"Hits returns the list of matching Namespaces","tag":"","parsedTags":{},"example":null,"metadata":{}},{"name":"More","nameLowerCamel":"more","type":{"typeID":"command-line-arguments.bool","typeName":"bool","objectName":"bool","objectNameLowerCamel":"bool","multiple":false,"package":"","isObject":false,"jsType":"boolean","swiftType":"Bool"},"omitEmpty":false,"comment":"More indicates that there may be more search results. If true, make the same Search request passing this Cursor.","tag":"","parsedTags":{},"example":null,"metadata":{}},{"name":"Error","nameLowerCamel":"error","type":{"typeID":"","typeName":"string","objectName":"","objectNameLowerCamel":"","multiple":false,"package":"","isObject":false,"jsType":"string","swiftType":"String"},"omitEmpty":true,"comment":"Error is string explaining what went wrong. Empty if everything was fine.","tag":"","parsedTags":null,"example":"something went wrong","metadata":{}}],"comment":"SearchNamespaceResponse is the output object for NamespaceService.Search","metadata":{}}],"imports":{"github.com/delving/hub3/ikuzo/domain":"domain"}}`
export const def = JSON.parse(jsonDef.replace(/\n/g, '\\n'));
