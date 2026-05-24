package execution

import (
	"testing"
)

func TestParseLLMResponse_ValidJSON(t *testing.T) {
	p := NewHunkParser()
	input := `[{"id":"h1","file_path":"main.go","action":"MODIFY","anchor_text":"func main() {}","new_text":"func main() { println(\"hi\") }"}]`

	hunks, err := p.ParseLLMResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if hunks[0].ID != "h1" {
		t.Fatalf("expected h1, got %s", hunks[0].ID)
	}
	if hunks[0].Action != ActionModify {
		t.Fatalf("expected MODIFY, got %s", hunks[0].Action)
	}
}

func TestParseLLMResponse_MarkdownWrapper(t *testing.T) {
	p := NewHunkParser()
	input := "```json\n[{\"id\":\"h1\",\"file_path\":\"main.go\",\"action\":\"INSERT\",\"new_text\":\"func x() {}\"}]\n```"

	hunks, err := p.ParseLLMResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
}

func TestParseLLMResponse_Empty(t *testing.T) {
	p := NewHunkParser()
	_, err := p.ParseLLMResponse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseLLMResponse_EmptyArray(t *testing.T) {
	p := NewHunkParser()
	_, err := p.ParseLLMResponse("[]")
	if err == nil {
		t.Fatal("expected error for empty array")
	}
}

func TestParseLLMResponse_MissingFilePath(t *testing.T) {
	p := NewHunkParser()
	input := `[{"id":"h1","file_path":"","action":"INSERT","new_text":"x"}]`
	_, err := p.ParseLLMResponse(input)
	if err == nil {
		t.Fatal("expected error for missing file path")
	}
}

func TestParseLLMResponse_InvalidAction(t *testing.T) {
	p := NewHunkParser()
	input := `[{"id":"h1","file_path":"main.go","action":"DELETE_EVERYTHING","anchor_text":"x"}]`
	_, err := p.ParseLLMResponse(input)
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestParseLLMResponse_MODIFYMissingAnchor(t *testing.T) {
	p := NewHunkParser()
	input := `[{"id":"h1","file_path":"main.go","action":"MODIFY","new_text":"x"}]`
	_, err := p.ParseLLMResponse(input)
	if err == nil {
		t.Fatal("expected error for MODIFY without anchor text")
	}
}

func TestParseLLMResponse_INSERTMissingNewText(t *testing.T) {
	p := NewHunkParser()
	input := `[{"id":"h1","file_path":"main.go","action":"INSERT","anchor_text":"x"}]`
	_, err := p.ParseLLMResponse(input)
	if err == nil {
		t.Fatal("expected error for INSERT without new text")
	}
}

func TestParseLLMResponse_DELETEWithoutAnchor(t *testing.T) {
	p := NewHunkParser()
	input := `[{"id":"h1","file_path":"main.go","action":"DELETE","anchor_text":""}]`
	_, err := p.ParseLLMResponse(input)
	if err == nil {
		t.Fatal("expected error for DELETE without anchor text")
	}
}

func TestParseLLMResponse_MultipleHunks(t *testing.T) {
	p := NewHunkParser()
	input := `[
		{"id":"h1","file_path":"a.go","action":"MODIFY","anchor_text":"func A()","new_text":"func A() int { return 1 }"},
		{"id":"h2","file_path":"b.go","action":"INSERT","new_text":"func B() {}"}
	]`

	hunks, err := p.ParseLLMResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}
}
