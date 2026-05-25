package ghost

import (
	"github.com/kode/kode/internal/execution"
)

const (
	weightTDD         = 0.0  // Gate: must pass or eliminated
	weightBlastRadius = 0.40
	weightTokenCost   = 0.30
	weightTime        = 0.20
	weightGates       = 0.10
)

type ScoreWeights struct {
	BlastRadius float64
	TokenCost   float64
	Time        float64
	Gates       float64
}

func defaultWeights() ScoreWeights {
	return ScoreWeights{
		BlastRadius: weightBlastRadius,
		TokenCost:   weightTokenCost,
		Time:        weightTime,
		Gates:       weightGates,
	}
}

func ScoreBranch(result *BranchResult, allResults []*BranchResult) float64 {
	if result.Status == execution.StatusFail {
		return -1.0
	}

	weights := defaultWeights()
	score := 0.0

	// Gates passed (proportion of available gates used as a percentage)
	maxGates := 6
	gatesScore := float64(result.GatesPassed) / float64(maxGates) * weights.Gates
	score += gatesScore

	// Blast radius (lower = better, normalized inversely)
	if len(allResults) > 0 {
		maxBlast := 1
		minBlast := result.BlastRadius
		for _, r := range allResults {
			if r.BlastRadius > maxBlast {
				maxBlast = r.BlastRadius
			}
			if r.BlastRadius < minBlast {
				minBlast = r.BlastRadius
			}
		}
		rangeBlast := maxBlast - minBlast
		if rangeBlast > 0 {
			normalized := 1.0 - float64(result.BlastRadius-minBlast)/float64(rangeBlast)
			score += normalized * weights.BlastRadius
		} else {
			score += 0.5 * weights.BlastRadius
		}
	} else {
		// No comparison available — score based on absolute value
		if result.BlastRadius <= 1 {
			score += 1.0 * weights.BlastRadius
		} else if result.BlastRadius <= 3 {
			score += 0.7 * weights.BlastRadius
		} else {
			score += 0.3 * weights.BlastRadius
		}
	}

	// Token cost (lower = better)
	if len(allResults) > 0 {
		minCost := result.TokenCost
		maxCost := result.TokenCost
		for _, r := range allResults {
			if r.TokenCost > maxCost {
				maxCost = r.TokenCost
			}
			if r.TokenCost < minCost {
				minCost = r.TokenCost
			}
		}
		rangeCost := maxCost - minCost
		if rangeCost > 0 {
			normalized := 1.0 - float64(result.TokenCost-minCost)/rangeCost
			score += normalized * weights.TokenCost
		} else {
			score += 0.5 * weights.TokenCost
		}
	}

	// Execution time (faster = better)
	if len(allResults) > 0 {
		maxTime := result.Duration
		minTime := result.Duration
		for _, r := range allResults {
			if r.Duration > maxTime {
				maxTime = r.Duration
			}
			if r.Duration < minTime {
				minTime = r.Duration
			}
		}
		rangeTime := maxTime - minTime
		if rangeTime > 0 {
			normalized := 1.0 - float64(result.Duration-minTime)/float64(rangeTime)
			score += normalized * weights.Time
		} else {
			score += 0.5 * weights.Time
		}
	}

	return score
}

func SelectWinner(results []*BranchResult) *BranchResult {
	var best *BranchResult
	bestScore := -2.0

	for _, r := range results {
		if r.Status == execution.StatusFail {
			continue
		}
		r.Score = ScoreBranch(r, results)
		if r.Score > bestScore {
			bestScore = r.Score
			best = r
		}
	}

	// If all failed, pick the one with fewest errors
	if best == nil {
		for _, r := range results {
			if best == nil || r.Error < best.Error {
				best = r
			}
		}
	}

	return best
}
