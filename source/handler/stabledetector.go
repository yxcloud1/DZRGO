package handler

import (
	"sync"
)

// StableValue 用接口表示任意可比较类型
type StableValue interface {
	comparable
}

// StableDetector 接口（统一方法）
type IDetector interface {
	Add(value interface{}) bool
	Reset()
}

// 单设备检测器
type stableDetector[T StableValue] struct {
	lastValue T
	count     int
	threshold int
	mu        sync.Mutex
}

func NewDetector[T StableValue](threshold int) *stableDetector[T] {
	if threshold < 1 {
		threshold = 1
	}
	return &stableDetector[T]{threshold: threshold}
}

func (sd *stableDetector[T]) Add(value interface{}) bool {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	v := value.(T)
	if v == sd.lastValue {
		sd.count++
	} else {
		sd.lastValue = v
		sd.count = 1
	}
	return sd.count >= sd.threshold
}

func (sd *stableDetector[T]) Reset() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.count = 0
	var zero T
	sd.lastValue = zero
}

// Manager 可以同时管理多种类型的设备
type Manager struct {
	detectors map[string]IDetector
	mu        sync.Mutex
}

// NewManager 创建一个管理器
func NewManager() *Manager {
	return &Manager{
		detectors: make(map[string]IDetector),
	}
}

// AddData 添加新数据，value 可以是任意支持 StableValue 的类型
func (m *Manager) AddData(deviceID string, value interface{}, threshold int) bool {
	m.mu.Lock()
	d, ok := m.detectors[deviceID]
	if !ok {
		// 根据类型动态创建检测器
		switch value.(type) {
		case int:
			d = NewDetector[int](threshold)
		case float64:
			d = NewDetector[float64](threshold)
		case string:
			d = NewDetector[string](threshold)
		default:
			m.mu.Unlock()
			panic("unsupported type")
		}
		m.detectors[deviceID] = d
	}
	m.mu.Unlock()
	return d.Add(value)
}

// Reset 重置某个设备
func (m *Manager) Reset(deviceID string) {
	m.mu.Lock()
	if d, ok := m.detectors[deviceID]; ok {
		d.Reset()
	}
	m.mu.Unlock()
}