package executetest

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query/execute"
)

func ProcessTestHelper(
	t *testing.T,
	data []execute.Block,
	want []*Block,
	create func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation,
) {
	t.Helper()

	d := NewDataset(RandomDatasetID())
	c := execute.NewBlockBuilderCache()
	c.SetTriggerSpec(execute.DefaultTriggerSpec)

	tx := create(d, c)

	parentID := RandomDatasetID()
	for _, b := range data {
		tx.Process(parentID, b)
	}

	got := BlocksFromCache(c)

	sort.Sort(SortedBlocks(got))
	sort.Sort(SortedBlocks(want))

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(want, got))
	}
}
