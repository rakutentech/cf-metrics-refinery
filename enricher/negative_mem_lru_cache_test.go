package enricher

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type nme struct {
	err error
}

var mockMeta = AppMetadata{
	App:       "app",
	AppGUID:   "app_guid",
	Space:     "space",
	SpaceGUID: "space_guid",
	Org:       "org",
	OrgGUID:   "org_guid",
}

func (n *nme) GetAppMetadata(appGUID string) (AppMetadata, error) {
	if n.err != nil {
		return AppMetadata{}, n.err
	}
	return mockMeta, nil
}

func mockNegativeCache(appGUID ...string) map[string]*appNotFound {
	nm := make(map[string]*appNotFound, len(appGUID))
	for _, a := range appGUID {
		nm[a] = &appNotFound{time.Now()}
	}
	return nm
}

func TestNegativeMemLRUCache_GetAppMetadata(t *testing.T) {
	errNotFound := errors.New("CF-AppNotFound")
	errOther := errors.New("Don't move. Error!")

	type fields struct {
		cache  map[string]*appNotFound
		parent Enricher
	}
	type args struct {
		appGUID string
	}
	tests := []struct {
		name                    string
		fields                  fields
		args                    args
		wantMeta                AppMetadata
		wantNegativeCacheLength int
		wantErr                 bool
	}{
		{"hit", fields{parent: &nme{}, cache: mockNegativeCache("app_guid")}, args{"app_guid"}, AppMetadata{}, 1, true},
		{"miss && get metadata", fields{parent: &nme{}}, args{"app_guid"}, mockMeta, 0, false},
		{"miss && err: app not found", fields{parent: &nme{err: errNotFound}}, args{"app_guid"}, AppMetadata{}, 1, true},
		{"miss && other errors", fields{parent: &nme{err: errOther}}, args{"app_guid"}, AppMetadata{}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.cache == nil {
				tt.fields.cache = NewNegativeMemLRUCache(tt.fields.parent).(*NegativeMemLRUCache).cache
			}
			n := NewNegativeMemLRUCache(tt.fields.parent).(*NegativeMemLRUCache)
			n.cache = tt.fields.cache
			gotMeta, err := n.GetAppMetadata(tt.args.appGUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMeta, tt.wantMeta) {
				t.Errorf("expect AppMetadata %v, got %v", tt.wantMeta, gotMeta)
			}
			if len(n.cache) != tt.wantNegativeCacheLength {
				t.Errorf("expect NegativeMemLRUCache length %v, got %v", tt.wantNegativeCacheLength, len(n.cache))
			}
		})
	}
}

func TestNegativeMemLRUCache_Expire(t *testing.T) {
	tests := []struct {
		name            string
		olderThan       time.Duration
		wantCacheLength int
	}{
		{"Cache doesn't expire", 10 * time.Minute, 1},
		{"Cache expire", 1 * time.Millisecond, 0},
	}

	for _, test := range tests {
		nm := NewNegativeMemLRUCache(nil).(*NegativeMemLRUCache)
		nm.cache["guid"] = &appNotFound{lastNotFound: time.Now()}

		time.Sleep(3 * time.Millisecond)
		nm.Expire(test.olderThan)
		if len(nm.cache) != test.wantCacheLength {
			t.Fatalf("expected cache length %d, got %d", test.wantCacheLength, len(nm.cache))
		}
	}
}

func TestNegativeMemLRUCache_Warmup(t *testing.T) {
	nm := NewNegativeMemLRUCache(nil).(*NegativeMemLRUCache)
	nm.cache["app_guid"] = &appNotFound{lastNotFound: time.Now()}

	tests := []struct {
		name            string
		negativecache   *NegativeMemLRUCache
		mds             []AppMetadata
		wantCacheLength int
	}{
		{"Negative cache doesn't expire", nm, nil, 1},
		{"Negative cache expire", nm, []AppMetadata{mockMeta}, 0},
	}

	for _, test := range tests {
		nm.Warmup(test.mds)
		if len(nm.cache) != test.wantCacheLength {
			t.Fatalf("expected cache length %d, got %d", test.wantCacheLength, len(nm.cache))
		}
	}
}
