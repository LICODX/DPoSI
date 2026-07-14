package types

import "time"

// NodeID adalah identitas unik node (32 bytes hash dari pubkey)
type NodeID [32]byte

// NodeCandidate merepresentasikan kandidat validator dalam siklus
type NodeCandidate struct {
	ID            NodeID
	Stake         uint64    // Deposit/collateral yang di-lock
	Score         float64   // Hasil speed test terakhir (0-100)
	Uptime        float64   // 0.0 - 1.0 (persentase)
	Address       string    // IP:Port untuk komunikasi P2P
	PublicKey     []byte    // Ed25519 public key (32 bytes)
	Produced      int       // Jumlah blok yang diproduksi di siklus ini (max 10)
	LastScoreTime int64     // Timestamp skor terakhir diupdate
	SlashCount    int       // Berapa kali di-slash dalam 30 hari
	OfflineCount  int       // Berapa kali offline dalam siklus ini
}

// SpeedMetric menyimpan hasil pengujian kecepatan
type SpeedMetric struct {
	DownloadMbps float64   // Kecepatan download (Mbps)
	UploadMbps   float64   // Kecepatan upload (Mbps)
	LatencyMs    float64   // Latensi rata-rata (ms)
	Uptime       float64   // Uptime dalam 30 hari (0.0 - 1.0)
	Timestamp    int64     // Unix timestamp
	VerifiedBy   []NodeID  // Node yang memverifikasi hasil ini
}

// DynamicScore menghitung skor berdasarkan formula DPoSI
// Score = (0.4 * Throughput_Mbps) + (0.4 * (1000 / Latency_ms)) + (0.2 * Uptime_Percentage)
func (m *SpeedMetric) CalculateDynamicScore() float64 {
	// Normalisasi: asumsikan BW max 1000 Mbps, Latency max 500ms
	const (
		maxBandwidth  = 1000.0
		maxLatency    = 500.0
	)

	// Component 1: Throughput (average upload & download)
	avgThroughput := (m.DownloadMbps + m.UploadMbps) / 2.0
	bwComponent := (avgThroughput / maxBandwidth) * 0.4
	if bwComponent > 0.4 {
		bwComponent = 0.4 // Cap maksimal
	}

	// Component 2: Latency
	latComponent := (maxLatency / m.LatencyMs) * 0.4
	if latComponent > 0.4 {
		latComponent = 0.4 // Cap maksimal
	}

	// Component 3: Uptime
	upComponent := m.Uptime * 0.2

	// Total score (0-100)
	return (bwComponent + latComponent + upComponent) * 100.0
}

// NodeScore menyimpan riwayat skor dan delegasi
type NodeScore struct {
	NodeID           NodeID
	CurrentScore     float64
	MovingAvgScore   float64   // Exponential moving average
	SpeedMetrics     []SpeedMetric
	DelegatedPower   uint64    // Total stake yang didelegasikan ke node ini
	LastRefreshTime  int64
	GeoLocation      string    // Region: "AS", "EU", "NA", dll
	IsOnline         bool
}

// Delegation merepresentasikan delegasi stake dari token holder ke validator
type Delegation struct {
	DelegatorID   NodeID
	ValidatorID   NodeID
	Amount        uint64
	Timestamp     int64
	UndelegateAt  int64     // 0 jika belum unlock
}

// Cycle merepresentasikan satu siklus produksi blok
// 1 Cycle = 100 node x 10 blok = 1000 blok
type Cycle struct {
	CycleNumber      int64
	StartHeight      int64
	EndHeight        int64
	ActiveNodes      []NodeID
	Schedule         map[int64]NodeID  // height -> NodeID producer
	StartTime        time.Time
	EndTime          time.Time
	FinalityTime     int64             // Waktu sampai finality (ms)
	SlashingCount    int
	TotalBlocks      int64
	FinalizedBlocks  int64
}

// GeometricHandicap menambahkan bonus untuk region dengan latensi tinggi
func GeometricHandicap(latencyMs float64, region string) float64 {
	// Region dengan latensi rata-rata tinggi mendapat bonus
	var regionLatencyAvg float64
	switch region {
	case "SEA": // Southeast Asia
		regionLatencyAvg = 80.0
	case "SA": // South America
		regionLatencyAvg = 150.0
	case "AF": // Africa
		regionLatencyAvg = 200.0
	case "AS": // Asia
		regionLatencyAvg = 100.0
	case "EU": // Europe
		regionLatencyAvg = 50.0
	case "NA": // North America
		regionLatencyAvg = 60.0
	default:
		regionLatencyAvg = 100.0
	}

	if latencyMs > regionLatencyAvg+100 {
		return 5.0 // Bonus 5 poin untuk region dengan latensi tinggi
	}
	return 0.0
}
