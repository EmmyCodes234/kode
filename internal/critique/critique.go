package critique

type Engine struct {
	lenses []Lens
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) RegisterLens(lens Lens) {
	e.lenses = append(e.lenses, lens)
}

func (e *Engine) Critique(filePath string, content string, ctx CritiqueContext) []Finding {
	var all []Finding
	for _, lens := range e.lenses {
		findings := lens.Critique(filePath, content, ctx)
		for i := range findings {
			if findings[i].FilePath == "" {
				findings[i].FilePath = filePath
			}
		}
		all = append(all, findings...)
	}
	return all
}

func (e *Engine) CritiqueAll(files map[string]string, ctx CritiqueContext) map[string][]Finding {
	result := make(map[string][]Finding, len(files))
	for path, content := range files {
		findings := e.Critique(path, content, ctx)
		if len(findings) > 0 {
			result[path] = findings
		}
	}
	return result
}
