package functions

import (
	"fmt"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
)

const GeoIpKind = "geoip"

type GeoIpOpSpec struct {
}

func init() {
	query.RegisterMethod(GeoIpKind, createGeoIpOpSpec)
	query.RegisterOpSpec(GeoIpKind, newGeoIpOp)
	plan.RegisterProcedureSpec(GeoIpKind, newGeoIpProcedure, GeoIpKind)
	execute.RegisterTransformation(GeoIpKind, createGeoIpTransformation)
}

func createGeoIpOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	spec := new(GeoIpOpSpec)

	return spec, nil
}

func newGeoIpOp() query.OperationSpec {
	return new(GeoIpOpSpec)
}

func (s *GeoIpOpSpec) Kind() query.OperationKind {
	return GeoIpKind
}

type GeoIpProcedureSpec struct {
}

func newGeoIpProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	return &GeoIpProcedureSpec{}, nil
}

func (s *GeoIpProcedureSpec) Kind() plan.ProcedureKind {
	return GeoIpKind
}
func (s *GeoIpProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(GeoIpProcedureSpec)
	return ns
}

func createGeoIpTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*GeoIpProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewGeoIpTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type geoipTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache
	db    *geoip2.Reader
}

func NewGeoIpTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *GeoIpProcedureSpec) (*geoipTransformation, error) {
	db, err := geoip2.Open("GeoIP2-City.mmdb")
	if err != nil {
		return nil, err
	}
	return &geoipTransformation{
		d:     d,
		cache: cache,
		db:    db,
	}, nil
}

func (t *geoipTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	return t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *geoipTransformation) Process(id execute.DatasetID, b execute.Block) error {
	builder, new := t.cache.BlockBuilder(b)
	if new {
		execute.AddBlockCols(b, builder)
		builder.AddCol(execute.ColMeta{
			Label:  "lat",
			Type:   execute.TFloat,
			Kind:   execute.ValueColKind,
			Common: false,
		})
		builder.AddCol(execute.ColMeta{
			Label:  "lon",
			Type:   execute.TFloat,
			Kind:   execute.ValueColKind,
			Common: false,
		})
	}

	values, err := b.Values()
	if err != nil {
		return err
	}

	cols := builder.Cols()
	valueIdx := execute.ValueIdx(cols)
	values.DoString(func(strs []string, rr execute.RowReader) {
		builder.AppendStrings(valueIdx, strs)
		for j, c := range cols {
			if c.Common {
				continue
			}
			if j == valueIdx {
				l := len(cols)
				for _, ipstr := range strs {
					// call out to GEOIP and get extra values
					// If you are using strings that may be invalid, check that ip is not nil
					ip := net.ParseIP(ipstr)
					record, err := t.db.City(ip)
					if err != nil {
						log.Println(err)
						continue
					}

					builder.AppendFloat(l-2, record.Location.Latitude)
					builder.AppendFloat(l-1, record.Location.Longitude)
				}
				continue
			}
			for i := range strs {
				switch c.Type {
				case execute.TBool:
					builder.AppendBool(j, rr.AtBool(i, j))
				case execute.TInt:
					builder.AppendInt(j, rr.AtInt(i, j))
				case execute.TUInt:
					builder.AppendUInt(j, rr.AtUInt(i, j))
				case execute.TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, j))
				case execute.TString:
					builder.AppendString(j, rr.AtString(i, j))
				case execute.TTime:
					builder.AppendTime(j, rr.AtTime(i, j))
				default:
					execute.PanicUnknownType(c.Type)
				}
			}
		}
	})

	return nil
}

func (t *geoipTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *geoipTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *geoipTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
	t.db.Close()
}
