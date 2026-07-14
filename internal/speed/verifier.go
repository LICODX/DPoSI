package speed

import (
	"encoding/json"
	"fmt"
	"dposi-blockchain/pkg/crypto"
	"dposi-blockchain/pkg/types"
)

// Verifier memverifikasi hasil speed test dari multiple validators
type Verifier struct {
	requiredVerifiers int
	verifyThreshold   float64 // Persentase verifier yang harus setuju (0.0-1.0)
}

// NewVerifier membuat instance Verifier baru
func NewVerifier(requiredCount int, threshold float64) *Verifier {
	return &Verifier{
		requiredVerifiers: requiredCount,
		verifyThreshold:   threshold,
	}
}

// VerificationResult merepresentasikan hasil verifikasi dari satu validator
type VerificationResult struct {
	ValidatorID   types.NodeID
	MetricHash    [32]byte // Hash dari metric yang diverifikasi
	IsValid       bool     // Apakah validator percaya metric valid
	Confidence    float64  // Confidence level (0.0-1.0)
	Signature     []byte   // Signature dari validator
	Timestamp     int64
}

// VerifyMetric memverifikasi apakah metric valid secara statistik
func (v *Verifier) VerifyMetric(metric *types.SpeedMetric, referenceMetrics []types.SpeedMetric) (bool, float64, error) {
	if len(referenceMetrics) == 0 {
		// Tidak ada reference, accept metric
		return true, 1.0, nil
	}

	// Calculate mean dan standard deviation
	meanDownload, stdDownload := calculateStats(referenceMetrics, func(m *types.SpeedMetric) float64 {
		return m.DownloadMbps
	})
	meanUpload, stdUpload := calculateStats(referenceMetrics, func(m *types.SpeedMetric) float64 {
		return m.UploadMbps
	})
	meanLatency, stdLatency := calculateStats(referenceMetrics, func(m *types.SpeedMetric) float64 {
		return m.LatencyMs
	})

	// Check outliers (more than 3 sigma away)
	downloadAnomaly := isOutlier(metric.DownloadMbps, meanDownload, stdDownload)
	uploadAnomaly := isOutlier(metric.UploadMbps, meanUpload, stdUpload)
	latencyAnomaly := isOutlier(metric.LatencyMs, meanLatency, stdLatency)

	anomalyCount := 0
	if downloadAnomaly {
		anomalyCount++
	}
	if uploadAnomaly {
		anomalyCount++
	}
	if latencyAnomaly {
		anomalyCount++
	}

	// If more than 1 anomaly, mark as invalid
	if anomalyCount > 1 {
		confidence := 0.2
		return false, confidence, nil
	}

	// Calculate confidence score
	confidence := calculateConfidence(
		metric.DownloadMbps, meanDownload, stdDownload,
		metric.UploadMbps, meanUpload, stdUpload,
		metric.LatencyMs, meanLatency, stdLatency,
	)

	isValid := confidence > 0.5
	return isValid, confidence, nil
}

// CreateVerificationProof membuat proof yang ditandatangani validator
func (v *Verifier) CreateVerificationProof(
	metric *types.SpeedMetric,
	validatorKeyPair *types.KeyPair,
	validatorID types.NodeID,
	isValid bool,
	confidence float64,
) (*VerificationResult, error) {
	// Hash metric untuk proof
	metricBytes, err := json.Marshal(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric: %w", err)
	}

	metricHash := crypto.SHA256(metricBytes)

	// Create verification data
	verifyData := fmt.Sprintf("%x:%v:%.2f:%d",
		metricHash,
		isValid,
		confidence,
		metric.Timestamp,
	)

	// Sign verification (menggunakan interface yang sesuai)
	// Note: Kita menggunakan metricHash[:] sebagai data untuk sign
	signature, err := signData(validatorKeyPair, metricHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign verification: %w", err)
	}

	result := &VerificationResult{
		ValidatorID: validatorID,
		MetricHash:  metricHash,
		IsValid:     isValid,
		Confidence:  confidence,
		Signature:   signature,
		Timestamp:   metric.Timestamp,
	}

	return result, nil
}

// AggregateVerifications menggabungkan hasil verifikasi dari multiple validators
func (v *Verifier) AggregateVerifications(verifications []VerificationResult) (bool, float64) {
	if len(verifications) == 0 {
		return false, 0.0
	}

	// Check if we have enough verifications
	if len(verifications) < v.requiredVerifiers {
		return false, 0.0
	}

	// Calculate weighted average of confidence
	totalConfidence := 0.0
	validCount := 0

	for _, ver := range verifications {
		if ver.IsValid {
			totalConfidence += ver.Confidence
			validCount++
		}
	}

	if validCount == 0 {
		return false, 0.0
	}

	avgConfidence := totalConfidence / float64(validCount)
	agreePercentage := float64(validCount) / float64(len(verifications))

	// Check if agreement is above threshold
	isConsensus := agreePercentage >= v.verifyThreshold

	return isConsensus, avgConfidence
}

// Helper functions

// calculateStats menghitung mean dan standard deviation
func calculateStats(metrics []types.SpeedMetric, getValue func(*types.SpeedMetric) float64) (float64, float64) {
	if len(metrics) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for i := range metrics {
		sum += getValue(&metrics[i])
	}
	mean := sum / float64(len(metrics))

	// Calculate standard deviation
	variance := 0.0
	for i := range metrics {
		diff := getValue(&metrics[i]) - mean
		variance += diff * diff
	}
	variance /= float64(len(metrics))
	stddev := sqrt(variance)

	return mean, stddev
}

// isOutlier mengecek apakah value adalah outlier (> 3 sigma)
func isOutlier(value, mean, stddev float64) bool {
	if stddev == 0 {
		return false
	}

	zscore := (value - mean) / stddev
	return zscore > 3.0 || zscore < -3.0
}

// calculateConfidence menghitung confidence score
func calculateConfidence(
	downloadVal, downloadMean, downloadStd,
	uploadVal, uploadMean, uploadStd,
	latencyVal, latencyMean, latencyStd float64,
) float64 {
	confidence := 1.0

	// Penalty untuk outliers (tapi tidak ekstrem)
	if downloadStd > 0 {
		zscore := (downloadVal - downloadMean) / downloadStd
		if zscore > 2.0 || zscore < -2.0 {
			confidence -= 0.2
		}
	}

	if uploadStd > 0 {
		zscore := (uploadVal - uploadMean) / uploadStd
		if zscore > 2.0 || zscore < -2.0 {
			confidence -= 0.2
		}
	}

	if latencyStd > 0 {
		zscore := (latencyVal - latencyMean) / latencyStd
		if zscore > 2.0 || zscore < -2.0 {
			confidence -= 0.2
		}
	}

	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// sqrt menghitung square root
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// signData adalah wrapper untuk signing data dengan KeyPair
func signData(kp *types.KeyPair, data []byte) ([]byte, error) {
	if kp == nil {
		return nil, fmt.Errorf("invalid keypair")
	}
	// Note: Anda perlu mengimplementasikan interface KeyPair signing
	// Untuk sekarang, return dummy signature
	return make([]byte, 64), nil
}

// KeyPair adalah interface untuk signing operations
type KeyPair interface {
	Sign(message []byte) ([]byte, error)
	Verify(message []byte, signature []byte) bool
}
