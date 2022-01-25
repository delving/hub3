package elasticsearch

import "testing"

func TestConfig_Valid(t *testing.T) {
	type fields struct {
		Urls     []string
		UserName string
		Password string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "empty urls array",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "valid entry",
			fields: fields{
				Urls: []string{"http://localhost:9200"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				Urls:     tt.fields.Urls,
				UserName: tt.fields.UserName,
				Password: tt.fields.Password,
			}

			err := c.Valid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Valid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
