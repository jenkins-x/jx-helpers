package gitlog

import "strings"

// Commit represents a git commit returned from 'git log`
type Commit struct {
	// SHA the git commit sha of the commit
	SHA string
	// Author the author name and email of the commit
	Author string
	// Date the date of the commit
	Date string
	// Comment the full commit comment
	Comment string
}

const (
	commitPrefix  = "commit "
	authorPrefix  = "Author:"
	datePrefix    = "Date:"
	commentPrefix = "    "
)

// ParseGitLog parses the output of git log into commits with SHA, author, date and comments
func ParseGitLog(text string) []*Commit {
	var answer []*Commit
	lines := strings.Split(text, "\n")
	buf := strings.Builder{}
	var c *Commit
	for _, line := range lines {
		if strings.HasPrefix(line, commitPrefix) {
			if buf.Len() > 0 && c != nil {
				c.Comment = strings.TrimSuffix(buf.String(), "\n")
			}
			buf.Reset()
			c = &Commit{
				SHA: strings.TrimPrefix(line, commitPrefix),
			}
			answer = append(answer, c)
		} else if c != nil {
			if strings.HasPrefix(line, authorPrefix) {
				c.Author = strings.TrimSpace(strings.TrimPrefix(line, authorPrefix))
			} else if strings.HasPrefix(line, datePrefix) {
				c.Date = strings.TrimSpace(strings.TrimPrefix(line, datePrefix))
			} else {
				if strings.HasPrefix(line, commentPrefix) {
					buf.WriteString(strings.TrimPrefix(line, commentPrefix))
					buf.WriteString("\n")
				}
				if strings.TrimSpace(line) == "" && buf.Len() > 0 {
					buf.WriteString("\n")
				}
			}
		}
	}
	if buf.Len() > 0 && c != nil {
		c.Comment = strings.TrimSuffix(buf.String(), "\n")
	}
	return answer
}
