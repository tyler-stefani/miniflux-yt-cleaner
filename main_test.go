package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	mf "miniflux.app/v2/client"
)

type mockMfClient struct {
	RemovedIDs []int64
}

func (m *mockMfClient) Entries(filter *mf.Filter) (*mf.EntryResultSet, error) {
	return &mf.EntryResultSet{
		Entries: mf.Entries{
			&mf.Entry{
				ID:          1,
				URL:         "youtube.com?v=SHORT",
				ReadingTime: 0,
			},
			&mf.Entry{
				ID:          2,
				URL:         "youtube.com?v=LIVE",
				ReadingTime: 0,
			},
			&mf.Entry{
				ID:          3,
				URL:         "youtube.com?v=STANDARD",
				ReadingTime: 10,
			},
		},
	}, nil
}

func (m *mockMfClient) UpdateEntries(ids []int64, updatedStatus string) error {
	m.RemovedIDs = ids
	return nil
}

type mockHttpClient struct{}

func (m *mockHttpClient) Head(url string) (*http.Response, error) {
	if strings.Contains(url, "SHORT") {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	} else {
		return &http.Response{
			StatusCode: 300,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		RemoveShorts bool
		RemoveLives  bool

		ExpectedIDsRemoved []int64
	}{
		{
			false,
			false,
			[]int64{},
		},
		{
			false,
			true,
			[]int64{2},
		},
		{
			true,
			false,
			[]int64{1},
		},
		{
			true,
			true,
			[]int64{1, 2},
		},
	}

	for _, tt := range tests {
		mfc := &mockMfClient{}
		clean(mfc, tt.RemoveShorts, tt.RemoveLives, &mockHttpClient{})
		assert.ElementsMatch(t, mfc.RemovedIDs, tt.ExpectedIDsRemoved)
	}
}
