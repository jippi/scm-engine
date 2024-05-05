package gitlab

type ContextMergeRequestDiffStat struct {
	Path      string `expr:"path"`
	Additions int    `expr:"additions"`
	Deletions int    `expr:"deletions"`
}
