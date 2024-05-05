package gitlab

type ContextGroup struct {
	ID          string `expr:"id" graphql:"id"`
	Name        string `expr:"name" graphql:"name"`
	Description string `expr:"description" graphql:"description"`
}
