package revision

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	// GitTimeLayout is the (default) time layout used by git.
	gitTimeLayout = "2006-01-02 15:04:05 -0700"
)

type State string

var (
	StatusAdded    State = "A"
	StatusDeleted  State = "D"
	StatusModified State = "M"
)

type DiffFile struct {
	State      State
	Path       string
	CommitID   string
	CommitDate time.Time
}

// Skip return an empty Diff because the line contained a commit-id.
func (d DiffFile) Skip() bool {
	return d.Path == ""
}

type logParser struct {
	commitID   string
	commitDate time.Time
	files      map[string]DiffFile
}

func newLogParser() logParser {
	return logParser{
		files: map[string]DiffFile{},
	}
}

func (p *logParser) parseLine(input string) (DiffFile, error) {
	var diff DiffFile

	if strings.HasPrefix(input, "\"") {
		input = strings.Trim(input, "\"")
	}

	parts := strings.Fields(input)
	if len(parts) < 2 {
		return diff, fmt.Errorf("unable to parse string: '%s'", input)
	}

	path := strings.TrimSpace(strings.Join(parts[1:], " "))

	diff.Path = path
	diff.CommitID = p.commitID
	diff.CommitDate = p.commitDate

	switch leader := parts[0]; leader {
	case "A":
		diff.State = StatusAdded
	case "M":
		diff.State = StatusModified
	case "D":
		diff.State = StatusDeleted
	default:
		if len(leader) == 40 || len(leader) == 7 {
			p.commitID = leader

			d, err := parseGitDate(path)
			if err != nil {
				return DiffFile{}, err
			}

			p.commitDate = d

			return DiffFile{}, nil
		}

		log.Printf("path: %s", leader)
		log.Printf("date: %s", path)

		return DiffFile{}, fmt.Errorf("unsupported state: '%s'", input)
	}

	return diff, nil
}

func (p *logParser) generate(ctx context.Context, input string, c chan<- DiffFile) error {
	if err := p.parse(input); err != nil {
		return err
	}

	for _, diff := range p.files {
		c <- diff
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

func (p *logParser) parse(input string) error {
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		input := scanner.Text()
		if input == "" {
			continue
		}

		diff, err := p.parseLine(scanner.Text())
		if err != nil {
			return err
		}

		if !diff.Skip() {
			p.files[diff.Path] = diff
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func parseGitDate(text string) (time.Time, error) {
	return time.Parse(gitTimeLayout, text)
}
