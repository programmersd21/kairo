package updater

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func parseChecksums(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	sums := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		sum := fields[0]
		name := fields[len(fields)-1]
		name = strings.TrimPrefix(name, "*")
		sums[name] = sum
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(sums) == 0 {
		return nil, fmt.Errorf("no checksums found in %s", path)
	}
	return sums, nil
}
