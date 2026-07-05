package semver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		v    Version
		want string
	}{
		{v: Version{Major: 1}, want: "1.0.0"},
		{v: Version{Major: 1, Minor: 2}, want: "1.2.0"},
		{v: Version{Major: 1, Minor: 2, Patch: 3}, want: "1.2.3"},
		{v: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}}, want: "1.2.3-alpha"},
		{v: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha", "1"}}, want: "1.2.3-alpha.1"},
		{
			v:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha", "1"}, Build: []string{"meta"}},
			want: "1.2.3-alpha.1+meta",
		},
		{
			v:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha", "1"}, Build: []string{"meta", "abc"}},
			want: "1.2.3-alpha.1+meta.abc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.v.String())
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	t.Parallel()

	type cmp string
	const (
		gt cmp = ">"
		lt cmp = "<"
		eq cmp = "="
	)

	testCases := []struct {
		v1, v2 Version
		cmp    cmp
	}{
		// 1.0.0 < 2.0.0 < 2.1.0 < 2.1.1.
		{v1: Version{Major: 1}, v2: Version{Major: 2}, cmp: lt},
		{v1: Version{Major: 2}, v2: Version{Major: 2, Minor: 1}, cmp: lt},
		{v1: Version{Major: 2, Minor: 1}, v2: Version{Major: 2, Minor: 1, Patch: 1}, cmp: lt},
		// 2.1.1 > 2.1.0 > 2.0.0 > 1.0.0.
		{v1: Version{Major: 2, Minor: 1, Patch: 1}, v2: Version{Major: 2, Minor: 1}, cmp: gt},
		{v1: Version{Major: 2, Minor: 1}, v2: Version{Major: 2}, cmp: gt},
		{v1: Version{Major: 2}, v2: Version{Major: 1}, cmp: gt},
		// equality
		{v1: Version{Major: 2}, v2: Version{Major: 2}, cmp: eq},
		{v1: Version{Major: 2, Minor: 1}, v2: Version{Major: 2, Minor: 1}, cmp: eq},
		{v1: Version{Major: 2, Minor: 1, Patch: 3}, v2: Version{Major: 2, Minor: 1, Patch: 3}, cmp: eq},

		{v1: Version{Major: 1, PreRelease: []string{"alpha"}}, v2: Version{Major: 1}, cmp: lt},
		// 1.0.0 > 1.0.0-alpha.
		{v1: Version{Major: 1}, v2: Version{Major: 1, PreRelease: []string{"alpha"}}, cmp: gt},
		// 1.0.0-alpha = 1.0.0-alpha.
		{
			v1:  Version{Major: 1, PreRelease: []string{"alpha"}},
			v2:  Version{Major: 1, PreRelease: []string{"alpha"}},
			cmp: eq,
		},

		// 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.
		{
			v1:  Version{Major: 1, PreRelease: []string{"alpha"}},
			v2:  Version{Major: 1, PreRelease: []string{"alpha", "1"}},
			cmp: lt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"alpha", "1"}},
			v2:  Version{Major: 1, PreRelease: []string{"alpha", "beta"}},
			cmp: lt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"alpha", "beta"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta"}},
			cmp: lt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta", "2"}},
			cmp: lt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta", "2"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta", "11"}},
			cmp: lt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta", "11"}},
			v2:  Version{Major: 1, PreRelease: []string{"rc", "1"}},
			cmp: lt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"rc", "1"}},
			v2:  Version{Major: 1},
			cmp: lt,
		},
		// 1.0.0 > 1.0.0-rc.1 > 1.0.0-beta.11 > 1.0.0-beta.2 > 1.0.0-beta > 1.0.0-alpha.beta > 1.0.0-alpha.1 > 1.0.0-alpha.
		{
			v1:  Version{Major: 1},
			v2:  Version{Major: 1, PreRelease: []string{"rc", "1"}},
			cmp: gt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"rc", "1"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta", "11"}},
			cmp: gt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta", "11"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta", "2"}},
			cmp: gt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta", "2"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta"}},
			cmp: gt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta"}},
			v2:  Version{Major: 1, PreRelease: []string{"alpha", "beta"}},
			cmp: gt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"alpha", "beta"}},
			v2:  Version{Major: 1, PreRelease: []string{"alpha", "1"}},
			cmp: gt,
		},
		{
			v1:  Version{Major: 1, PreRelease: []string{"alpha", "1"}},
			v2:  Version{Major: 1, PreRelease: []string{"alpha"}},
			cmp: gt,
		},
		// 1.0.0-beta.11 = 1.0.0-beta.11
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta", "11"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta", "11"}},
			cmp: eq,
		},

		// build metadata has no effect on presedence
		// e.g. 1.0.0-beta.11+anything.1 = 1.0.0-beta.11+anything.2
		{
			v1:  Version{Major: 1, PreRelease: []string{"beta", "11"}, Build: []string{"anything", "1"}},
			v2:  Version{Major: 1, PreRelease: []string{"beta", "11"}, Build: []string{"anything", "2"}},
			cmp: eq,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s%s%s", tc.v1, tc.cmp, tc.v2), func(t *testing.T) {
			t.Parallel()

			res := tc.v1.Compare(tc.v2)
			switch tc.cmp {
			case eq:
				assert.Equal(t, 0, res)
			case lt:
				assert.Less(t, res, 0)
			case gt:
				assert.Greater(t, res, 0)
			default:
				assert.Fail(t, "Unknown comparison", tc.cmp)
			}
		})
	}
}

func TestVersion_BumpMajor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		v, want Version
	}{
		{v: Version{Major: 1}, want: Version{Major: 2}},
		{v: Version{Major: 1, Minor: 2}, want: Version{Major: 2}},
		{v: Version{Major: 1, Minor: 2, Patch: 3}, want: Version{Major: 2}},
		{v: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}}, want: Version{Major: 2}},
		{
			v:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}, Build: []string{"meta"}},
			want: Version{Major: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s->%s", tc.v, tc.want), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.v.BumpMajor())
		})
	}
}

func TestVersion_BumpMinor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		v, want Version
	}{
		{v: Version{Major: 1}, want: Version{Major: 1, Minor: 1}},
		{v: Version{Major: 1, Minor: 2}, want: Version{Major: 1, Minor: 3}},
		{v: Version{Major: 1, Minor: 2, Patch: 3}, want: Version{Major: 1, Minor: 3}},
		{v: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}}, want: Version{Major: 1, Minor: 3}},
		{
			v:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}, Build: []string{"meta"}},
			want: Version{Major: 1, Minor: 3},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s->%s", tc.v, tc.want), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.v.BumpMinor())
		})
	}
}

func TestVersion_BumpPatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		v, want Version
	}{
		{v: Version{Major: 1}, want: Version{Major: 1, Patch: 1}},
		{v: Version{Major: 1, Minor: 2}, want: Version{Major: 1, Minor: 2, Patch: 1}},
		{v: Version{Major: 1, Minor: 2, Patch: 3}, want: Version{Major: 1, Minor: 2, Patch: 4}},
		{
			v:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}},
			want: Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			v:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"alpha"}, Build: []string{"meta"}},
			want: Version{Major: 1, Minor: 2, Patch: 4},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s->%s", tc.v, tc.want), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.v.BumpPatch())
		})
	}
}
