// Package service 提供业务逻辑和领域服务
package service

// ╔══════════════════════════════════════════════════════════════╗
// ║  [CHANCE CUSTOM] 暗地加倍率 - 独立文件保护                    ║
// ║  创建时间：2026-04-16                                         ║
// ║  说明：用户设置倍率 +1 作为实际扣费倍率                          ║
// ║        例：用户设置 1，实际扣 2 倍；用户设置 2，实际扣 3 倍           ║
// ║  警告：删除此文件会导致计费倍率异常！                            ║
// ╚══════════════════════════════════════════════════════════════╝

// ApplyHiddenRateMultiplier 应用暗地加倍率逻辑
// 参数:
//   - baseCost: 基础成本
//   - userRateMultiplier: 用户设置的倍率
// 返回:
//   - 实际扣费金额（用户设置倍率 +1）
//
// 示例:
//   - 用户设置 1 → 实际扣费 = baseCost × (1+1) = baseCost × 2
//   - 用户设置 2 → 实际扣费 = baseCost × (2+1) = baseCost × 3
func ApplyHiddenRateMultiplier(baseCost float64, userRateMultiplier float64) float64 {
	return baseCost * (userRateMultiplier + 1.0)
}

// GetHiddenRateMultiplier 返回实际使用的倍率（用于内部计算）
// 参数:
//   - userRateMultiplier: 用户设置的倍率
// 返回:
//   - 实际倍率（用户设置 +1）
func GetHiddenRateMultiplier(userRateMultiplier float64) float64 {
	return userRateMultiplier + 1.0
}
