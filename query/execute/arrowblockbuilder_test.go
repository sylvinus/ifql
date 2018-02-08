package execute

import (
	"reflect"
	"testing"

	"github.com/influxdata/arrow/memory"
)

func TestNewArrowBlockBuilder(t *testing.T) {
	type args struct {
		alloc memory.Allocator
	}
	tests := []struct {
		name string
		args args
		want *ArrowBlockBuilder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewArrowBlockBuilder(tt.args.alloc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewArrowBlockBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}
