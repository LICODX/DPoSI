package types

// TxType adalah tipe transaksi
type TxType uint8

const (
	TxTransfer TxType = iota  // Transfer token
	TxDelegate               // Delegasi stake
	TxUndelegate             // Buka delegasi
	TxSpeedTest              // Penawaran kecepatan
	TxSlashReport            // Laporan slash
	TxSpeedUpdate            // Update skor kecepatan
)

// TransactionData adalah data dinamis sesuai TxType
type TransactionData interface {
	Validate() bool
}

// TransferData untuk transfer token
type TransferData struct {
	From   NodeID
	To     NodeID
	Amount uint64
}

func (t *TransferData) Validate() bool {
	return t.Amount > 0
}

// DelegateData untuk delegasi stake
type DelegateData struct {
	Delegator NodeID
	Validator NodeID
	Amount    uint64
	LockTime  int64  // Berapa lama lock (dalam blok)
}

func (d *DelegateData) Validate() bool {
	return d.Amount > 0 && d.LockTime > 0
}

// UndelegateData untuk membuka delegasi
type UndelegateData struct {
	Delegator NodeID
	Validator NodeID
	Amount    uint64
}

func (u *UndelegateData) Validate() bool {
	return u.Amount > 0
}

// SpeedTestData untuk penawaran speed test
type SpeedTestData struct {
	TesterID   NodeID
	TargetID   NodeID
	Metric     SpeedMetric
	ProofHash  [32]byte  // Hash dari proof
}

func (s *SpeedTestData) Validate() bool {
	return len(s.ProofHash) > 0
}

// SpeedUpdateData untuk update skor kecepatan
type SpeedUpdateData struct {
	NodeID    NodeID
	NewScore  float64
	Metric    SpeedMetric
	Voters    []NodeID  // Validator yang verify
}

func (s *SpeedUpdateData) Validate() bool {
	return s.NewScore >= 0 && s.NewScore <= 100
}

// SlashReportData untuk melaporkan validator yang misbehave
type SlashReportData struct {
	ReporterID   NodeID
	OffenderID   NodeID
	Reason       string  // "offline", "double-sign", "invalid-block"
	Proof        []byte  // Evidence
	Timestamp    int64
}

func (s *SlashReportData) Validate() bool {
	return len(s.Reason) > 0 && len(s.Proof) > 0
}
