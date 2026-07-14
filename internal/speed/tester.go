package speed

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"
	"dposi-blockchain/pkg/types"
)

// SpeedTester melakukan pengujian kecepatan ke node target
type SpeedTester struct {
	timeout time.Duration
	dataSize int64
}

// NewSpeedTester membuat instance SpeedTester baru
func NewSpeedTester(timeoutSec int32, dataSizeKB int64) *SpeedTester {
	return &SpeedTester{
		timeout:  time.Duration(timeoutSec) * time.Second,
		dataSize: dataSizeKB * 1024, // Convert KB to bytes
	}
}

// TestNode melakukan speed test ke target node
func (st *SpeedTester) TestNode(ctx context.Context, targetAddress string) (*types.SpeedMetric, error) {
	if targetAddress == "" {
		return nil, fmt.Errorf("invalid target address")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, st.timeout)
	defer cancel()

	metric := &types.SpeedMetric{
		Timestamp: time.Now().Unix(),
	}

	// Test download speed
	downloadSpeed, err := st.testDownloadSpeed(ctx, targetAddress)
	if err != nil {
		return nil, fmt.Errorf("download test failed: %w", err)
	}
	metric.DownloadMbps = downloadSpeed

	// Test upload speed
	uploadSpeed, err := st.testUploadSpeed(ctx, targetAddress)
	if err != nil {
		return nil, fmt.Errorf("upload test failed: %w", err)
	}
	metric.UploadMbps = uploadSpeed

	// Test latency
	latency, err := st.testLatency(ctx, targetAddress)
	if err != nil {
		return nil, fmt.Errorf("latency test failed: %w", err)
	}
	metric.LatencyMs = latency

	return metric, nil
}

// testDownloadSpeed mengukur kecepatan download dari target
func (st *SpeedTester) testDownloadSpeed(ctx context.Context, targetAddress string) (float64, error) {
	conn, err := st.dialWithContext(ctx, targetAddress)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Send download request
	request := []byte("DOWNLOAD_TEST\n")
	_, err = conn.Write(request)
	if err != nil {
		return 0, err
	}

	// Measure time to receive data
	start := time.Now()
	buffer := make([]byte, 4096)
	totalBytes := int64(0)

	for totalBytes < st.dataSize {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			n, err := conn.Read(buffer)
			if err != nil {
				break
			}
			totalBytes += int64(n)
		}
	}

	duration := time.Since(start)
	if duration == 0 {
		return 0, fmt.Errorf("duration too short")
	}

	// Calculate Mbps: (bytes * 8 bits) / (duration in seconds * 1,000,000)
	mbps := (float64(totalBytes) * 8.0) / (duration.Seconds() * 1_000_000.0)
	return mbps, nil
}

// testUploadSpeed mengukur kecepatan upload ke target
func (st *SpeedTester) testUploadSpeed(ctx context.Context, targetAddress string) (float64, error) {
	conn, err := st.dialWithContext(ctx, targetAddress)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Send upload request
	request := []byte("UPLOAD_TEST\n")
	_, err = conn.Write(request)
	if err != nil {
		return 0, err
	}

	// Generate random data
	data := make([]byte, st.dataSize)
	rand.Read(data)

	// Measure time to send data
	start := time.Now()
	totalBytes := int64(0)

	for totalBytes < st.dataSize {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			chunkSize := int64(4096)
			if totalBytes+chunkSize > st.dataSize {
				chunkSize = st.dataSize - totalBytes
			}

			n, err := conn.Write(data[totalBytes : totalBytes+chunkSize])
			if err != nil {
				return 0, err
			}
			totalBytes += int64(n)
		}
	}

	duration := time.Since(start)
	if duration == 0 {
		return 0, fmt.Errorf("duration too short")
	}

	// Calculate Mbps
	mbps := (float64(totalBytes) * 8.0) / (duration.Seconds() * 1_000_000.0)
	return mbps, nil
}

// testLatency mengukur latency (round-trip time) ke target
func (st *SpeedTester) testLatency(ctx context.Context, targetAddress string) (float64, error) {
	totalLatency := int64(0)
	testCount := 10

	for i := 0; i < testCount; i++ {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			conn, err := st.dialWithContext(ctx, targetAddress)
			if err != nil {
				continue
			}

			start := time.Now()
			
			// Send ping
			_, err = conn.Write([]byte("PING\n"))
			if err != nil {
				conn.Close()
				continue
			}

			// Read pong
			buffer := make([]byte, 5)
			conn.SetReadDeadline(time.Now().Add(st.timeout))
			_, err = conn.Read(buffer)
			conn.Close()

			if err != nil {
				continue
			}

			latency := time.Since(start).Milliseconds()
			totalLatency += latency
		}
	}

	if testCount == 0 {
		return 0, fmt.Errorf("all latency tests failed")
	}

	avgLatency := float64(totalLatency) / float64(testCount)
	return avgLatency, nil
}

// dialWithContext membuat koneksi TCP dengan context timeout
func (st *SpeedTester) dialWithContext(ctx context.Context, address string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: st.timeout,
	}
	return dialer.DialContext(ctx, "tcp", address)
}

// BatchTestNodes melakukan speed test ke multiple nodes secara paralel
func (st *SpeedTester) BatchTestNodes(ctx context.Context, addresses []string) map[string]*types.SpeedMetric {
	results := make(map[string]*types.SpeedMetric)
	
	for _, addr := range addresses {
		metric, err := st.TestNode(ctx, addr)
		if err != nil {
			fmt.Printf("Speed test failed for %s: %v\n", addr, err)
			continue
		}
		results[addr] = metric
	}

	return results
}

// SelectRandomNodes memilih N node acak dari daftar node
func SelectRandomNodes(nodes []types.NodeID, count int) []types.NodeID {
	if count > len(nodes) {
		count = len(nodes)
	}

	selected := make([]types.NodeID, 0, count)
	indices := rand.Perm(len(nodes))

	for i := 0; i < count; i++ {
		selected = append(selected, nodes[indices[i]])
	}

	return selected
}
