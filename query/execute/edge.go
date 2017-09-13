package execute

type Edge interface {
	Parents() []Dataset
	Children() []Dataset
	Run()
}

type edge struct {
	parents  []Dataset
	children []Dataset
}
