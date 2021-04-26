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

package ead

import (
	"testing"

	"github.com/delving/hub3/hub3/ead/eadpb"
	"github.com/delving/hub3/hub3/fragments"
)

const metsTestFname = "testdata/c4209494-e7f7-44e7-9bde-c216249752ec_mets.xml"

func newTestCfg() (*NodeConfig, *fragments.Tree) {
	cfg := &NodeConfig{
		Counter:     &NodeCounter{},
		MetsCounter: &MetsCounter{},
		OrgID:       "NL-HaNA",
		Spec:        "1.04.18.03",
		Title:       []string{"long title"},
		TitleShort:  "short title",
		Revision:    0,
		PeriodDesc:  nil,
		MimeTypes:   map[string][]string{},
		Errors:      nil,
		// CreateTree:  CreateTree,
	}

	return cfg, &fragments.Tree{}
}

func TestFindingAidMimeTypes(t *testing.T) {
	mets, err := readMETS(metsTestFname)
	if err != nil || mets == nil {
		t.Errorf("unable to read mets file: %#v", err)
	}

	cfg, tree := newTestCfg()

	daoCfg := newDaoConfig(cfg, tree)

	findingAid, err := mets.newFindingAid(&daoCfg)
	if err != nil {
		t.Errorf("unable to create finding-aid: %#v", err)
	}

	counter := findingAid.GetMimeTypes()
	if len(counter) != 1 {
		t.Errorf("not all mimetypes added, got %d: want %d", len(counter), 1)
	}

	count, ok := counter["image/jpeg"]
	if !ok || count != 66 {
		t.Errorf("not all mimetypes counted, got %d: want %d", count, 66)
	}

	if len(getMimeTypes(&findingAid)) != 1 {
		t.Errorf("wrong count for mime-types, got %d: want %d", len(getMimeTypes(&findingAid)), 1)
	}

	if getMimeTypes(&findingAid)[0] != "image/jpeg" {
		t.Errorf("not the right mime-type array, got %#v", getMimeTypes(&findingAid))
	}
}

func TestReadMETS(t *testing.T) {
	mets, err := readMETS(metsTestFname)
	if err != nil || mets == nil {
		t.Errorf("unable to read mets file: %#v", err)
	}

	filesGroups := mets.CfileSec.CfileGrp
	if len(filesGroups) != 3 {
		t.Errorf("Not all filesecs are parsed, got %d: want %d", len(filesGroups), 2)
	}

	files := filesGroups[0].Cfile
	if len(files) != 66 {
		t.Errorf("Not all files are parsed, got %d: want %d", len(files), 66)
	}

	cfg, tree := newTestCfg()

	daoCfg := newDaoConfig(cfg, tree)

	findingAid, err := mets.newFindingAid(&daoCfg)
	if err != nil {
		t.Errorf("unable to create finding-aid: %#v", err)
	}

	if findingAid.GetFileCount() != 66 {
		t.Errorf("Not all entries are parsed, got %d: want %d", findingAid.GetFileCount(), 66)
	}

	tests := []struct {
		name        string
		file        *eadpb.File
		sortKey     int32
		fileSize    int32
		fileUUID    string
		deepzoomURI string
	}{
		{
			name:        "first",
			file:        findingAid.Files[0],
			sortKey:     1,
			fileSize:    int32(2878603),
			fileUUID:    "171ec840-1cf8-46d2-bfba-af8cf5e1239c",
			deepzoomURI: "https://service.archief.nl/iip/c4/20/94/94/e7/f7/44/e7/9b/de/c2/16/24/97/52/ec/171ec840-1cf8-46d2-bfba-af8cf5e1239c.jp2/info.json",
		},
		{
			name:        "last",
			file:        findingAid.Files[findingAid.GetFileCount()-1],
			sortKey:     66,
			fileSize:    int32(3171003),
			fileUUID:    "ae58d143-af1a-4639-85fd-936a196e864d",
			deepzoomURI: "https://service.archief.nl/iip/c4/20/94/94/e7/f7/44/e7/9b/de/c2/16/24/97/52/ec/ae58d143-af1a-4639-85fd-936a196e864d.jp2/info.json",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.file.GetSortKey() != tt.sortKey {
				t.Errorf("sorting is not correct. expected %d; got %d", tt.sortKey, tt.file.GetSortKey())
			}

			if tt.file.GetFileSize() != tt.fileSize {
				t.Errorf("size is not correct want %d; got %d", tt.fileSize, tt.file.GetFileSize())
			}

			if tt.file.GetFileuuid() != tt.fileUUID {
				t.Errorf("UUID is not correct, want %s; got %s", tt.fileUUID, tt.file.GetFileuuid())
			}

			if tt.file.GetDeepzoomURI() != tt.deepzoomURI {
				t.Errorf("DeepzoomURI is not correct, want %s; got %s", tt.deepzoomURI, tt.file.GetDeepzoomURI())
			}
		})
	}
}

func Test_createDeepZoomURI(t *testing.T) {
	type args struct {
		file  *eadpb.File
		duuid string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"prod environment simple uri",
			args{
				&eadpb.File{
					MimeType:     "image/jpg",
					Fileuuid:     "00000333-0000-0000-0810-000000000002",
					ThumbnailURI: "https://service.archief.nl/gaf/api/file/00000333-0000-0000-0810-000000000002",
				},
				"00000000000000000810000000000002",
			},
			"https://service.archief.nl/iip?IIIF=/00/00/00/00/00/00/00/00/08/10/00/00/00/" +
				"00/00/02/00000333-0000-0000-0810-000000000002.jp2/info.json",
		},
		{
			"test environment",
			args{
				&eadpb.File{
					MimeType:     "image/jpg",
					Fileuuid:     "1717314a-014c-465f-804a-0b5c72262240",
					ThumbnailURI: "https://service.test.archief.nl/gaf/api/file/v1/default/1717314a-014c-465f-804a-0b5c72262240",
				},
				"00000000000000000810000000000002",
			},
			"https://service.test.archief.nl/iip?IIIF=/00/00/00/00/00/00/00/00/08/10/00/00/00" +
				"/00/00/02/1717314a-014c-465f-804a-0b5c72262240.jp2/info.json",
		},
		{
			"acpt environment",
			args{
				&eadpb.File{
					MimeType:     "image/jpg",
					Fileuuid:     "1717314a-014c-465f-804a-0b5c72262240",
					ThumbnailURI: "https://service.acpt.archief.nl/gaf/api/file/v1/default/1717314a-014c-465f-804a-0b5c72262240",
				},
				"00000000000000000810000000000002",
			},
			"https://service.acpt.archief.nl/iip?IIIF=/00/00/00/00/00/00/00/00/08/10/00/00/" +
				"00/00/00/02/1717314a-014c-465f-804a-0b5c72262240.jp2/info.json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := createDeepZoomURI(tt.args.file, tt.args.duuid); got != tt.want {
				t.Errorf("createDeepZoomURI() = %v, want %v", got, tt.want)
			}
		})
	}
}
