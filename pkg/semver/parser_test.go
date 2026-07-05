package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		raw  string
		want Version
		err  bool
	}{
		// success
		{raw: "1.0.0", want: Version{Major: 1}},
		{raw: "1.2.0", want: Version{Major: 1, Minor: 2}},
		{raw: "1.2.3", want: Version{Major: 1, Minor: 2, Patch: 3}},
		{raw: "11.46.124", want: Version{Major: 11, Minor: 46, Patch: 124}},
		{raw: "1.2.3-alpha", want: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}}},
		{raw: "1.2.3-alpha.1", want: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha", "1"}}},
		{raw: "1.2.3+meta", want: Version{Major: 1, Minor: 2, Patch: 3, Build: []string{"meta"}}},
		{raw: "1.2.3+meta.1", want: Version{Major: 1, Minor: 2, Patch: 3, Build: []string{"meta", "1"}}},
		{
			raw:  "1.2.3-alpha+meta",
			want: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}, Build: []string{"meta"}},
		},
		{
			raw:  "1.2.3-alpha.1+meta.1",
			want: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha", "1"}, Build: []string{"meta", "1"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()

			v, err := Parse(tc.raw)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, v)
		})
	}
}
