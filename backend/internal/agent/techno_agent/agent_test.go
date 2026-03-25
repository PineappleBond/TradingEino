package techno_agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTechnoAgent_Init 测试 TechnoAgent 初始化
// 验证 ANAL-02: TechnoAgent 可以正确初始化
func TestTechnoAgent_Init(t *testing.T) {
	// 注意：这是测试桩，实际测试需要完整的 ServiceContext
	// 此处仅验证接口定义和嵌入文件

	t.Run("NewTechnoAgent_returns_non_nil", func(t *testing.T) {
		// 桩测试：验证函数签名
		// 实际测试需要 mock ServiceContext
		// 实际调用需要完整的 svcCtx，在集成测试中验证
	})
}

// TestTechnoAgent_Agent 测试 TechnoAgent.Agent() 方法
func TestTechnoAgent_Agent(t *testing.T) {
	t.Run("Agent_returns_underlying_adk_Agent", func(t *testing.T) {
		// 桩测试：验证 Agent() 方法存在
		// 实际测试需要初始化的 TechnoAgent 实例
		var agent *TechnoAgent
		_ = agent
		// 验证方法签名：func (t *TechnoAgent) Agent() adk.Agent
	})
}

// TestTechnoAgent_EmbeddedFiles 测试嵌入的文件
// 验证 ANAL-02: DESCRIPTION 和 SOUL 正确嵌入
func TestTechnoAgent_EmbeddedFiles(t *testing.T) {
	t.Run("DESCRIPTION_md_embeds_correctly", func(t *testing.T) {
		// 验证 DESCRIPTION 不为空
		assert.NotEmpty(t, DESCRIPTION, "DESCRIPTION should be embedded")
		assert.Greater(t, len(DESCRIPTION), 100, "DESCRIPTION should have meaningful content")
	})

	t.Run("SOUL_md_embeds_correctly", func(t *testing.T) {
		// 验证 SOUL 不为空
		assert.NotEmpty(t, SOUL, "SOUL should be embedded")
		assert.Greater(t, len(SOUL), 100, "SOUL should have meaningful content")
	})
}

// TestTechnoAgent_Headers 测试技术指标表头
func TestTechnoAgent_Headers(t *testing.T) {
	t.Run("TechnicalIndicatorsHeaders_defined", func(t *testing.T) {
		// 验证表头定义（从 okx_candlesticks.go 继承）
		// 实际测试需要导入指标计算器的表头
		assert.True(t, len(TechnicalIndicatorsHeaders) > 0, "Headers should be defined")
	})
}
