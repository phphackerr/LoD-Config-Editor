//go:build !windows

package paths_scanner

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) FindConfigOrExeParallel() []string {
	return []string{}
}

func (s *Scanner) CheckAndFindPaths() ([]string, error) {
	return []string{}, nil
}
