package ifql

import (
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewAST(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "select",
			raw:  `select()`,
			want: &Function{
				Name:     "select",
				Args:     []*FunctionArg{},
				Children: []*Function{},
			},
		},
		{
			name: "select with database",
			raw:  `select(database:"telegraf")`,
			want: &Function{
				Name: "select",
				Args: []*FunctionArg{
					{
						Name: "database",
						Arg:  &StringLiteral{"telegraf"},
					},
				},
				Children: []*Function{},
			},
		},
		/*{
			name: "select with range",
			raw:  `select(database:"telegraf").range(start:-1h, end:10m)`,
		},
		{
			name: "select with range",
			raw:  `select(database:"telegraf").where(exp:{(("other"="mem") and ("this"="that")) or ("this"!="that")}).range(start:-1h, end:10m).window(period:1m).count()`,
		},*/
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
