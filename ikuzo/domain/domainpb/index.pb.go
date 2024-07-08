// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v5.27.2
// source: ikuzo/domain/domainpb/index.proto

package domainpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Action describes the actions
type ActionType int32

const (
	ActionType_MODIFY_INDEX  ActionType = 0
	ActionType_DROP_ORPHANS  ActionType = 1
	ActionType_DELETE_RECORD ActionType = 2
)

// Enum value maps for ActionType.
var (
	ActionType_name = map[int32]string{
		0: "MODIFY_INDEX",
		1: "DROP_ORPHANS",
		2: "DELETE_RECORD",
	}
	ActionType_value = map[string]int32{
		"MODIFY_INDEX":  0,
		"DROP_ORPHANS":  1,
		"DELETE_RECORD": 2,
	}
)

func (x ActionType) Enum() *ActionType {
	p := new(ActionType)
	*p = x
	return p
}

func (x ActionType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ActionType) Descriptor() protoreflect.EnumDescriptor {
	return file_ikuzo_domain_domainpb_index_proto_enumTypes[0].Descriptor()
}

func (ActionType) Type() protoreflect.EnumType {
	return &file_ikuzo_domain_domainpb_index_proto_enumTypes[0]
}

func (x ActionType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ActionType.Descriptor instead.
func (ActionType) EnumDescriptor() ([]byte, []int) {
	return file_ikuzo_domain_domainpb_index_proto_rawDescGZIP(), []int{0}
}

// IndexType describes the supported index mapping types
type IndexType int32

const (
	IndexType_V2              IndexType = 0
	IndexType_V1              IndexType = 1
	IndexType_FRAGMENTS       IndexType = 2
	IndexType_DIGITAL_OBJECTS IndexType = 3
	IndexType_SUGGEST         IndexType = 4
)

// Enum value maps for IndexType.
var (
	IndexType_name = map[int32]string{
		0: "V2",
		1: "V1",
		2: "FRAGMENTS",
		3: "DIGITAL_OBJECTS",
		4: "SUGGEST",
	}
	IndexType_value = map[string]int32{
		"V2":              0,
		"V1":              1,
		"FRAGMENTS":       2,
		"DIGITAL_OBJECTS": 3,
		"SUGGEST":         4,
	}
)

func (x IndexType) Enum() *IndexType {
	p := new(IndexType)
	*p = x
	return p
}

func (x IndexType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (IndexType) Descriptor() protoreflect.EnumDescriptor {
	return file_ikuzo_domain_domainpb_index_proto_enumTypes[1].Descriptor()
}

func (IndexType) Type() protoreflect.EnumType {
	return &file_ikuzo_domain_domainpb_index_proto_enumTypes[1]
}

func (x IndexType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use IndexType.Descriptor instead.
func (IndexType) EnumDescriptor() ([]byte, []int) {
	return file_ikuzo_domain_domainpb_index_proto_rawDescGZIP(), []int{1}
}

// IndexMessage is used to queue messages for indexing by ElasticSearch.
type IndexMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OrganisationID string     `protobuf:"bytes,1,opt,name=OrganisationID,proto3" json:"OrganisationID,omitempty"`
	DatasetID      string     `protobuf:"bytes,2,opt,name=DatasetID,proto3" json:"DatasetID,omitempty"`
	RecordID       string     `protobuf:"bytes,3,opt,name=RecordID,proto3" json:"RecordID,omitempty"`
	IndexName      string     `protobuf:"bytes,4,opt,name=IndexName,proto3" json:"IndexName,omitempty"`
	Deleted        bool       `protobuf:"varint,5,opt,name=Deleted,proto3" json:"Deleted,omitempty"`
	Revision       *Revision  `protobuf:"bytes,6,opt,name=Revision,proto3" json:"Revision,omitempty"`
	Source         []byte     `protobuf:"bytes,7,opt,name=Source,proto3" json:"Source,omitempty"`
	ActionType     ActionType `protobuf:"varint,8,opt,name=ActionType,proto3,enum=domainpb.ActionType" json:"ActionType,omitempty"`
	IndexType      IndexType  `protobuf:"varint,9,opt,name=IndexType,proto3,enum=domainpb.IndexType" json:"IndexType,omitempty"`
}

func (x *IndexMessage) Reset() {
	*x = IndexMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ikuzo_domain_domainpb_index_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IndexMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IndexMessage) ProtoMessage() {}

func (x *IndexMessage) ProtoReflect() protoreflect.Message {
	mi := &file_ikuzo_domain_domainpb_index_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IndexMessage.ProtoReflect.Descriptor instead.
func (*IndexMessage) Descriptor() ([]byte, []int) {
	return file_ikuzo_domain_domainpb_index_proto_rawDescGZIP(), []int{0}
}

func (x *IndexMessage) GetOrganisationID() string {
	if x != nil {
		return x.OrganisationID
	}
	return ""
}

func (x *IndexMessage) GetDatasetID() string {
	if x != nil {
		return x.DatasetID
	}
	return ""
}

func (x *IndexMessage) GetRecordID() string {
	if x != nil {
		return x.RecordID
	}
	return ""
}

func (x *IndexMessage) GetIndexName() string {
	if x != nil {
		return x.IndexName
	}
	return ""
}

func (x *IndexMessage) GetDeleted() bool {
	if x != nil {
		return x.Deleted
	}
	return false
}

func (x *IndexMessage) GetRevision() *Revision {
	if x != nil {
		return x.Revision
	}
	return nil
}

func (x *IndexMessage) GetSource() []byte {
	if x != nil {
		return x.Source
	}
	return nil
}

func (x *IndexMessage) GetActionType() ActionType {
	if x != nil {
		return x.ActionType
	}
	return ActionType_MODIFY_INDEX
}

func (x *IndexMessage) GetIndexType() IndexType {
	if x != nil {
		return x.IndexType
	}
	return IndexType_V2
}

// Version of the record in the time-revision-store.
type Revision struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SHA  string `protobuf:"bytes,1,opt,name=SHA,proto3" json:"SHA,omitempty"`
	Path string `protobuf:"bytes,2,opt,name=Path,proto3" json:"Path,omitempty"`
	// for legacy use only
	Number int32 `protobuf:"varint,3,opt,name=Number,proto3" json:"Number,omitempty"`
	// group for orphan control
	GroupID string `protobuf:"bytes,4,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
}

func (x *Revision) Reset() {
	*x = Revision{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ikuzo_domain_domainpb_index_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Revision) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Revision) ProtoMessage() {}

func (x *Revision) ProtoReflect() protoreflect.Message {
	mi := &file_ikuzo_domain_domainpb_index_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Revision.ProtoReflect.Descriptor instead.
func (*Revision) Descriptor() ([]byte, []int) {
	return file_ikuzo_domain_domainpb_index_proto_rawDescGZIP(), []int{1}
}

func (x *Revision) GetSHA() string {
	if x != nil {
		return x.SHA
	}
	return ""
}

func (x *Revision) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Revision) GetNumber() int32 {
	if x != nil {
		return x.Number
	}
	return 0
}

func (x *Revision) GetGroupID() string {
	if x != nil {
		return x.GroupID
	}
	return ""
}

var File_ikuzo_domain_domainpb_index_proto protoreflect.FileDescriptor

var file_ikuzo_domain_domainpb_index_proto_rawDesc = []byte{
	0x0a, 0x21, 0x69, 0x6b, 0x75, 0x7a, 0x6f, 0x2f, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x2f, 0x64,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x08, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x22, 0xd9, 0x02,
	0x0a, 0x0c, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x26,
	0x0a, 0x0e, 0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x73, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x73, 0x65,
	0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x44, 0x61, 0x74, 0x61, 0x73,
	0x65, 0x74, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x49, 0x44,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x49, 0x44,
	0x12, 0x1c, 0x0a, 0x09, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x07, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x12, 0x2e, 0x0a, 0x08, 0x52, 0x65, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x64, 0x6f, 0x6d,
	0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x08,
	0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x53, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x12, 0x34, 0x0a, 0x0a, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e,
	0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x41, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x31, 0x0a, 0x09, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x54,
	0x79, 0x70, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x64, 0x6f, 0x6d, 0x61,
	0x69, 0x6e, 0x70, 0x62, 0x2e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09,
	0x49, 0x6e, 0x64, 0x65, 0x78, 0x54, 0x79, 0x70, 0x65, 0x22, 0x62, 0x0a, 0x08, 0x52, 0x65, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x53, 0x48, 0x41, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x53, 0x48, 0x41, 0x12, 0x12, 0x0a, 0x04, 0x50, 0x61, 0x74, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x50, 0x61, 0x74, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x4e, 0x75, 0x6d,
	0x62, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x2a, 0x43, 0x0a,
	0x0a, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x4d,
	0x4f, 0x44, 0x49, 0x46, 0x59, 0x5f, 0x49, 0x4e, 0x44, 0x45, 0x58, 0x10, 0x00, 0x12, 0x10, 0x0a,
	0x0c, 0x44, 0x52, 0x4f, 0x50, 0x5f, 0x4f, 0x52, 0x50, 0x48, 0x41, 0x4e, 0x53, 0x10, 0x01, 0x12,
	0x11, 0x0a, 0x0d, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x5f, 0x52, 0x45, 0x43, 0x4f, 0x52, 0x44,
	0x10, 0x02, 0x2a, 0x4c, 0x0a, 0x09, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x06, 0x0a, 0x02, 0x56, 0x32, 0x10, 0x00, 0x12, 0x06, 0x0a, 0x02, 0x56, 0x31, 0x10, 0x01, 0x12,
	0x0d, 0x0a, 0x09, 0x46, 0x52, 0x41, 0x47, 0x4d, 0x45, 0x4e, 0x54, 0x53, 0x10, 0x02, 0x12, 0x13,
	0x0a, 0x0f, 0x44, 0x49, 0x47, 0x49, 0x54, 0x41, 0x4c, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54,
	0x53, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x47, 0x47, 0x45, 0x53, 0x54, 0x10, 0x04,
	0x42, 0x17, 0x5a, 0x15, 0x69, 0x6b, 0x75, 0x7a, 0x6f, 0x2f, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x2f, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_ikuzo_domain_domainpb_index_proto_rawDescOnce sync.Once
	file_ikuzo_domain_domainpb_index_proto_rawDescData = file_ikuzo_domain_domainpb_index_proto_rawDesc
)

func file_ikuzo_domain_domainpb_index_proto_rawDescGZIP() []byte {
	file_ikuzo_domain_domainpb_index_proto_rawDescOnce.Do(func() {
		file_ikuzo_domain_domainpb_index_proto_rawDescData = protoimpl.X.CompressGZIP(file_ikuzo_domain_domainpb_index_proto_rawDescData)
	})
	return file_ikuzo_domain_domainpb_index_proto_rawDescData
}

var file_ikuzo_domain_domainpb_index_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_ikuzo_domain_domainpb_index_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_ikuzo_domain_domainpb_index_proto_goTypes = []interface{}{
	(ActionType)(0),      // 0: domainpb.ActionType
	(IndexType)(0),       // 1: domainpb.IndexType
	(*IndexMessage)(nil), // 2: domainpb.IndexMessage
	(*Revision)(nil),     // 3: domainpb.Revision
}
var file_ikuzo_domain_domainpb_index_proto_depIdxs = []int32{
	3, // 0: domainpb.IndexMessage.Revision:type_name -> domainpb.Revision
	0, // 1: domainpb.IndexMessage.ActionType:type_name -> domainpb.ActionType
	1, // 2: domainpb.IndexMessage.IndexType:type_name -> domainpb.IndexType
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_ikuzo_domain_domainpb_index_proto_init() }
func file_ikuzo_domain_domainpb_index_proto_init() {
	if File_ikuzo_domain_domainpb_index_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ikuzo_domain_domainpb_index_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IndexMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ikuzo_domain_domainpb_index_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Revision); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ikuzo_domain_domainpb_index_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ikuzo_domain_domainpb_index_proto_goTypes,
		DependencyIndexes: file_ikuzo_domain_domainpb_index_proto_depIdxs,
		EnumInfos:         file_ikuzo_domain_domainpb_index_proto_enumTypes,
		MessageInfos:      file_ikuzo_domain_domainpb_index_proto_msgTypes,
	}.Build()
	File_ikuzo_domain_domainpb_index_proto = out.File
	file_ikuzo_domain_domainpb_index_proto_rawDesc = nil
	file_ikuzo_domain_domainpb_index_proto_goTypes = nil
	file_ikuzo_domain_domainpb_index_proto_depIdxs = nil
}
