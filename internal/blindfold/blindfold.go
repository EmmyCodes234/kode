package blindfold

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Obfuscator struct {
	mu          sync.Mutex
	index       int
	forward     map[string]string
	reverse     map[string]string
	salt        string
	re          *regexp.Regexp
}

type Mapping struct {
	Original string
	Obfuscated string
}

func NewObfuscator(salt string) *Obfuscator {
	return &Obfuscator{
		salt:    salt,
		forward: make(map[string]string),
		reverse: make(map[string]string),
		re:      regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]{0,63}`),
	}
}

func (o *Obfuscator) Obfuscate(content string) string {
	o.mu.Lock()
	defer o.mu.Unlock()

	known := make(map[string]bool)
	matches := o.re.FindAllString(content, -1)
	for _, m := range matches {
		if isKeyword(m) {
			continue
		}
		known[m] = true
	}

	var sorted []string
	for k := range known {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, ident := range sorted {
		if _, ok := o.forward[ident]; !ok {
			code := o.hashToCode(ident)
			o.forward[ident] = code
			o.reverse[code] = ident
		}
	}

	result := content
	for orig, code := range o.forward {
		if strings.Contains(result, orig) {
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(orig) + `\b`)
			result = re.ReplaceAllString(result, code)
		}
	}

	return result
}

func (o *Obfuscator) Deobfuscate(content string) string {
	o.mu.Lock()
	defer o.mu.Unlock()

	result := content
	for code, orig := range o.reverse {
		if strings.Contains(result, code) {
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(code) + `\b`)
			result = re.ReplaceAllString(result, orig)
		}
	}
	return result
}

func (o *Obfuscator) Mappings() []Mapping {
	o.mu.Lock()
	defer o.mu.Unlock()
	var out []Mapping
	for orig, code := range o.forward {
		out = append(out, Mapping{Original: orig, Obfuscated: code})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Original < out[j].Original
	})
	return out
}

func (o *Obfuscator) hashToCode(ident string) string {
	hash := sha256.Sum256([]byte(o.salt + ":" + ident))
	code := fmt.Sprintf("ZK%04d", int(hash[0])*256+int(hash[1])%10000)
	if o.reverse[code] != "" && o.reverse[code] != ident {
		return o.hashToCode(ident + "_")
	}
	return code
}

var keywords = map[string]bool{
	"package": true, "import": true, "func": true, "var": true, "const": true,
	"type": true, "struct": true, "interface": true, "map": true, "chan": true,
	"return": true, "if": true, "else": true, "for": true, "range": true,
	"switch": true, "case": true, "default": true, "break": true, "continue": true,
	"defer": true, "go": true, "select": true, "nil": true, "true": true, "false": true,
	"string": true, "int": true, "bool": true, "float64": true, "error": true,
	"byte": true, "rune": true, "complex64": true, "complex128": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"float32": true,
	"append": true, "len": true, "cap": true, "make": true, "new": true,
	"panic": true, "recover": true, "close": true, "delete": true,
	"print": true, "println": true,
}

func isKeyword(s string) bool {
	return keywords[s]
}
