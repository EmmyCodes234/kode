package graph

type NodeID string

type NodeKind string

const (
	NodeKindFile      NodeKind = "file"
	NodeKindFunction  NodeKind = "function"
	NodeKindMethod    NodeKind = "method"
	NodeKindStruct    NodeKind = "struct"
	NodeKindInterface NodeKind = "interface"
	NodeKindImport    NodeKind = "import"
)

type EdgeKind string

const (
	EdgeKindCalls      EdgeKind = "calls"
	EdgeKindImports    EdgeKind = "imports"
	EdgeKindDefines    EdgeKind = "defines"
	EdgeKindImplements EdgeKind = "implements"
)

type Confidence string

const (
	ConfidenceCertain   Confidence = "certain"
	ConfidenceResolved  Confidence = "resolved"
	ConfidenceAmbiguous Confidence = "ambiguous"
)

type Node struct {
	ID        NodeID   `json:"id"`
	FilePath  string   `json:"file_path"`
	Kind      NodeKind `json:"kind"`
	Name      string   `json:"name"`
	Signature string   `json:"signature"`
	Content   string   `json:"content,omitempty"`
	TokenCost int      `json:"token_cost"`
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
}

type Edge struct {
	Source     NodeID     `json:"source"`
	Target     NodeID     `json:"target"`
	Kind       EdgeKind   `json:"kind"`
	Confidence Confidence `json:"confidence"`
}

type ContextGraph struct {
	Nodes       map[NodeID]*Node `json:"nodes"`
	Edges       []*Edge          `json:"edges"`
	TotalTokens int              `json:"total_tokens"`
	ProjectRoot string           `json:"project_root"`
}

func NewContextGraph() *ContextGraph {
	return &ContextGraph{
		Nodes: make(map[NodeID]*Node),
	}
}

func (g *ContextGraph) AddNode(n *Node) {
	g.Nodes[n.ID] = n
}

func (g *ContextGraph) AddEdge(e *Edge) {
	g.Edges = append(g.Edges, e)
}

func EstimatedTokenCount(text string) int {
	n := len(text) / 4
	if n < 1 {
		return 1
	}
	return n
}

type ContextPacket struct {
	Files       []ContextFile `json:"files"`
	Edges       []*Edge       `json:"edges"`
	TotalTokens int           `json:"total_tokens"`
}

type ContextFile struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Signature string `json:"signature,omitempty"`
	Content   string `json:"content,omitempty"`
}

func (g *ContextGraph) ContextPacket(maxTokens int) (*ContextPacket, error) {
	packet := &ContextPacket{
		Files: make([]ContextFile, 0),
		Edges: g.Edges,
	}
	budget := maxTokens

	for _, node := range g.Nodes {
		if budget <= 0 {
			break
		}
		cf := ContextFile{
			Path:      node.FilePath,
			Kind:      string(node.Kind),
			Signature: node.Signature,
		}
		if node.TokenCost <= budget && node.Content != "" {
			cf.Content = node.Content
			budget -= node.TokenCost
		} else if node.Signature != "" {
			budget -= EstimatedTokenCount(node.Signature)
		}
		packet.Files = append(packet.Files, cf)
	}
	packet.TotalTokens = maxTokens - budget
	return packet, nil
}
