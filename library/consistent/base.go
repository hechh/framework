package consistent

import "hash/fnv"

// mix64 使用 splitmix64 终结器对 64 位值做雪崩混合（双射，保持均匀性）。
// FNV-64a 雪崩性差：对只差末尾字节的输入（如顺序 ID、"id:index"）输出近乎相邻，
// 会导致每个节点的虚拟点成簇、顺序 key 挤在窄带。mix64 可将其打散到全 64 位空间。
func mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

// hashKey 计算字符串key的哈希值（FNV-64a + mix64 消除相关性）
func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return mix64(h.Sum64())
}

// hashUint64 计算uint64 key的哈希值（直接 mix64，避免 FNV 对顺序 key 成簇）
func hashUint64(key uint64) uint64 {
	return mix64(key)
}

// hashVirtualKey 计算以播种 key 生成的第 index 个虚拟点的哈希值：
// FNV-64a(key + ":" + index) + mix64。同一节点相邻 index 经 mix64 雪崩打散，
// 使该节点的虚拟点在环上均匀分布。
func CalcHashVirtualKey(key string, index int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	// 追加 ':' 与 index 的十进制（int 最长 20 位，含负号），避免 fmt.Sprintf 分配
	var buf [1 + 21]byte
	n := 0
	buf[n] = ':'
	n++
	if index == 0 {
		buf[n] = '0'
		n++
	} else {
		temp := [20]byte{}
		i := 20
		t := index
		neg := false
		if t < 0 {
			neg = true
			t = -t
		}
		for t > 0 {
			i--
			temp[i] = byte('0' + t%10)
			t /= 10
		}
		if neg {
			i--
			temp[i] = '-'
		}
		copy(buf[n:], temp[i:])
		n += 20 - i
	}
	h.Write(buf[:n])
	return mix64(h.Sum64())
}

// hashVirtualNode 计算虚拟节点的哈希值
func CalcHashVirtualNode(nodeID uint32, index int) uint64 {
	h := fnv.New64a()
	// 手动格式化字符串避免 fmt.Sprintf 的内存分配
	// 格式: "nodeID:index"，nodeID 最大 10位，index 最大 10位，加冒号 1位，共 21 字节
	var buf [21]byte
	n := 0

	// 写入 nodeID
	if nodeID == 0 {
		buf[n] = '0'
		n++
	} else {
		temp := [10]byte{}
		i := 10
		t := nodeID
		for t > 0 {
			i--
			temp[i] = byte('0' + t%10)
			t /= 10
		}
		copy(buf[n:], temp[i:])
		n += 10 - i
	}

	buf[n] = ':'
	n++

	// 写入 index
	if index == 0 {
		buf[n] = '0'
		n++
	} else {
		temp := [10]byte{}
		i := 10
		t := index
		neg := false
		if t < 0 {
			neg = true
			t = -t
		}
		for t > 0 {
			i--
			temp[i] = byte('0' + t%10)
			t /= 10
		}
		if neg {
			i--
			temp[i] = '-'
		}
		copy(buf[n:], temp[i:])
		n += 10 - i
	}

	h.Write(buf[:n])
	return mix64(h.Sum64())
}
