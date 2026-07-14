package state

import (
	"fmt"
	"sort"
	"dposi-blockchain/pkg/types"
)

// CandidatePool mengelola kumpulan calon validator (node dengan stake)
type CandidatePool struct {
	candidates map[types.NodeID]*types.NodeCandidate
	sortedList []*types.NodeCandidate // Sorted by score descending
}

// NewCandidatePool membuat instance CandidatePool baru
func NewCandidatePool() *CandidatePool {
	return &CandidatePool{
		candidates: make(map[types.NodeID]*types.NodeCandidate),
		sortedList: make([]*types.NodeCandidate, 0),
	}
}

// AddCandidate menambahkan node sebagai calon validator
func (cp *CandidatePool) AddCandidate(candidate *types.NodeCandidate) error {
	if candidate == nil {
		return fmt.Errorf("candidate cannot be nil")
	}

	if candidate.Stake == 0 {
		return fmt.Errorf("stake must be greater than 0")
	}

	cp.candidates[candidate.NodeID] = candidate
	cp.resort() // Re-sort setelah perubahan

	return nil
}

// UpdateCandidate memperbarui candidate yang sudah ada
func (cp *CandidatePool) UpdateCandidate(nodeID types.NodeID, updates func(*types.NodeCandidate)) error {
	candidate, exists := cp.candidates[nodeID]
	if !exists {
		return fmt.Errorf("candidate not found")
	}

	updates(candidate)
	cp.resort()

	return nil
}

// RemoveCandidate menghapus candidate
func (cp *CandidatePool) RemoveCandidate(nodeID types.NodeID) error {
	if _, exists := cp.candidates[nodeID]; !exists {
		return fmt.Errorf("candidate not found")
	}

	delete(cp.candidates, nodeID)
	cp.resort()

	return nil
}

// GetCandidate mengambil candidate berdasarkan NodeID
func (cp *CandidatePool) GetCandidate(nodeID types.NodeID) (*types.NodeCandidate, error) {
	candidate, exists := cp.candidates[nodeID]
	if !exists {
		return nil, fmt.Errorf("candidate not found")
	}

	return candidate, nil
}

// GetTopCandidates mengambil top N candidates berdasarkan score
func (cp *CandidatePool) GetTopCandidates(count int) []*types.NodeCandidate {
	if count > len(cp.sortedList) {
		count = len(cp.sortedList)
	}

	result := make([]*types.NodeCandidate, count)
	copy(result, cp.sortedList[:count])

	return result
}

// GetActiveCandidates mengambil semua candidate yang status online
func (cp *CandidatePool) GetActiveCandidates() []*types.NodeCandidate {
	active := make([]*types.NodeCandidate, 0)

	for _, candidate := range cp.sortedList {
		if candidate.IsOnline {
			active = append(active, candidate)
		}
	}

	return active
}

// SelectValidators memilih validator untuk siklus baru berdasarkan skor dan stake
func (cp *CandidatePool) SelectValidators(maxCount int) []*types.NodeCandidate {
	selected := make([]*types.NodeCandidate, 0)
	activeCount := 0

	// Prefer online candidates
	for _, candidate := range cp.sortedList {
		if activeCount >= maxCount {
			break
		}

		if candidate.IsOnline && candidate.Score > 0 {
			selected = append(selected, candidate)
			activeCount++
		}
	}

	// If not enough online candidates, include offline ones
	if activeCount < maxCount {
		for _, candidate := range cp.sortedList {
			if activeCount >= maxCount {
				break
			}

			if !candidate.IsOnline && candidate.Score > 0 {
				selected = append(selected, candidate)
				activeCount++
			}
		}
	}

	return selected
}

// SelectByDelegation memilih validator berdasarkan weighted stake (own + delegated)
func (cp *CandidatePool) SelectByDelegation(maxCount int) []*types.NodeCandidate {
	selected := make([]*types.NodeCandidate, 0)

	// Create weighted candidates
	type weighted struct {
		candidate    *types.NodeCandidate
		totalStake   uint64
	}

	weighted_list := make([]weighted, 0)

	for _, candidate := range cp.sortedList {
		if candidate.IsOnline {
			totalStake := candidate.Stake + candidate.DelegatedStake
			weighted_list = append(weighted_list, weighted{
				candidate:  candidate,
				totalStake: totalStake,
			})
		}
	}

	// Sort by total stake descending
	sort.Slice(weighted_list, func(i, j int) bool {
		return weighted_list[i].totalStake > weighted_list[j].totalStake
	})

	// Select top maxCount
	count := maxCount
	if count > len(weighted_list) {
		count = len(weighted_list)
	}

	for i := 0; i < count; i++ {
		selected = append(selected, weighted_list[i].candidate)
	}

	return selected
}

// GetCandidateRank mengambil ranking candidate (1-based)
func (cp *CandidatePool) GetCandidateRank(nodeID types.NodeID) (int, error) {
	for i, candidate := range cp.sortedList {
		if candidate.NodeID == nodeID {
			return i + 1, nil // 1-based ranking
		}
	}

	return 0, fmt.Errorf("candidate not found")
}

// GetTotalStake menghitung total stake dalam pool
func (cp *CandidatePool) GetTotalStake() uint64 {
	total := uint64(0)

	for _, candidate := range cp.candidates {
		total += candidate.Stake
	}

	return total
}

// GetTotalDelegatedStake menghitung total delegated stake
func (cp *CandidatePool) GetTotalDelegatedStake() uint64 {
	total := uint64(0)

	for _, candidate := range cp.candidates {
		total += candidate.DelegatedStake
	}

	return total
}

// Size mengambil jumlah candidates dalam pool
func (cp *CandidatePool) Size() int {
	return len(cp.candidates)
}

// GetAll mengambil semua candidates
func (cp *CandidatePool) GetAll() []*types.NodeCandidate {
	result := make([]*types.NodeCandidate, 0, len(cp.candidates))

	for _, candidate := range cp.candidates {
		result = append(result, candidate)
	}

	return result
}

// resort mengurutkan ulang candidates berdasarkan score
func (cp *CandidatePool) resort() {
	cp.sortedList = make([]*types.NodeCandidate, 0)

	for _, candidate := range cp.candidates {
		cp.sortedList = append(cp.sortedList, candidate)
	}

	// Sort by score descending, then by stake descending
	sort.Slice(cp.sortedList, func(i, j int) bool {
		if cp.sortedList[i].Score != cp.sortedList[j].Score {
			return cp.sortedList[i].Score > cp.sortedList[j].Score
		}
		return cp.sortedList[i].Stake > cp.sortedList[j].Stake
	})
}

// FilterByMinScore mengambil candidates dengan score >= minScore
func (cp *CandidatePool) FilterByMinScore(minScore float64) []*types.NodeCandidate {
	result := make([]*types.NodeCandidate, 0)

	for _, candidate := range cp.sortedList {
		if candidate.Score >= minScore {
			result = append(result, candidate)
		}
	}

	return result
}

// FilterByMinStake mengambil candidates dengan stake >= minStake
func (cp *CandidatePool) FilterByMinStake(minStake uint64) []*types.NodeCandidate {
	result := make([]*types.NodeCandidate, 0)

	for _, candidate := range cp.sortedList {
		if candidate.Stake >= minStake {
			result = append(result, candidate)
		}
	}

	return result
}
