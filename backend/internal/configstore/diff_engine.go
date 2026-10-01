package configstore

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"caddy-manager/internal/models"
)

// GenerateDiff computes line-by-line unified diff and statistics between current and proposed JSON
func GenerateDiff(currentJSON, proposedJSON string) models.ConfigDiffResponse {
	currentLines := splitLines(currentJSON)
	proposedLines := splitLines(proposedJSON)

	unifiedDiff, additions, deletions, modified := computeSimpleUnifiedDiff(currentLines, proposedLines)

	return models.ConfigDiffResponse{
		UnifiedDiff: unifiedDiff,
		Stats: models.DiffStats{
			Additions: additions,
			Deletions: deletions,
			Modified:  modified,
		},
	}
}

func splitLines(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func computeSimpleUnifiedDiff(oldLines, newLines []string) (string, int, int, int) {
	var buf bytes.Buffer
	additions := 0
	deletions := 0
	modified := 0

	// Simple Myers-style / line comparison for JSON config
	m := len(oldLines)
	n := len(newLines)

	// Longest Common Subsequence matrix
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if oldLines[i] == newLines[j] {
				dp[i+1][j+1] = dp[i][j] + 1
			} else {
				if dp[i+1][j] >= dp[i][j+1] {
					dp[i+1][j+1] = dp[i+1][j]
				} else {
					dp[i+1][j+1] = dp[i][j+1]
				}
			}
		}
	}

	// Backtrack to build diff
	type diffOp struct {
		kind rune // ' ', '-', '+'
		line string
	}
	var diffs []diffOp

	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			diffs = append(diffs, diffOp{kind: ' ', line: oldLines[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			diffs = append(diffs, diffOp{kind: '+', line: newLines[j-1]})
			additions++
			j--
		} else if i > 0 && (j == 0 || dp[i][j-1] < dp[i-1][j]) {
			diffs = append(diffs, diffOp{kind: '-', line: oldLines[i-1]})
			deletions++
			i--
		}
	}

	// Reverse diffs array since we backtracked from end
	for k := 0; k < len(diffs)/2; k++ {
		opp := len(diffs) - 1 - k
		diffs[k], diffs[opp] = diffs[opp], diffs[k]
	}

	buf.WriteString("--- Active Caddy Config\n")
	buf.WriteString("+++ Proposed Config\n")

	for _, d := range diffs {
		fmt.Fprintf(&buf, "%c %s\n", d.kind, d.line)
	}

	return buf.String(), additions, deletions, modified
}
