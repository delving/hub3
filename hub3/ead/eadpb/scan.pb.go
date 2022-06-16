// Copyright 2017 Delving B.V.
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
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: hub3/ead/eadpb/scan.proto

package eadpb

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

type ViewType int32

const (
	// Default type.
	ViewType_DOWNLOADONLY ViewType = 0
	// DZI tiles.
	ViewType_DZI ViewType = 1
	// JPEG files for thumbnails.
	ViewType_JPEG  ViewType = 2
	ViewType_PDF   ViewType = 3
	ViewType_AUDIO ViewType = 4
	ViewType_VIDEO ViewType = 5
)

// Enum value maps for ViewType.
var (
	ViewType_name = map[int32]string{
		0: "DOWNLOADONLY",
		1: "DZI",
		2: "JPEG",
		3: "PDF",
		4: "AUDIO",
		5: "VIDEO",
	}
	ViewType_value = map[string]int32{
		"DOWNLOADONLY": 0,
		"DZI":          1,
		"JPEG":         2,
		"PDF":          3,
		"AUDIO":        4,
		"VIDEO":        5,
	}
)

func (x ViewType) Enum() *ViewType {
	p := new(ViewType)
	*p = x
	return p
}

func (x ViewType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ViewType) Descriptor() protoreflect.EnumDescriptor {
	return file_hub3_ead_eadpb_scan_proto_enumTypes[0].Descriptor()
}

func (ViewType) Type() protoreflect.EnumType {
	return &file_hub3_ead_eadpb_scan_proto_enumTypes[0]
}

func (x ViewType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ViewType.Descriptor instead.
func (ViewType) EnumDescriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{0}
}

type File struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View ViewType `protobuf:"varint,1,opt,name=view,proto3,enum=eadpb.ViewType" json:"view,omitempty"`
	// For scans this is the NL-HaNA_1.01.02_11_P1 type filename.
	Filename string `protobuf:"bytes,3,opt,name=filename,proto3" json:"filename,omitempty"`
	// The file-uuid from GAF
	Fileuuid string `protobuf:"bytes,4,opt,name=fileuuid,proto3" json:"fileuuid,omitempty"`
	// The duuid is not needed for the frontend as of now I think, but could be useful.
	FileSize int32 `protobuf:"varint,6,opt,name=fileSize,proto3" json:"fileSize,omitempty"`
	// Absolute web accessible url to the file source so internet web clients can access them.
	// These urls are examples and not the real url patterns.
	// For DZI example: [https://test.nationaalarchief.nl/gaf/iip/1.04.02_1.dzi]
	// Or JPEG thumb example: [https://test.nationaalarchief.nl/gaf/iip/thumb-100x100.jpg]
	// Or PDF file example: [https://test.nationaalarchief.nl/gaf/file/1.04.02.pdf]
	ThumbnailURI string `protobuf:"bytes,2,opt,name=thumbnailURI,proto3" json:"thumbnailURI,omitempty"`
	// Optional uri for downloading the DZI XML description.
	DeepzoomURI string `protobuf:"bytes,7,opt,name=deepzoomURI,proto3" json:"deepzoomURI,omitempty"`
	// Optional uri for downloading the original.
	DownloadURI string `protobuf:"bytes,8,opt,name=downloadURI,proto3" json:"downloadURI,omitempty"`
	// Mime-Type of the file
	MimeType string `protobuf:"bytes,9,opt,name=mimeType,proto3" json:"mimeType,omitempty"`
	// The relative sort order of the File within the inventoryID
	SortKey int32 `protobuf:"varint,10,opt,name=sortKey,proto3" json:"sortKey,omitempty"`
	// Extra file metadata like collection, creator, date.
	MetaData map[string]*Values `protobuf:"bytes,11,rep,name=metaData,proto3" json:"metaData,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *File) Reset() {
	*x = File{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*File) ProtoMessage() {}

func (x *File) ProtoReflect() protoreflect.Message {
	mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use File.ProtoReflect.Descriptor instead.
func (*File) Descriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{0}
}

func (x *File) GetView() ViewType {
	if x != nil {
		return x.View
	}
	return ViewType_DOWNLOADONLY
}

func (x *File) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

func (x *File) GetFileuuid() string {
	if x != nil {
		return x.Fileuuid
	}
	return ""
}

func (x *File) GetFileSize() int32 {
	if x != nil {
		return x.FileSize
	}
	return 0
}

func (x *File) GetThumbnailURI() string {
	if x != nil {
		return x.ThumbnailURI
	}
	return ""
}

func (x *File) GetDeepzoomURI() string {
	if x != nil {
		return x.DeepzoomURI
	}
	return ""
}

func (x *File) GetDownloadURI() string {
	if x != nil {
		return x.DownloadURI
	}
	return ""
}

func (x *File) GetMimeType() string {
	if x != nil {
		return x.MimeType
	}
	return ""
}

func (x *File) GetSortKey() int32 {
	if x != nil {
		return x.SortKey
	}
	return 0
}

func (x *File) GetMetaData() map[string]*Values {
	if x != nil {
		return x.MetaData
	}
	return nil
}

type Values struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The display label
	Label string `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	// List of metadata string values
	Text []string `protobuf:"bytes,2,rep,name=text,proto3" json:"text,omitempty"`
}

func (x *Values) Reset() {
	*x = Values{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Values) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Values) ProtoMessage() {}

func (x *Values) ProtoReflect() protoreflect.Message {
	mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Values.ProtoReflect.Descriptor instead.
func (*Values) Descriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{1}
}

func (x *Values) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *Values) GetText() []string {
	if x != nil {
		return x.Text
	}
	return nil
}

type Pager struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HasNext      bool  `protobuf:"varint,1,opt,name=hasNext,proto3" json:"hasNext,omitempty"`
	HasPrevious  bool  `protobuf:"varint,2,opt,name=hasPrevious,proto3" json:"hasPrevious,omitempty"`
	TotalCount   int32 `protobuf:"varint,3,opt,name=totalCount,proto3" json:"totalCount,omitempty"`
	NrPages      int32 `protobuf:"varint,4,opt,name=nrPages,proto3" json:"nrPages,omitempty"`
	PageCurrent  int32 `protobuf:"varint,5,opt,name=pageCurrent,proto3" json:"pageCurrent,omitempty"`
	PageNext     int32 `protobuf:"varint,6,opt,name=pageNext,proto3" json:"pageNext,omitempty"`
	PagePrevious int32 `protobuf:"varint,7,opt,name=pagePrevious,proto3" json:"pagePrevious,omitempty"`
	PageSize     int32 `protobuf:"varint,8,opt,name=pageSize,proto3" json:"pageSize,omitempty"`
	// Optional property where the result is centered around.
	ActiveFilename string `protobuf:"bytes,9,opt,name=activeFilename,proto3" json:"activeFilename,omitempty"`
	// The index of the place of the active File in the Files array.
	ActiveSortKey int32 `protobuf:"varint,10,opt,name=activeSortKey,proto3" json:"activeSortKey,omitempty"`
	// Optional. When the request was a paging request, i.e. without the FindingAid block.
	Paging bool `protobuf:"varint,11,opt,name=paging,proto3" json:"paging,omitempty"`
}

func (x *Pager) Reset() {
	*x = Pager{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Pager) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pager) ProtoMessage() {}

func (x *Pager) ProtoReflect() protoreflect.Message {
	mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Pager.ProtoReflect.Descriptor instead.
func (*Pager) Descriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{2}
}

func (x *Pager) GetHasNext() bool {
	if x != nil {
		return x.HasNext
	}
	return false
}

func (x *Pager) GetHasPrevious() bool {
	if x != nil {
		return x.HasPrevious
	}
	return false
}

func (x *Pager) GetTotalCount() int32 {
	if x != nil {
		return x.TotalCount
	}
	return 0
}

func (x *Pager) GetNrPages() int32 {
	if x != nil {
		return x.NrPages
	}
	return 0
}

func (x *Pager) GetPageCurrent() int32 {
	if x != nil {
		return x.PageCurrent
	}
	return 0
}

func (x *Pager) GetPageNext() int32 {
	if x != nil {
		return x.PageNext
	}
	return 0
}

func (x *Pager) GetPagePrevious() int32 {
	if x != nil {
		return x.PagePrevious
	}
	return 0
}

func (x *Pager) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *Pager) GetActiveFilename() string {
	if x != nil {
		return x.ActiveFilename
	}
	return ""
}

func (x *Pager) GetActiveSortKey() int32 {
	if x != nil {
		return x.ActiveSortKey
	}
	return 0
}

func (x *Pager) GetPaging() bool {
	if x != nil {
		return x.Paging
	}
	return false
}

type FindingAid struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The dataset identifier and EAD identifier.
	ArchiveID string `protobuf:"bytes,1,opt,name=archiveID,proto3" json:"archiveID,omitempty"`
	// The long title for the Archive.
	ArchiveTitle string `protobuf:"bytes,2,opt,name=archiveTitle,proto3" json:"archiveTitle,omitempty"`
	// The unit idendifier for a given c-level.
	InventoryID string `protobuf:"bytes,3,opt,name=inventoryID,proto3" json:"inventoryID,omitempty"`
	// The tree-path under which the inventory is stored.
	InventoryPath string `protobuf:"bytes,4,opt,name=inventoryPath,proto3" json:"inventoryPath,omitempty"`
	// The unit-title of the inventory.
	InventoryTitle string `protobuf:"bytes,5,opt,name=inventoryTitle,proto3" json:"inventoryTitle,omitempty"`
	// The deliverable uuid of the METS file where all the Files are extracted from.
	Duuid string `protobuf:"bytes,6,opt,name=duuid,proto3" json:"duuid,omitempty"`
	// Return true if the files in the WHOLE (not just the current page) set are DZI tiles.
	HasOnlyTiles bool `protobuf:"varint,7,opt,name=hasOnlyTiles,proto3" json:"hasOnlyTiles,omitempty"`
	// Sorted array of mime-types for current deliverable-uuid
	MimeTypes map[string]int32 `protobuf:"bytes,8,rep,name=mimeTypes,proto3" json:"mimeTypes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// number of linked digital objects
	FileCount int32 `protobuf:"varint,9,opt,name=fileCount,proto3" json:"fileCount,omitempty"`
	// the linked Files to the FindingAid
	Files []*File `protobuf:"bytes,10,rep,name=files,proto3" json:"files,omitempty"`
	// filter keys
	FilterTypes []string `protobuf:"bytes,11,rep,name=filterTypes,proto3" json:"filterTypes,omitempty"`
	// scan navigation
	HasScanNavigation bool `protobuf:"varint,12,opt,name=hasScanNavigation,proto3" json:"hasScanNavigation,omitempty"`
}

func (x *FindingAid) Reset() {
	*x = FindingAid{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindingAid) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindingAid) ProtoMessage() {}

func (x *FindingAid) ProtoReflect() protoreflect.Message {
	mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindingAid.ProtoReflect.Descriptor instead.
func (*FindingAid) Descriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{3}
}

func (x *FindingAid) GetArchiveID() string {
	if x != nil {
		return x.ArchiveID
	}
	return ""
}

func (x *FindingAid) GetArchiveTitle() string {
	if x != nil {
		return x.ArchiveTitle
	}
	return ""
}

func (x *FindingAid) GetInventoryID() string {
	if x != nil {
		return x.InventoryID
	}
	return ""
}

func (x *FindingAid) GetInventoryPath() string {
	if x != nil {
		return x.InventoryPath
	}
	return ""
}

func (x *FindingAid) GetInventoryTitle() string {
	if x != nil {
		return x.InventoryTitle
	}
	return ""
}

func (x *FindingAid) GetDuuid() string {
	if x != nil {
		return x.Duuid
	}
	return ""
}

func (x *FindingAid) GetHasOnlyTiles() bool {
	if x != nil {
		return x.HasOnlyTiles
	}
	return false
}

func (x *FindingAid) GetMimeTypes() map[string]int32 {
	if x != nil {
		return x.MimeTypes
	}
	return nil
}

func (x *FindingAid) GetFileCount() int32 {
	if x != nil {
		return x.FileCount
	}
	return 0
}

func (x *FindingAid) GetFiles() []*File {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *FindingAid) GetFilterTypes() []string {
	if x != nil {
		return x.FilterTypes
	}
	return nil
}

func (x *FindingAid) GetHasScanNavigation() bool {
	if x != nil {
		return x.HasScanNavigation
	}
	return false
}

type ViewResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The block with the page
	Pager *Pager `protobuf:"bytes,1,opt,name=pager,proto3" json:"pager,omitempty"`
	// Optional. The FindingAid is only shown on the first request and is empty for paging requests.
	FindingAid *FindingAid `protobuf:"bytes,2,opt,name=findingAid,proto3" json:"findingAid,omitempty"`
	// Sorted by filename.
	Files []*File `protobuf:"bytes,3,rep,name=files,proto3" json:"files,omitempty"`
}

func (x *ViewResponse) Reset() {
	*x = ViewResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ViewResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ViewResponse) ProtoMessage() {}

func (x *ViewResponse) ProtoReflect() protoreflect.Message {
	mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ViewResponse.ProtoReflect.Descriptor instead.
func (*ViewResponse) Descriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{4}
}

func (x *ViewResponse) GetPager() *Pager {
	if x != nil {
		return x.Pager
	}
	return nil
}

func (x *ViewResponse) GetFindingAid() *FindingAid {
	if x != nil {
		return x.FindingAid
	}
	return nil
}

func (x *ViewResponse) GetFiles() []*File {
	if x != nil {
		return x.Files
	}
	return nil
}

type ViewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required; example 2.13.39
	ArchiveID string `protobuf:"bytes,1,opt,name=archiveID,proto3" json:"archiveID,omitempty"`
	// Either InventoryID or InventoryPath is Required; e.g. 3.1.
	InventoryID string `protobuf:"bytes,2,opt,name=inventoryID,proto3" json:"inventoryID,omitempty"`
	// Either InventoryID or InventoryPath is Required; e.g. @3~3.1.
	InvertoryPath string `protobuf:"bytes,3,opt,name=invertoryPath,proto3" json:"invertoryPath,omitempty"`
	// Optional, defaults to 50 tiledImages per set.
	PageSize int32 `protobuf:"varint,4,opt,name=pageSize,proto3" json:"pageSize,omitempty"`
	// Optional, defaults to 1.
	Page int32 `protobuf:"varint,5,opt,name=page,proto3" json:"page,omitempty"`
	// Optional to jump into a set with this file centered.
	Fileuuid string `protobuf:"bytes,6,opt,name=fileuuid,proto3" json:"fileuuid,omitempty"`
	// Optional to jump into a set with this file centered.
	Filename string `protobuf:"bytes,7,opt,name=filename,proto3" json:"filename,omitempty"`
	// Optional thumbnail configuration options.
	ThumbnailConf string `protobuf:"bytes,8,opt,name=thumbnailConf,proto3" json:"thumbnailConf,omitempty"`
	// Optional. When paging is true the FindingAid block is not included in the response.
	Paging bool `protobuf:"varint,9,opt,name=paging,proto3" json:"paging,omitempty"`
	// Optional. sortKey is an integer representing the order of the file in the files array.
	// Note: starts at 1, i.e. not zero-based.
	SortKey int64 `protobuf:"varint,10,opt,name=sortKey,proto3" json:"sortKey,omitempty"`
}

func (x *ViewRequest) Reset() {
	*x = ViewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ViewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ViewRequest) ProtoMessage() {}

func (x *ViewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_hub3_ead_eadpb_scan_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ViewRequest.ProtoReflect.Descriptor instead.
func (*ViewRequest) Descriptor() ([]byte, []int) {
	return file_hub3_ead_eadpb_scan_proto_rawDescGZIP(), []int{5}
}

func (x *ViewRequest) GetArchiveID() string {
	if x != nil {
		return x.ArchiveID
	}
	return ""
}

func (x *ViewRequest) GetInventoryID() string {
	if x != nil {
		return x.InventoryID
	}
	return ""
}

func (x *ViewRequest) GetInvertoryPath() string {
	if x != nil {
		return x.InvertoryPath
	}
	return ""
}

func (x *ViewRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ViewRequest) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *ViewRequest) GetFileuuid() string {
	if x != nil {
		return x.Fileuuid
	}
	return ""
}

func (x *ViewRequest) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

func (x *ViewRequest) GetThumbnailConf() string {
	if x != nil {
		return x.ThumbnailConf
	}
	return ""
}

func (x *ViewRequest) GetPaging() bool {
	if x != nil {
		return x.Paging
	}
	return false
}

func (x *ViewRequest) GetSortKey() int64 {
	if x != nil {
		return x.SortKey
	}
	return 0
}

var File_hub3_ead_eadpb_scan_proto protoreflect.FileDescriptor

var file_hub3_ead_eadpb_scan_proto_rawDesc = []byte{
	0x0a, 0x19, 0x68, 0x75, 0x62, 0x33, 0x2f, 0x65, 0x61, 0x64, 0x2f, 0x65, 0x61, 0x64, 0x70, 0x62,
	0x2f, 0x73, 0x63, 0x61, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x65, 0x61, 0x64,
	0x70, 0x62, 0x22, 0xa0, 0x03, 0x0a, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x23, 0x0a, 0x04, 0x76,
	0x69, 0x65, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0f, 0x2e, 0x65, 0x61, 0x64, 0x70,
	0x62, 0x2e, 0x56, 0x69, 0x65, 0x77, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77,
	0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x75, 0x75, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x75, 0x75, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x53, 0x69, 0x7a, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x53, 0x69, 0x7a, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x74, 0x68, 0x75, 0x6d, 0x62, 0x6e, 0x61, 0x69,
	0x6c, 0x55, 0x52, 0x49, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x74, 0x68, 0x75, 0x6d,
	0x62, 0x6e, 0x61, 0x69, 0x6c, 0x55, 0x52, 0x49, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x65, 0x70,
	0x7a, 0x6f, 0x6f, 0x6d, 0x55, 0x52, 0x49, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64,
	0x65, 0x65, 0x70, 0x7a, 0x6f, 0x6f, 0x6d, 0x55, 0x52, 0x49, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x55, 0x52, 0x49, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x55, 0x52, 0x49, 0x12, 0x1a, 0x0a, 0x08,
	0x6d, 0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x6d, 0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x6f, 0x72, 0x74,
	0x4b, 0x65, 0x79, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x73, 0x6f, 0x72, 0x74, 0x4b,
	0x65, 0x79, 0x12, 0x35, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x44, 0x61, 0x74, 0x61, 0x18, 0x0b,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6c,
	0x65, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x4a, 0x0a, 0x0d, 0x4d, 0x65, 0x74,
	0x61, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x23, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x65, 0x61,
	0x64, 0x70, 0x62, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x32, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x22, 0xe1, 0x02, 0x0a, 0x05, 0x50, 0x61,
	0x67, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x68, 0x61, 0x73, 0x4e, 0x65, 0x78, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x68, 0x61, 0x73, 0x4e, 0x65, 0x78, 0x74, 0x12, 0x20, 0x0a,
	0x0b, 0x68, 0x61, 0x73, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0b, 0x68, 0x61, 0x73, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x12,
	0x1e, 0x0a, 0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x18, 0x0a, 0x07, 0x6e, 0x72, 0x50, 0x61, 0x67, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x07, 0x6e, 0x72, 0x50, 0x61, 0x67, 0x65, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x61, 0x67,
	0x65, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b,
	0x70, 0x61, 0x67, 0x65, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x61, 0x67, 0x65, 0x4e, 0x65, 0x78, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70,
	0x61, 0x67, 0x65, 0x4e, 0x65, 0x78, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x61, 0x67, 0x65, 0x50,
	0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x70,
	0x61, 0x67, 0x65, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70,
	0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x46, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x24, 0x0a, 0x0d, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x53, 0x6f, 0x72, 0x74, 0x4b, 0x65, 0x79,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x53, 0x6f,
	0x72, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x22, 0x87, 0x04,
	0x0a, 0x0a, 0x46, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x41, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09,
	0x61, 0x72, 0x63, 0x68, 0x69, 0x76, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x61, 0x72, 0x63, 0x68, 0x69, 0x76, 0x65, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x72,
	0x63, 0x68, 0x69, 0x76, 0x65, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x61, 0x72, 0x63, 0x68, 0x69, 0x76, 0x65, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x49, 0x44, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x49, 0x44,
	0x12, 0x24, 0x0a, 0x0d, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x50, 0x61, 0x74,
	0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f,
	0x72, 0x79, 0x50, 0x61, 0x74, 0x68, 0x12, 0x26, 0x0a, 0x0e, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74,
	0x6f, 0x72, 0x79, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x64, 0x75, 0x75, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x64,
	0x75, 0x75, 0x69, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x68, 0x61, 0x73, 0x4f, 0x6e, 0x6c, 0x79, 0x54,
	0x69, 0x6c, 0x65, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x68, 0x61, 0x73, 0x4f,
	0x6e, 0x6c, 0x79, 0x54, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x3e, 0x0a, 0x09, 0x6d, 0x69, 0x6d, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x65, 0x61,
	0x64, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x41, 0x69, 0x64, 0x2e, 0x4d,
	0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x6d,
	0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x66, 0x69, 0x6c,
	0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x21, 0x0a, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18,
	0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x46, 0x69,
	0x6c, 0x65, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x66, 0x69, 0x6c,
	0x74, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x73, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x73, 0x12, 0x2c, 0x0a, 0x11, 0x68,
	0x61, 0x73, 0x53, 0x63, 0x61, 0x6e, 0x4e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x0c, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x68, 0x61, 0x73, 0x53, 0x63, 0x61, 0x6e, 0x4e,
	0x61, 0x76, 0x69, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x3c, 0x0a, 0x0e, 0x4d, 0x69, 0x6d,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x88, 0x01, 0x0a, 0x0c, 0x56, 0x69, 0x65, 0x77,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x05, 0x70, 0x61, 0x67, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e,
	0x50, 0x61, 0x67, 0x65, 0x72, 0x52, 0x05, 0x70, 0x61, 0x67, 0x65, 0x72, 0x12, 0x31, 0x0a, 0x0a,
	0x66, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x41, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67,
	0x41, 0x69, 0x64, 0x52, 0x0a, 0x66, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x41, 0x69, 0x64, 0x12,
	0x21, 0x0a, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b,
	0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x05, 0x66, 0x69, 0x6c,
	0x65, 0x73, 0x22, 0xb3, 0x02, 0x0a, 0x0b, 0x56, 0x69, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x72, 0x63, 0x68, 0x69, 0x76, 0x65, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x72, 0x63, 0x68, 0x69, 0x76, 0x65, 0x49, 0x44,
	0x12, 0x20, 0x0a, 0x0b, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x49, 0x44, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79,
	0x49, 0x44, 0x12, 0x24, 0x0a, 0x0d, 0x69, 0x6e, 0x76, 0x65, 0x72, 0x74, 0x6f, 0x72, 0x79, 0x50,
	0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x69, 0x6e, 0x76, 0x65, 0x72,
	0x74, 0x6f, 0x72, 0x79, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x67, 0x65,
	0x53, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65,
	0x53, 0x69, 0x7a, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x75, 0x75, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x75, 0x75, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x24, 0x0a, 0x0d, 0x74, 0x68, 0x75, 0x6d, 0x62, 0x6e, 0x61, 0x69, 0x6c, 0x43, 0x6f, 0x6e,
	0x66, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x68, 0x75, 0x6d, 0x62, 0x6e, 0x61,
	0x69, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x12, 0x18,
	0x0a, 0x07, 0x73, 0x6f, 0x72, 0x74, 0x4b, 0x65, 0x79, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x07, 0x73, 0x6f, 0x72, 0x74, 0x4b, 0x65, 0x79, 0x2a, 0x4e, 0x0a, 0x08, 0x56, 0x69, 0x65, 0x77,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x44, 0x4f, 0x57, 0x4e, 0x4c, 0x4f, 0x41, 0x44,
	0x4f, 0x4e, 0x4c, 0x59, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x44, 0x5a, 0x49, 0x10, 0x01, 0x12,
	0x08, 0x0a, 0x04, 0x4a, 0x50, 0x45, 0x47, 0x10, 0x02, 0x12, 0x07, 0x0a, 0x03, 0x50, 0x44, 0x46,
	0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x41, 0x55, 0x44, 0x49, 0x4f, 0x10, 0x04, 0x12, 0x09, 0x0a,
	0x05, 0x56, 0x49, 0x44, 0x45, 0x4f, 0x10, 0x05, 0x32, 0x40, 0x0a, 0x0d, 0x56, 0x69, 0x65, 0x77,
	0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x2f, 0x0a, 0x04, 0x4c, 0x69, 0x73,
	0x74, 0x12, 0x12, 0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x56, 0x69, 0x65, 0x77, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x65, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x56, 0x69,
	0x65, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x10, 0x5a, 0x0e, 0x68, 0x75,
	0x62, 0x33, 0x2f, 0x65, 0x61, 0x64, 0x2f, 0x65, 0x61, 0x64, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_hub3_ead_eadpb_scan_proto_rawDescOnce sync.Once
	file_hub3_ead_eadpb_scan_proto_rawDescData = file_hub3_ead_eadpb_scan_proto_rawDesc
)

func file_hub3_ead_eadpb_scan_proto_rawDescGZIP() []byte {
	file_hub3_ead_eadpb_scan_proto_rawDescOnce.Do(func() {
		file_hub3_ead_eadpb_scan_proto_rawDescData = protoimpl.X.CompressGZIP(file_hub3_ead_eadpb_scan_proto_rawDescData)
	})
	return file_hub3_ead_eadpb_scan_proto_rawDescData
}

var file_hub3_ead_eadpb_scan_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_hub3_ead_eadpb_scan_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_hub3_ead_eadpb_scan_proto_goTypes = []interface{}{
	(ViewType)(0),        // 0: eadpb.ViewType
	(*File)(nil),         // 1: eadpb.File
	(*Values)(nil),       // 2: eadpb.Values
	(*Pager)(nil),        // 3: eadpb.Pager
	(*FindingAid)(nil),   // 4: eadpb.FindingAid
	(*ViewResponse)(nil), // 5: eadpb.ViewResponse
	(*ViewRequest)(nil),  // 6: eadpb.ViewRequest
	nil,                  // 7: eadpb.File.MetaDataEntry
	nil,                  // 8: eadpb.FindingAid.MimeTypesEntry
}
var file_hub3_ead_eadpb_scan_proto_depIdxs = []int32{
	0, // 0: eadpb.File.view:type_name -> eadpb.ViewType
	7, // 1: eadpb.File.metaData:type_name -> eadpb.File.MetaDataEntry
	8, // 2: eadpb.FindingAid.mimeTypes:type_name -> eadpb.FindingAid.MimeTypesEntry
	1, // 3: eadpb.FindingAid.files:type_name -> eadpb.File
	3, // 4: eadpb.ViewResponse.pager:type_name -> eadpb.Pager
	4, // 5: eadpb.ViewResponse.findingAid:type_name -> eadpb.FindingAid
	1, // 6: eadpb.ViewResponse.files:type_name -> eadpb.File
	2, // 7: eadpb.File.MetaDataEntry.value:type_name -> eadpb.Values
	6, // 8: eadpb.ViewerService.List:input_type -> eadpb.ViewRequest
	5, // 9: eadpb.ViewerService.List:output_type -> eadpb.ViewResponse
	9, // [9:10] is the sub-list for method output_type
	8, // [8:9] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_hub3_ead_eadpb_scan_proto_init() }
func file_hub3_ead_eadpb_scan_proto_init() {
	if File_hub3_ead_eadpb_scan_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_hub3_ead_eadpb_scan_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*File); i {
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
		file_hub3_ead_eadpb_scan_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Values); i {
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
		file_hub3_ead_eadpb_scan_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Pager); i {
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
		file_hub3_ead_eadpb_scan_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindingAid); i {
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
		file_hub3_ead_eadpb_scan_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ViewResponse); i {
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
		file_hub3_ead_eadpb_scan_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ViewRequest); i {
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
			RawDescriptor: file_hub3_ead_eadpb_scan_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_hub3_ead_eadpb_scan_proto_goTypes,
		DependencyIndexes: file_hub3_ead_eadpb_scan_proto_depIdxs,
		EnumInfos:         file_hub3_ead_eadpb_scan_proto_enumTypes,
		MessageInfos:      file_hub3_ead_eadpb_scan_proto_msgTypes,
	}.Build()
	File_hub3_ead_eadpb_scan_proto = out.File
	file_hub3_ead_eadpb_scan_proto_rawDesc = nil
	file_hub3_ead_eadpb_scan_proto_goTypes = nil
	file_hub3_ead_eadpb_scan_proto_depIdxs = nil
}
