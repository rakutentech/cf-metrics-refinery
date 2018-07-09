package enricher

import (
	"errors"
	"testing"
)

type en struct {
	err error
}

func (e *en) GetAppMetadata(appGUID string) (AppMetadata, error) {
	if e.err != nil {
		return AppMetadata{}, e.err
	}
	return AppMetadata{
		App:       "app" + appGUID,
		AppGUID:   appGUID,
		Space:     "space" + appGUID,
		SpaceGUID: appGUID,
		Org:       "org" + appGUID,
		OrgGUID:   appGUID,
	}, nil
}

func TestGetAppMetadataSuccess(t *testing.T) {
	ok := false
	c := NewCfCallback(&en{}, func(err error) {
		ok = true
	})
	guid := "00000000-0000-0000-0000-000000000000"
	_, err := c.GetAppMetadata(guid)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !ok {
		t.Fatal("commit callback not invoked correctly")
	}
}

func TestGetAppMetadataParentFailure(t *testing.T) {
	errParent := errors.New("Error for testing")
	c := NewCfCallback(&en{err: errParent}, func(err error) {
		if err != errParent {
			t.Fatalf("expected error %v, got %v", errParent, err)
		}
	})
	guid := "00000000-0000-0000-0000-000000000000"
	_, err := c.GetAppMetadata(guid)
	if err != errParent {
		t.Fatalf("expected error %v, got %v", errParent, err)
	}
}
