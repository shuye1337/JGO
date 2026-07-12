package jdk

import (
	"jgo/internal/config"
	"testing"
)

func TestFindJDK_ExactMatch(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JDKs: map[string]config.JDKInfo{
			"Temurin-21": {
				Name:    "Temurin-21",
				Version: "21.0.1",
				Major:   21,
				Source:  "temurin",
				Path:    "/path/to/temurin-21",
			},
		},
	}

	mgr := NewManager(cfg)

	// 测试精确匹配
	jdk, key, ok := mgr.findJDK("Temurin-21")
	if !ok {
		t.Fatal("Expected to find JDK 'Temurin-21'")
	}
	if key != "Temurin-21" {
		t.Errorf("Expected key 'Temurin-21', got '%s'", key)
	}
	if jdk.Name != "Temurin-21" {
		t.Errorf("Expected JDK name 'Temurin-21', got '%s'", jdk.Name)
	}
	if jdk.Version != "21.0.1" {
		t.Errorf("Expected version '21.0.1', got '%s'", jdk.Version)
	}
}

func TestFindJDK_CaseInsensitiveMatch(t *testing.T) {
	cfg := &config.Config{
		JDKs: map[string]config.JDKInfo{
			"Temurin-21": {
				Name:    "Temurin-21",
				Version: "21.0.1",
				Major:   21,
				Source:  "temurin",
				Path:    "/path/to/temurin-21",
			},
		},
	}

	mgr := NewManager(cfg)

	// 测试大小写不敏感匹配
	testCases := []struct {
		input    string
		expected string
	}{
		{"temurin-21", "Temurin-21"},
		{"TEMURIN-21", "Temurin-21"},
		{"Temurin-21", "Temurin-21"},
		{"Temurin-21", "Temurin-21"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			jdk, key, ok := mgr.findJDK(tc.input)
			if !ok {
				t.Fatalf("Expected to find JDK with input '%s'", tc.input)
			}
			if key != tc.expected {
				t.Errorf("Expected key '%s', got '%s'", tc.expected, key)
			}
			if jdk.Name != "Temurin-21" {
				t.Errorf("Expected JDK name 'Temurin-21', got '%s'", jdk.Name)
			}
		})
	}
}

func TestFindJDK_NotFound(t *testing.T) {
	cfg := &config.Config{
		JDKs: map[string]config.JDKInfo{
			"Temurin-21": {
				Name:    "Temurin-21",
				Version: "21.0.1",
				Major:   21,
				Source:  "temurin",
				Path:    "/path/to/temurin-21",
			},
		},
	}

	mgr := NewManager(cfg)

	// 测试未找到的情况
	_, _, ok := mgr.findJDK("NonExistent-JDK")
	if ok {
		t.Fatal("Expected not to find JDK 'NonExistent-JDK'")
	}
}

func TestFindJDK_EmptyMap(t *testing.T) {
	cfg := &config.Config{
		JDKs: make(map[string]config.JDKInfo),
	}

	mgr := NewManager(cfg)

	// 测试空配置
	_, _, ok := mgr.findJDK("AnyJDK")
	if ok {
		t.Fatal("Expected not to find JDK in empty map")
	}
}

func TestFindJDK_MultipleJDKs(t *testing.T) {
	cfg := &config.Config{
		JDKs: map[string]config.JDKInfo{
			"Temurin-21": {
				Name:    "Temurin-21",
				Version: "21.0.1",
				Major:   21,
				Source:  "temurin",
				Path:    "/path/to/temurin-21",
			},
			"Corretto-17": {
				Name:    "Corretto-17",
				Version: "17.0.9",
				Major:   17,
				Source:  "corretto",
				Path:    "/path/to/corretto-17",
			},
		},
	}

	mgr := NewManager(cfg)

	// 测试在多个JDK中查找
	jdk, key, ok := mgr.findJDK("corretto-17")
	if !ok {
		t.Fatal("Expected to find JDK 'corretto-17'")
	}
	if key != "Corretto-17" {
		t.Errorf("Expected key 'Corretto-17', got '%s'", key)
	}
	if jdk.Source != "corretto" {
		t.Errorf("Expected source 'corretto', got '%s'", jdk.Source)
	}
}

func TestFindJDK_PreferExactMatch(t *testing.T) {
	// 测试精确匹配优先于大小写不敏感匹配
	cfg := &config.Config{
		JDKs: map[string]config.JDKInfo{
			"JDK": {
				Name:    "JDK",
				Version: "1.0",
				Major:   1,
				Source:  "test",
				Path:    "/path/to/jdk",
			},
			"jdk": {
				Name:    "jdk",
				Version: "2.0",
				Major:   2,
				Source:  "test",
				Path:    "/path/to/jdk2",
			},
		},
	}

	mgr := NewManager(cfg)

	// 应该找到精确匹配的"JDK"
	jdk, key, ok := mgr.findJDK("JDK")
	if !ok {
		t.Fatal("Expected to find JDK 'JDK'")
	}
	if key != "JDK" {
		t.Errorf("Expected key 'JDK', got '%s'", key)
	}
	if jdk.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", jdk.Version)
	}

	// 应该找到精确匹配的"jdk"
	jdk, key, ok = mgr.findJDK("jdk")
	if !ok {
		t.Fatal("Expected to find JDK 'jdk'")
	}
	if key != "jdk" {
		t.Errorf("Expected key 'jdk', got '%s'", key)
	}
	if jdk.Version != "2.0" {
		t.Errorf("Expected version '2.0', got '%s'", jdk.Version)
	}
}

func TestFindJDK_EmptyString(t *testing.T) {
	cfg := &config.Config{
		JDKs: map[string]config.JDKInfo{
			"Temurin-21": {
				Name:    "Temurin-21",
				Version: "21.0.1",
				Major:   21,
				Source:  "temurin",
				Path:    "/path/to/temurin-21",
			},
		},
	}

	mgr := NewManager(cfg)

	// 测试空字符串输入
	_, _, ok := mgr.findJDK("")
	if ok {
		t.Fatal("Expected not to find JDK with empty string")
	}
}
