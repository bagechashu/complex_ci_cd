/**
 * theme.ts - 全局主题配置 (CRT 复古绿屏风格)
 */

// CRT 复古绿屏调色盘 - 现代白+绿 (Minimalist Retro)
export const CRT_COLORS = {
  // 主色系 - 森林绿
  primary: '#2d8659',        // 森林绿
  primaryHover: '#43a047',   // 亮绿
  primaryActive: '#1e6e3f',  // 深绿
  
  // 背景色系 - 现代浅灰
  bgDark: '#f5f5f5',         // 浅灰 - 侧边栏
  bgMedium: '#f0f0f0',       // 更浅灰 - 顶部栏
  bgLight: '#e5e5e5',        // 淡灰 - 内容区
  bgCard: '#ffffff',         // 白色 - 卡片背景
  
  // 文字色系 - 深色文字（适配浅色背景）
  textPrimary: '#1a1a1a',    // 深黑 - 主文字
  textSecondary: '#4a4a4a',  // 深灰 - 辅助文字
  textLight: '#727272',      // 中灰 - 嵌入式文字
  
  // 状态色
  success: '#2d8659',
  error: '#e63946',
  warning: '#f77f00',
  info: '#2d8659'
}

// Naive UI 主题覆盖配置
export const themeOverrides: Record<string, any> = {
  common: {
    primaryColor: CRT_COLORS.primary,
    successColor: CRT_COLORS.success,
    errorColor: CRT_COLORS.error,
    warningColor: CRT_COLORS.warning,
    infoColor: CRT_COLORS.info,
    
    // 文字颜色
    textColorBase: CRT_COLORS.textPrimary,
  },
  
  Button: {
    // 主按钮
    colorPrimary: CRT_COLORS.primary,
    colorPrimaryHover: CRT_COLORS.primaryHover,
    colorPrimaryPressed: CRT_COLORS.primaryActive,
    
    // 成功按钮
    colorSuccessHover: CRT_COLORS.primaryHover,
    colorSuccessPressed: CRT_COLORS.primaryActive,
    
    // 默认按钮
    border: '1px solid #d0d0d0',
    borderHover: `1px solid ${CRT_COLORS.primary}`,
    textColorGhost: CRT_COLORS.textPrimary,
    textColorGhostHover: CRT_COLORS.primary,
  },
  
  Card: {
    borderColor: '#e0e0e0',
    // 白色卡片
    color: CRT_COLORS.bgCard,
    colorTarget: CRT_COLORS.bgCard,
  },
  
  Input: {
    colorTarget: CRT_COLORS.bgCard,
    border: '1px solid #d0d0d0',
    borderHover: `1px solid ${CRT_COLORS.primary}`,
    borderFocus: `1px solid ${CRT_COLORS.primary}`,
    textColor: CRT_COLORS.textPrimary,
  },
  
  Select: {
    colorTarget: CRT_COLORS.bgCard,
    border: '1px solid #d0d0d0',
    borderHover: `1px solid ${CRT_COLORS.primary}`,
    borderFocus: `1px solid ${CRT_COLORS.primary}`,
    textColor: CRT_COLORS.textPrimary,
  },
  
  Form: {
    labelTextColor: CRT_COLORS.textPrimary,
    asteriskColor: CRT_COLORS.error,
  },
  
  Steps: {
    stepsNumHeaderBg: CRT_COLORS.primary,
    stepsHeaderProcessLineColor: CRT_COLORS.primary,
  },
  
  Tag: {
    colorCheckable: CRT_COLORS.primary,
    colorCheckableHover: CRT_COLORS.primaryHover,
  },
  
  Divider: {
    color: '#e0e0e0',
  },
  
  Progress: {
    railColor: '#f0f0f0',
    fillColor: CRT_COLORS.primary,
  }
}

export default themeOverrides
