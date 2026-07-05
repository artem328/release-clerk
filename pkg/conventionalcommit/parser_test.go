package conventionalcommit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		raw  string
		want Commit
	}{
		{
			name: "Commit message with description and breaking change footer",
			raw: "feat: allow provided config object to extend other configs\n\n" +
				"BREAKING CHANGE: `extends` key in config file is now used for extending other config files",
			want: Commit{
				Raw: "feat: allow provided config object to extend other configs\n\n" +
					"BREAKING CHANGE: `extends` key in config file is now used for extending other config files",
				RawHeader:   "feat: allow provided config object to extend other configs",
				RawFooter:   "BREAKING CHANGE: `extends` key in config file is now used for extending other config files",
				Type:        "feat",
				Scope:       "",
				Description: "allow provided config object to extend other configs",
				Body:        "",
				Footers: []Footer{
					{Token: "BREAKING CHANGE", Value: "`extends` key in config file is now used for extending other config files"},
				},
				IsBreaking:     true,
				IsConventional: true,
			},
		},
		{
			name: "Commit message with ! to draw attention to breaking change",
			raw:  "feat!: send an email to the customer when a product is shipped",
			want: Commit{
				Raw:            "feat!: send an email to the customer when a product is shipped",
				RawHeader:      "feat!: send an email to the customer when a product is shipped",
				RawFooter:      "",
				Type:           "feat",
				Scope:          "",
				Description:    "send an email to the customer when a product is shipped",
				Body:           "",
				Footers:        nil,
				IsBreaking:     true,
				IsConventional: true,
			},
		},
		{
			name: "Commit message with scope and ! to draw attention to breaking change",
			raw:  "feat(api)!: send an email to the customer when a product is shipped",
			want: Commit{
				Raw:            "feat(api)!: send an email to the customer when a product is shipped",
				RawHeader:      "feat(api)!: send an email to the customer when a product is shipped",
				RawFooter:      "",
				Type:           "feat",
				Scope:          "api",
				Description:    "send an email to the customer when a product is shipped",
				Body:           "",
				Footers:        nil,
				IsBreaking:     true,
				IsConventional: true,
			},
		},
		{
			name: "Commit message with both ! and BREAKING CHANGE footer",
			raw: "feat!: drop support for Node 6\n\n" +
				"BREAKING CHANGE: use JavaScript features not available in Node 6.",
			want: Commit{
				Raw: "feat!: drop support for Node 6\n\n" +
					"BREAKING CHANGE: use JavaScript features not available in Node 6.",
				RawHeader:   "feat!: drop support for Node 6",
				RawFooter:   "BREAKING CHANGE: use JavaScript features not available in Node 6.",
				Type:        "feat",
				Scope:       "",
				Description: "drop support for Node 6",
				Body:        "",
				Footers: []Footer{
					{Token: "BREAKING CHANGE", Value: "use JavaScript features not available in Node 6."},
				},
				IsBreaking:     true,
				IsConventional: true,
			},
		},
		{
			name: "Commit message with no body",
			raw:  "docs: correct spelling of CHANGELOG",
			want: Commit{
				Raw:            "docs: correct spelling of CHANGELOG",
				RawHeader:      "docs: correct spelling of CHANGELOG",
				RawFooter:      "",
				Type:           "docs",
				Scope:          "",
				Description:    "correct spelling of CHANGELOG",
				Body:           "",
				Footers:        nil,
				IsBreaking:     false,
				IsConventional: true,
			},
		},
		{
			name: "Commit message with scope",
			raw:  "feat(lang): add Polish language",
			want: Commit{
				Raw:            "feat(lang): add Polish language",
				RawHeader:      "feat(lang): add Polish language",
				RawFooter:      "",
				Type:           "feat",
				Scope:          "lang",
				Description:    "add Polish language",
				Body:           "",
				Footers:        nil,
				IsBreaking:     false,
				IsConventional: true,
			},
		},
		{
			name: "Commit message with multi-paragraph body and multiple footers",
			raw: "fix: prevent racing of requests\n" +
				"\n" +
				"Introduce a request id and a reference to latest request. Dismiss\n" +
				"incoming responses other than from latest request.\n" +
				"\n" +
				"Remove timeouts which were used to mitigate the racing issue but are\nobsolete now.\n" +
				"\n" +
				"Reviewed-by: Z\n" +
				"Refs: #123",
			want: Commit{
				Raw: "fix: prevent racing of requests\n" +
					"\n" +
					"Introduce a request id and a reference to latest request. Dismiss\n" +
					"incoming responses other than from latest request.\n" +
					"\n" +
					"Remove timeouts which were used to mitigate the racing issue but are\n" +
					"obsolete now." +
					"\n" +
					"\n" +
					"Reviewed-by: Z\n" +
					"Refs: #123",
				RawHeader: "fix: prevent racing of requests",
				RawFooter: "Reviewed-by: Z\n" +
					"Refs: #123",
				Type:        "fix",
				Scope:       "",
				Description: "prevent racing of requests",
				Body: "Introduce a request id and a reference to latest request. Dismiss\n" +
					"incoming responses other than from latest request.\n" +
					"\n" +
					"Remove timeouts which were used to mitigate the racing issue but are\n" +
					"obsolete now.",
				Footers: []Footer{
					{Token: "Reviewed-by", Value: "Z"},
					{Token: "Refs", Value: "#123"},
				},
				IsBreaking:     false,
				IsConventional: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Parse(tc.raw)

			assert.Equal(t, tc.want, got)
		})
	}

	invalidTestCases := []struct {
		name string
		raw  string
	}{
		{
			name: "Invalid header",
			raw:  "initial commit",
		},
		{
			name: "Space in type",
			raw:  "my feat: hello",
		},
	}

	for _, tc := range invalidTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Parse(tc.raw)

			assert.Equal(t, Commit{Raw: tc.raw}, got)
		})
	}
}
