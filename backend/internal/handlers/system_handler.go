package handlers

import (
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type SystemMetrics struct {
	Timestamp   time.Time      `json:"timestamp"`
	CPU         CPUMetrics     `json:"cpu"`
	Memory      MemoryMetrics  `json:"memory"`
	Disk        DiskMetrics    `json:"disk"`
	Network     NetworkMetrics `json:"network"`
	DDoS        DDoSMetrics    `json:"ddos"`
	Uptime      int64          `json:"uptime_seconds"`
	NumCPU      int            `json:"num_cpus"`
	Goroutines  int            `json:"goroutines"`
}

type CPUMetrics struct {
	UsagePercent    float64   `json:"usage_percent"`
	PerCoreUsage    []float64 `json:"per_core_usage"`
	TotalCores      int       `json:"total_cores"`
}

type MemoryMetrics struct {
	TotalMB       float64 `json:"total_mb"`
	UsedMB        float64 `json:"used_mb"`
	FreeMB        float64 `json:"free_mb"`
	UsagePercent  float64 `json:"usage_percent"`
}

type DiskMetrics struct {
	TotalGB       float64 `json:"total_gb"`
	UsedGB        float64 `json:"used_gb"`
	FreeGB        float64 `json:"free_gb"`
	UsagePercent  float64 `json:"usage_percent"`
}

type NetworkMetrics struct {
	BytesSent     uint64  `json:"bytes_sent"`
	BytesRecv     uint64  `json:"bytes_recv"`
	UploadSpeed   float64 `json:"upload_kbps"`
	DownloadSpeed float64 `json:"download_kbps"`
}

type DDoSMetrics struct {
	Status           string  `json:"status"` // "Normal", "Under Attack", "High Traffic"
	RequestsPerSec   float64 `json:"requests_per_sec"`
	IngressRateMbps  float64 `json:"ingress_rate_mbps"`
	DetectionReason  string  `json:"detection_reason,omitempty"`
}

type SystemHandler struct {
	mu          sync.RWMutex
	lastMetrics *SystemMetrics
	lastNetStat net.IOCountersStat
	lastNetTime time.Time
}

func NewSystemHandler() *SystemHandler {
	handler := &SystemHandler{
		lastNetTime: time.Now(),
	}
	// Start background goroutine to update metrics every 5 seconds
	go handler.updateMetricsLoop()
	return handler
}

func (h *SystemHandler) updateMetricsLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.updateMetrics()
	}
}

func (h *SystemHandler) updateMetrics() {
	now := time.Now()
	metrics := &SystemMetrics{
		Timestamp:  now,
		NumCPU:     runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
		Uptime:     int64(time.Since(time.Time{}).Seconds()),
	}

	// CPU Metrics
	cpuPercent, err := cpu.Percent(time.Second, true) // per-core
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPU.PerCoreUsage = cpuPercent
		metrics.CPU.TotalCores = len(cpuPercent)
		
		var total float64
		for _, core := range cpuPercent {
			total += core
		}
		metrics.CPU.UsagePercent = math.Round(total/float64(len(cpuPercent))*100) / 100
	}

	// Memory Metrics
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.Memory.TotalMB = math.Round(float64(memInfo.Total)/1024/1024*100) / 100
		metrics.Memory.UsedMB = math.Round(float64(memInfo.Used)/1024/1024*100) / 100
		metrics.Memory.FreeMB = math.Round(float64(memInfo.Free)/1024/1024*100) / 100
		metrics.Memory.UsagePercent = math.Round(memInfo.UsedPercent*100) / 100
	}

	// Disk Metrics
	diskInfo, err := disk.Usage("/")
	if err == nil {
		metrics.Disk.TotalGB = math.Round(float64(diskInfo.Total)/1024/1024/1024*100) / 100
		metrics.Disk.UsedGB = math.Round(float64(diskInfo.Used)/1024/1024/1024*100) / 100
		metrics.Disk.FreeGB = math.Round(float64(diskInfo.Free)/1024/1024/1024*100) / 100
		metrics.Disk.UsagePercent = math.Round(diskInfo.UsedPercent*100) / 100
	}

	// Network Metrics
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		currentNet := netStats[0]
		duration := now.Sub(h.lastNetTime).Seconds()
		
		if duration > 0 {
			metrics.Network.BytesSent = currentNet.BytesSent
			metrics.Network.BytesRecv = currentNet.BytesRecv
			
			// Calculate speeds in KB/s
			sentDiff := currentNet.BytesSent - h.lastNetStat.BytesSent
			recvDiff := currentNet.BytesRecv - h.lastNetStat.BytesRecv
			
			metrics.Network.UploadSpeed = math.Round(float64(sentDiff)/1024/duration*100) / 100
			metrics.Network.DownloadSpeed = math.Round(float64(recvDiff)/1024/duration*100) / 100
			
			// DDoS Detection Logic (Simplified simulation)
			// Ingress rate in Mbps
			ingressMbps := (float64(recvDiff) * 8) / 1024 / 1024 / duration
			metrics.DDoS.IngressRateMbps = math.Round(ingressMbps*100) / 100
			
			// Simulated status based on rate
			if ingressMbps > 500 { // 500 Mbps threshold for "Attack"
				metrics.DDoS.Status = "Under Attack"
				metrics.DDoS.DetectionReason = "Extremely high ingress traffic"
			} else if ingressMbps > 100 { // 100 Mbps threshold for "High Traffic"
				metrics.DDoS.Status = "High Traffic"
				metrics.DDoS.DetectionReason = "Increased network activity detected"
			} else {
				metrics.DDoS.Status = "Normal"
			}
		}
		
		h.lastNetStat = currentNet
		h.lastNetTime = now
	}

	h.mu.Lock()
	h.lastMetrics = metrics
	h.mu.Unlock()
}

func (h *SystemHandler) GetSystemMetrics(c *gin.Context) {
	h.mu.RLock()
	metrics := h.lastMetrics
	h.mu.RUnlock()

	if metrics == nil {
		h.updateMetrics()
		h.mu.RLock()
		metrics = h.lastMetrics
		h.mu.RUnlock()
	}

	c.JSON(http.StatusOK, metrics)
}
