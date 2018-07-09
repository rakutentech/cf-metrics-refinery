package enricher

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type me struct {
	m map[string]AppMetadata
}

func (m *me) GetAppMetadata(appGUID string) (AppMetadata, error) {
	if md, ok := m.m[appGUID]; ok {
		return md, nil
	} else {
		return AppMetadata{}, errors.New("no such app")
	}
}

func mockEnricher(appGUID ...string) Enricher {
	m := make(map[string]AppMetadata, len(appGUID))
	for _, a := range appGUID {
		m[a] = mockData(a)
	}
	return &me{m}
}

func mockData(a string) AppMetadata {
	return AppMetadata{
		App:       "app" + a,
		AppGUID:   a,
		Space:     "space" + a,
		SpaceGUID: a,
		Org:       "org" + a,
		OrgGUID:   a,
	}
}

func mockCache(appGUID ...string) map[string]*appMetadata {
	m := make(map[string]*appMetadata, len(appGUID))
	for _, a := range appGUID {
		m[a] = &appMetadata{mockData(a), time.Now()}
	}
	return m
}

func TestMemLRUCache_GetAppMetadata(t *testing.T) {
	type fields struct {
		cache  map[string]*appMetadata
		parent Enricher
	}
	type args struct {
		appGUID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    AppMetadata
		wantErr bool
	}{
		{"empty miss", fields{parent: mockEnricher()}, args{"guid1"}, AppMetadata{}, true},
		{"hit", fields{parent: mockEnricher("guid1")}, args{"guid1"}, mockData("guid1"), false},
		{"miss", fields{parent: mockEnricher("guid1")}, args{"guid2"}, AppMetadata{}, true},
		{"hit twice", fields{parent: mockEnricher(), cache: mockCache("guid1")}, args{"guid1"}, mockData("guid1"), false},
		{"miss twice", fields{parent: mockEnricher("guid1"), cache: mockCache("guid1")}, args{"guid2"}, AppMetadata{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.cache == nil {
				tt.fields.cache = NewMemLRUCache(tt.fields.parent).(*MemLRUCache).cache
			}

			e := NewMemLRUCache(tt.fields.parent).(*MemLRUCache)
			e.cache = tt.fields.cache

			got, err := e.GetAppMetadata(tt.args.appGUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemLRUCache.GetAppMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemLRUCache.GetAppMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemLRUCache_Expire(t *testing.T) {
	tests := []struct {
		name            string
		olderThan       time.Duration
		wantCacheLength int
	}{
		{"Cache doesn't expire", 10 * time.Minute, 1},
		{"Cache expire", 1 * time.Millisecond, 0},
	}

	for _, test := range tests {
		appMeta := &appMetadata{
			AppMetadata: mockData("guid1"),
			lastSeen:    time.Now(),
		}

		em := NewMemLRUCache(nil).(*MemLRUCache)
		em.cache[appMeta.AppGUID] = appMeta

		time.Sleep(3 * time.Millisecond)
		em.Expire(test.olderThan)
		if !reflect.DeepEqual(len(em.cache), test.wantCacheLength) {
			t.Fatalf("TestMemLRUCache_Expire: expected cache length %d, got %d", test.wantCacheLength, len(em.cache))
		}
	}
}

func TestMemLRUCache_WarmupEmptyCache(t *testing.T) {
	tests := []struct {
		name            string
		mds             []AppMetadata
		wantErr         bool
		wantCacheLength int
	}{
		{"Warmup fetching apps metadata Error", nil, true, 0},
		{"Warmup empty cache successfully", []AppMetadata{mockData("guid1")}, false, 1},
	}

	for _, test := range tests {
		em := NewMemLRUCache(nil).(*MemLRUCache)

		em.Warmup(test.mds)
		if len(em.cache) != test.wantCacheLength {
			t.Fatalf("TestMemLRUCache_Warmup %s: expected cache length %d, got %d", test.name, test.wantCacheLength, len(em.cache))
		}
	}
}

func TestMemLRUCache_WarmupExsitingCache(t *testing.T) {
	wantAppMetadataGUID0 := mockData("guid0")
	wantAppMetadataGUID1 := mockData("guid1")
	wantAppMetadataGUID2 := mockData("guid2")
	var AppMeta = AppMetadata{
		App:       "guid1",
		AppGUID:   "00000000-0000-0000-0000-000000000000",
		Space:     "space",
		SpaceGUID: "10000000-0000-0000-0000-000000000000",
		Org:       "org",
		OrgGUID:   "20000000-0000-0000-0000-000000000000",
	}
	em := NewMemLRUCache(nil).(*MemLRUCache)
	em.cache["guid1"] = &appMetadata{AppMetadata: AppMeta}
	em.cache["guid0"] = &appMetadata{AppMetadata: wantAppMetadataGUID0}

	em.Warmup([]AppMetadata{wantAppMetadataGUID1, wantAppMetadataGUID2})
	if !reflect.DeepEqual(em.cache["guid0"].AppMetadata, wantAppMetadataGUID0) {
		t.Fatalf("TestMemLRUCache_WarmupExsitingCache: expect guid0 Metadata %v, got %v \n", wantAppMetadataGUID0, em.cache["guid0"].AppMetadata)
	}
	if !reflect.DeepEqual(em.cache["guid1"].AppMetadata, wantAppMetadataGUID1) {
		t.Fatalf("TestMemLRUCache_WarmupExsitingCache: expect guid1 Metadata %v, got %v \n", wantAppMetadataGUID1, em.cache["guid1"].AppMetadata)
	}
	if !reflect.DeepEqual(em.cache["guid2"].AppMetadata, wantAppMetadataGUID2) {
		t.Fatalf("TestMemLRUCache_WarmupExsitingCache: expect guid2 Metadata %v, got %v \n", wantAppMetadataGUID2, em.cache["guid2"].AppMetadata)
	}
}
