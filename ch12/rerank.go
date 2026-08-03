package main

// Rerank 在混合检索结果上做一次重排:对高度重复的内容降权,保证多样性。
func Rerank(hits []Hit, k int) []Hit {
	out := make([]Hit, 0, len(hits))
	for _, h := range hits {
		dup := false
		for _, kept := range out {
			if overlap(h.Text, kept.Text) > 0.8 {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, h)
		}
		if len(out) >= k {
			break
		}
	}
	return out
}

// overlap 计算两个文本的字面重叠比例(简单实现)。
func overlap(a, b string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	ra, rb := []rune(a), []rune(b)
	common := 0
	seen := map[rune]bool{}
	for _, r := range ra {
		seen[r] = true
	}
	for _, r := range rb {
		if seen[r] {
			common++
		}
	}
	return float64(common) / float64(len(rb))
}
