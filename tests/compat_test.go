package contract_test

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestCompatRangeSemantics 校验兼容矩阵的语义口径（合同独立 semver、xray 只读固定）。
func TestCompatRangeSemantics(t *testing.T) {
	var m map[string]string
	if err := yaml.Unmarshal(readFile(t, "compat/v1/compat.yaml"), &m); err != nil {
		t.Fatalf("解析兼容矩阵失败: %v", err)
	}
	want := map[string]string{
		"contract": ">=1.0.0 <2.0.0",
		"core":     ">=1.0.0 <2.0.0",
		"agent":    ">=1.0.0 <2.0.0",
		"schema":   "1.x",
		"xray":     "26.11.9",
	}
	for key, value := range want {
		if m[key] != value {
			t.Errorf("compat[%s] = %q, 期望 %q", key, m[key], value)
		}
	}
}
