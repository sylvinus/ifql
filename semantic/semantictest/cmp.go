package semantictest

import (
	"regexp"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/semantic"
)

var CmpOptions = []cmp.Option{
	cmp.AllowUnexported(semantic.ArrayExpression{}),
	cmp.AllowUnexported(semantic.ObjectExpression{}),
	cmp.AllowUnexported(semantic.FunctionExpression{}),
	cmpopts.IgnoreUnexported(semantic.IdentifierExpression{}),
	cmpopts.IgnoreUnexported(semantic.FunctionParam{}),
	cmp.Comparer(func(x, y *regexp.Regexp) bool { return x.String() == y.String() }),
}
