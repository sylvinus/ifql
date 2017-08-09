package ifql

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

func NewAST(ifql string) (*Function, error) {
	f, err := Parse("", []byte(ifql))
	if err != nil {
		return nil, err
	}
	return f.(*Function), nil
}
