package mapping

import "testing"

func Test_validate(t *testing.T) {
	type args struct {
		keys map[string]string
	}

	tests := []struct {
		name        string
		args        args
		wantOld     string
		wantCurrent string
		wantOk      bool
	}{
		{
			"mismatched",
			args{
				map[string]string{
					"123": "hello world",
				},
			},
			"123",
			"45ab6734b21e6968",
			false,
		},
		{
			"matched",
			args{
				map[string]string{
					"45ab6734b21e6968": "hello world",
				},
			},
			"",
			"",
			true,
		},
		{
			"v2 mapping",
			args{map[string]string{v2MappingSha: v2Mapping}},
			"",
			"",
			true,
		},
		{
			"v2 update mapping",
			args{map[string]string{v2UpdateMappingSha: v2MappingUpdate}},
			"",
			"",
			true,
		},
		{
			"fragment mapping",
			args{map[string]string{fragmentMappingSha: fragmentMapping}},
			"",
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			gotOld, gotCurrent, gotOk := validate(tt.args.keys)
			if gotOld != tt.wantOld {
				t.Errorf("validate() %s gotOld = %v, want %v", tt.name, gotOld, tt.wantOld)
			}
			if gotCurrent != tt.wantCurrent {
				t.Errorf("validate() %s gotCurrent  = %v, want %v", tt.name, gotCurrent, tt.wantCurrent)
			}
			if gotOk != tt.wantOk {
				t.Errorf("validate() %s gotOk = %v, want %v", tt.name, gotOk, tt.wantOk)
			}
		})
	}
}
