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

package imageproxy

import (
	"testing"
)

func TestService_portsAllowed(t *testing.T) {
	type fields struct {
		allowPorts []string
	}

	type args struct {
		targetURL string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			"empty url",
			fields{allowPorts: []string{"80", "443"}},
			args{targetURL: "https://www.example.com"},
			true,
			false,
		},
		{
			"explicit port",
			fields{allowPorts: []string{"80", "443"}},
			args{targetURL: "https://www.example.com:443"},
			true,
			false,
		},
		{
			"invalid port",
			fields{allowPorts: []string{"80", "443"}},
			args{targetURL: "https://www.example.com:943"},
			false,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				allowPorts: tt.fields.allowPorts,
			}
			got, err := s.portsAllowed(tt.args.targetURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.portsAllowed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Service.portsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
