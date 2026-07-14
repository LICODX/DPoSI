package speed

import (
	"fmt"
	"time"
	"dposi-blockchain/pkg/types"
)

// ScoreCalculator menghitung skor dinamis berdasarkan metrik jaringan
type ScoreCalculator struct {
	maxBandwidth  float64
	maxLatency    float64
	ewmaAlpha     float64 // Exponential Moving Average factor (0-1)
}

// NewScoreCalculator membuat instance ScoreCalculator baru
func NewScoreCalculator(maxBW, maxLat, alpha float64) *ScoreCalculator {
	return &ScoreCalculator{
		maxBandwidth: maxBW,
		maxLatency:   maxLat,
		ewmaAlpha:    alpha,
	}
}

// CalculateScore menghitung skor dari SpeedMetric
// Score = (0.4 * Throughput_Mbps) + (0.4 * (1000 / Latency_ms)) + (0.2 * Uptime_Percentage)
func (sc *ScoreCalculator) CalculateScore(metric *types.SpeedMetric) float64 {
	// Component 1: Throughput (40%) - average of upload and download
	avgThroughput := (metric.DownloadMbps + metric.UploadMbps) / 2.0
	bwComponent := (avgThroughput / sc.maxBandwidth) * 0.4
	if bwComponent > 0.4 {
		bwComponent = 0.4 // Cap maksimal
	}

	// Component 2: Latency (40%) - inverse relationship
	latComponent := (sc.maxLatency / metric.LatencyMs) * 0.4
	if latComponent > 0.4 {
		latComponent = 0.4 // Cap maksimal
	}

	// Component 3: Uptime (20%)
	upComponent := metric.Uptime * 0.2

	// Total score (0-100)
	totalScore := (bwComponent + latComponent + upComponent) * 100.0

	// Cap at 100
	if totalScore > 100.0 {
		totalScore = 100.0
	}

	return totalScore
}

// UpdateMovingAverage menghitung Exponential Moving Average dari skor
func (sc *ScoreCalculator) UpdateMovingAverage(currentAvg, newScore float64) float64 {
	if currentAvg == 0 {
		return newScore // First measurement
	}

	// EMA = (newScore * alpha) + (currentAvg * (1 - alpha))
	return (newScore * sc.ewmaAlpha) + (currentAvg * (1.0 - sc.ewmaAlpha))
}

// CalculateAverageMetric menghitung rata-rata dari multiple metrics
func (sc *ScoreCalculator) CalculateAverageMetric(metrics []types.SpeedMetric) types.SpeedMetric {
	if len(metrics) == 0 {
		return types.SpeedMetric{}
	}

	avg := types.SpeedMetric{
		Timestamp: time.Now().Unix(),
	}

	totalDownload := 0.0
	totalUpload := 0.0
	totalLatency := 0.0
	totalUptime := 0.0

	for _, m := range metrics {
		totalDownload += m.DownloadMbps
		totalUpload += m.UploadMbps
		totalLatency += m.LatencyMs
		totalUptime += m.Uptime
	}

	count := float64(len(metrics))
	avg.DownloadMbps = totalDownload / count
	avg.UploadMbps = totalUpload / count
	avg.LatencyMs = totalLatency / count
	avg.Uptime = totalUptime / count

	return avg
}

// ApplyGeometricHandicap menambahkan bonus untuk region dengan latensi tinggi
func (sc *ScoreCalculator) ApplyGeometricHandicap(score float64, latencyMs float64, region string) float64 {
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

	// Bonus untuk region dengan latensi tinggi
	if latencyMs > regionLatencyAvg+100 {
		return score + 5.0 // Bonus 5 poin
	}

	return score
}

// ValidateMetric memvalidasi bahwa metric memiliki nilai yang masuk akal
func (sc *ScoreCalculator) ValidateMetric(metric *types.SpeedMetric) error {
	if metric.DownloadMbps < 0 || metric.UploadMbps < 0 {
		return fmt.Errorf("negative throughput values")
	}

	if metric.LatencyMs <= 0 || metric.LatencyMs > 10000 { // Max 10 seconds
		return fmt.Errorf("invalid latency: %f ms", metric.LatencyMs)
	}

	if metric.Uptime < 0 || metric.Uptime > 1.0 {
		return fmt.Errorf("invalid uptime: %f", metric.Uptime)
	}

	if metric.Timestamp == 0 {
		return fmt.Errorf("invalid timestamp")
	}

	return nil
}

// NormalizeScore menormalisasi skor ke range 0-100
func (sc *ScoreCalculator) NormalizeScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// ScorePercentile menghitung persentil dari skor dalam daftar
func (sc *ScoreCalculator) ScorePercentile(score float64, allScores []float64) float64 {
	if len(allScores) == 0 {
		return 0.0
	}

	count := 0
	for _, s := range allScores {
		if s <= score {
			count++
		}
	}

	return (float64(count) / float64(len(allScores))) * 100.0
}
