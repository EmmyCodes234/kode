package graph

import "context"

type ImportResolver interface {
	Language() string

	ResolveImport(ctx context.Context, importPath string, projectRoot string) ([]string, error)

	ResolveMethodCall(ctx context.Context, pkg string, method string, projectRoot string) (string, int, error)
}

type ResolverRegistry struct {
	resolvers map[string]ImportResolver
}

func NewResolverRegistry() *ResolverRegistry {
	return &ResolverRegistry{
		resolvers: make(map[string]ImportResolver),
	}
}

func (r *ResolverRegistry) Register(resolver ImportResolver) {
	r.resolvers[resolver.Language()] = resolver
}

func (r *ResolverRegistry) Get(language string) (ImportResolver, bool) {
	res, ok := r.resolvers[language]
	return res, ok
}
