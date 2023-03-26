package runner

type Job struct {
	Instance  string
	Name      string
	Namespace string
	Template  string
	Args      []string
}
