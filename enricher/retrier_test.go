package enricher

import (
	"errors"
	"reflect"
	"testing"
)

type FailingEnricher struct {
	failures int
}

func (e *FailingEnricher) GetAppMetadata(app_guid string) (AppMetadata, error) {
	if e.failures > 0 {
		e.failures--
		return AppMetadata{}, errors.New("so much fail")
	}
	return AppMetadata{App: app_guid, AppGUID: app_guid}, nil
}

func TestRetrier_GetAppMetadata(t *testing.T) {
	type fields struct {
		parent Enricher
	}
	type args struct {
		app_guid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    AppMetadata
		wantErr bool
	}{
		{"fail always", fields{&FailingEnricher{1 << 30}}, args{"guid"}, AppMetadata{}, true},
		{"fail 3 times", fields{&FailingEnricher{3}}, args{"guid"}, AppMetadata{}, true},
		{"fail 2 times", fields{&FailingEnricher{2}}, args{"guid"}, AppMetadata{App: "guid", AppGUID: "guid"}, false},
		{"fail 0 times", fields{&FailingEnricher{0}}, args{"guid"}, AppMetadata{App: "guid", AppGUID: "guid"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewRetrier(tt.fields.parent)
			got, err := e.GetAppMetadata(tt.args.app_guid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrier.GetAppMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Retrier.GetAppMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
