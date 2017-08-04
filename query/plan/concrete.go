package plan

type Plan interface {
	Operations() []Operation
	Datasets() []Dataset
}

type Dataset interface {
	Bounded() bool
	Bounds() Bounds
	setBounds(Bounds)

	Windowed() bool
	Window() Window
	setWindow(Window)

	Source() Source
	setSource(Source)
}

type Planner interface {
	Plan(AbstractPlan) (Plan, error)
}

type planner struct {
	plan *plan
}

type plan struct {
}
