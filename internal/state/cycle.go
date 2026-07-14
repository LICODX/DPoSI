package state

import (
	"fmt"
	"time"
	"dposi-blockchain/pkg/types"
)

// CycleManager mengelola siklus produksi blok
type CycleManager struct {
	maxNodesPerCycle int32
	blocksPerNode    int32
	blockTimeMs      int64
}

// NewCycleManager membuat instance CycleManager baru
func NewCycleManager(maxNodes, blocksPerNode int32, blockTimeMs int64) *CycleManager {
	return &CycleManager{
		maxNodesPerCycle: maxNodes,
		blocksPerNode:    blocksPerNode,
		blockTimeMs:      blockTimeMs,
	}
}

// CreateNewCycle membuat siklus baru dengan node aktif yang dipilih
func (cm *CycleManager) CreateNewCycle(cycleNumber int64, activeNodes []types.NodeID, startHeight int64) *types.Cycle {
	if int32(len(activeNodes)) > cm.maxNodesPerCycle {
		activeNodes = activeNodes[:cm.maxNodesPerCycle]
	}

	totalBlocks := int64(cm.maxNodesPerCycle) * int64(cm.blocksPerNode)
	endHeight := startHeight + totalBlocks - 1

	blockTimeSeconds := cm.blockTimeMs / 1000
	durationSeconds := totalBlocks * blockTimeSeconds

	cycle := &types.Cycle{
		CycleNumber:  cycleNumber,
		StartHeight:  startHeight,
		EndHeight:    endHeight,
		ActiveNodes:  activeNodes,
		Schedule:     make(map[int64]types.NodeID),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Duration(durationSeconds) * time.Second),
		TotalBlocks:  totalBlocks,
	}

	return cycle
}

// BuildDeterministicSchedule membuat schedule dengan Deterministic Shuffled Round-Robin
// Blok dari satu node tersebar di seluruh siklus, bukan berurutan
func (cm *CycleManager) BuildDeterministicSchedule(cycle *types.Cycle) error {
	if len(cycle.ActiveNodes) == 0 {
		return fmt.Errorf("no active nodes for cycle")
	}

	nodes := cycle.ActiveNodes
	nodeCount := int32(len(nodes))
	totalBlocks := int64(cm.maxNodesPerCycle) * int64(cm.blocksPerNode)

	// Deterministic shuffle menggunakan seed dari cycle number
	shuffledNodes := make([]types.NodeID, len(nodes))
	copy(shuffledNodes, nodes)

	// Simple deterministic shuffle based on cycle number
	seed := cycle.CycleNumber
	for i := 0; i < len(shuffledNodes); i++ {
		j := (i + int(seed)) % len(shuffledNodes)
		shuffledNodes[i], shuffledNodes[j] = shuffledNodes[j], shuffledNodes[i]
	}

	// Build round-robin schedule dengan spreading
	// Alih-alih: [N0, N0, ..., N0, N1, N1, ..., N1]
	// Gunakan: [N0, N1, N2, ..., N0, N1, N2, ...]
	blockIndex := int64(0)
	blocksPerNode := cm.blocksPerNode

	for nodeSlot := int32(0); nodeSlot < blocksPerNode; nodeSlot++ {
		for nodeIdx := int32(0); nodeIdx < nodeCount; nodeIdx++ {
			if blockIndex >= totalBlocks {
				break
			}

			height := cycle.StartHeight + blockIndex
			cycle.Schedule[height] = shuffledNodes[nodeIdx]
			blockIndex++
		}
	}

	return nil
}

// GetProducerForHeight mengambil node mana yang harus produksi blok pada height tertentu
func (cm *CycleManager) GetProducerForHeight(cycle *types.Cycle, height int64) (types.NodeID, error) {
	if height < cycle.StartHeight || height > cycle.EndHeight {
		return types.NodeID{}, fmt.Errorf("height %d out of cycle range [%d, %d]",
			height, cycle.StartHeight, cycle.EndHeight)
	}

	producer, exists := cycle.Schedule[height]
	if !exists {
		return types.NodeID{}, fmt.Errorf("no producer scheduled for height %d", height)
	}

	return producer, nil
}

// GetSlotInCycle menghitung slot (posisi) dalam siklus untuk height tertentu
func (cm *CycleManager) GetSlotInCycle(cycle *types.Cycle, height int64) (int64, error) {
	if height < cycle.StartHeight || height > cycle.EndHeight {
		return 0, fmt.Errorf("height out of cycle range")
	}

	return height - cycle.StartHeight, nil
}

// GetNodeProducedBlocks menghitung berapa blok yang sudah diproduksi node dalam siklus
func (cm *CycleManager) GetNodeProducedBlocks(cycle *types.Cycle, nodeID types.NodeID, currentHeight int64) int32 {
	count := int32(0)

	for height, producer := range cycle.Schedule {
		if height > currentHeight {
			break
		}

		if producer == nodeID {
			count++
		}
	}

	return count
}

// GetRemainingBlocksForNode menghitung berapa blok yang masih akan diproduksi node
func (cm *CycleManager) GetRemainingBlocksForNode(cycle *types.Cycle, nodeID types.NodeID, currentHeight int64) int32 {
	count := int32(0)

	for height, producer := range cycle.Schedule {
		if height <= currentHeight {
			continue
		}

		if producer == nodeID {
			count++
		}
	}

	return count
}

// EstimateCycleEnd memperkirakan waktu kapan siklus selesai
func (cm *CycleManager) EstimateCycleEnd(cycle *types.Cycle, currentHeight int64) time.Time {
	remainingBlocks := cycle.EndHeight - currentHeight
	if remainingBlocks <= 0 {
		return cycle.EndTime
	}

	blockTimeSeconds := cm.blockTimeMs / 1000
	remainingSeconds := remainingBlocks * blockTimeSeconds

	return time.Now().Add(time.Duration(remainingSeconds) * time.Second)
}

// ValidateSchedule memvalidasi bahwa schedule konsisten dan complete
func (cm *CycleManager) ValidateSchedule(cycle *types.Cycle) error {
	if len(cycle.Schedule) == 0 {
		return fmt.Errorf("schedule is empty")
	}

	totalBlocks := cycle.EndHeight - cycle.StartHeight + 1
	if int64(len(cycle.Schedule)) != totalBlocks {
		return fmt.Errorf("schedule incomplete: expected %d blocks, got %d",
			totalBlocks, len(cycle.Schedule))
	}

	// Verify setiap height ada mapping
	for height := cycle.StartHeight; height <= cycle.EndHeight; height++ {
		if _, exists := cycle.Schedule[height]; !exists {
			return fmt.Errorf("missing schedule entry for height %d", height)
		}
	}

	// Verify setiap node memproduksi tepat blocksPerNode blok
	nodeCounts := make(map[types.NodeID]int32)
	for _, producer := range cycle.Schedule {
		nodeCounts[producer]++
	}

	for nodeID, count := range nodeCounts {
		if count != cm.blocksPerNode {
			return fmt.Errorf("node %x produced %d blocks, expected %d",
				nodeID, count, cm.blocksPerNode)
		}
	}

	return nil
}
