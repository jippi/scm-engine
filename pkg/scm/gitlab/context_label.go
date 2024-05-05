package gitlab

type ContextLabel struct {
	ID          string `expr:"id" graphql:"id"`
	Title       string `expr:"title" graphql:"title"`
	Color       string `expr:"color" graphql:"color"`
	Description string `expr:"description" graphql:"description"`
}

type ContextLabelNodes struct {
	Nodes []ContextLabel `graphql:"nodes"`
}
