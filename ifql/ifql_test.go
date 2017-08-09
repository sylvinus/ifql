package ifql

import (
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "double-quoted string",
			raw:  `"I'm afraid you're just too darn loud"`,
			want: `I'm afraid you're just too darn loud`,
		},
		{
			name:    "test double quoted string EOL",
			raw:     "\"I'm afraid you're just too darn loud\n",
			wantErr: true,
		},
		{
			name: "double-quoted string with escape",
			raw:  `"I'm afraid you\"re just too darn loud"`,
			want: `I'm afraid you"re just too darn loud`,
		},
		{
			name: "select",
			raw:  `select()`,
			want: `telegraf`,
		},
		{
			name: "select",
			raw:  `select(database:"telegraf")`,
			want: `telegraf`,
		},
		{
			name: "select with range",
			raw:  `select(database:"telegraf").range(start:-1h, end:10m)`,
			want: `telegraf`,
		},
		{
			name: "select with range",
			raw:  `select(database:"telegraf").where(exp:{(("other"="mem") and ("this"="that")) or ("this"!="that")}).range(start:-1h, end:10m).window(period:1m).count()`,
			want: `telegraf`,
		},

		{
			name: "Positive Seconds",
			raw:  `10`,
			want: 10.0,
		},
		{
			name: "Negative Seconds",
			raw:  `-10`,
			want: -10.0,
		},
		{
			name: "Milliseconds",
			raw:  `3.141`,
			want: 3.141,
		},
		{
			name: "Microseconds",
			raw:  `3.141592`,
			want: 3.141592,
		},
		{
			name: "Nanoseconds",
			raw:  `3.141592653`,
			want: 3.141592653,
		},
		{
			name: "RFC3339 Date",
			raw:  `2006-01-02T15:04:05.999999999Z`,
			want: `2006-01-02 15:04:05.999999999 +0000 UTC`,
		},
		{
			name: "RFC3339 Date numeric zone",
			raw:  `2006-01-02T15:04:05.999999999-05:00`,
			want: `2006-01-02 15:04:05.999999999 -0500 -0500`,
		},
		{
			name: "Golang Duration",
			raw:  `1h2m3s4us`,
			want: "1h2m3.000004s",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, err := Parse("", []byte(tt.raw))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			log.Printf("%+#v", g)
			//got := string(g.([]byte))
			if !cmp.Equal(tt.want, g) {
				t.Errorf("Parse() = %s", cmp.Diff(tt.want, g))
			}
		})
	}
}
