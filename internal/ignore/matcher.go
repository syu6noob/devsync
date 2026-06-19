package ignore

import (
	"bufio"
	"os"
	"path"
	"regexp"
	"strings"
)

type Rule struct {
	Pattern  string
	Negative bool
	DirOnly  bool
	Anchored bool
	HasSlash bool
	Regex    *regexp.Regexp
}

type Matcher struct {
	Rules []Rule
}

func DefaultRules() []string {
	return []string{
		".devsync/",
		".git/",
		"node_modules/",
		"dist/",
		"build/",
	}
}

func Load(ignoreFile string, defaults []string) (*Matcher, error) {
	m := &Matcher{}
	for _, d := range defaults {
		m.addLine(d)
	}
	f, err := os.Open(ignoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		m.addLine(s.Text())
	}
	return m, s.Err()
}

func (m *Matcher) addLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	negative := false
	if strings.HasPrefix(line, "!") {
		negative = true
		line = strings.TrimPrefix(line, "!")
	}

	line = strings.ReplaceAll(line, "\\", "/")
	anchored := strings.HasPrefix(line, "/")
	line = strings.TrimPrefix(line, "/")
	dirOnly := strings.HasSuffix(line, "/")
	line = strings.TrimSuffix(line, "/")
	line = path.Clean(line)
	if line == "." || line == "" {
		return
	}

	r := Rule{
		Pattern:  line,
		Negative: negative,
		DirOnly:  dirOnly,
		Anchored: anchored,
		HasSlash: strings.Contains(line, "/"),
	}
	r.Regex = regexp.MustCompile(globToRegex(line, anchored, r.HasSlash, dirOnly))
	m.Rules = append(m.Rules, r)
}

func (m *Matcher) Match(rel string, isDir bool) bool {
	rel = strings.ReplaceAll(rel, "\\", "/")
	rel = strings.TrimPrefix(path.Clean(rel), "./")
	if rel == "." || rel == "" {
		return false
	}
	ignored := false
	for _, r := range m.Rules {
		if r.DirOnly && !isDir && !pathHasDir(rel, r.Pattern) {
			continue
		}
		if r.Regex.MatchString(rel) {
			ignored = !r.Negative
		}
	}
	return ignored
}

func pathHasDir(rel, dir string) bool {
	if rel == dir || strings.HasPrefix(rel, dir+"/") {
		return true
	}
	parts := strings.Split(rel, "/")
	for _, p := range parts[:len(parts)-1] {
		if p == dir {
			return true
		}
	}
	return false
}

func globToRegex(pattern string, anchored, hasSlash, dirOnly bool) string {
	var b strings.Builder
	if anchored || hasSlash {
		b.WriteString("^")
	} else {
		b.WriteString("(^|.*/)")
	}

	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '*' {
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
			continue
		}
		if c == '?' {
			b.WriteString("[^/]")
			continue
		}
		b.WriteString(regexp.QuoteMeta(string(c)))
	}

	if dirOnly {
		b.WriteString("(/.*)?$")
	} else {
		b.WriteString("$")
	}
	return b.String()
}
