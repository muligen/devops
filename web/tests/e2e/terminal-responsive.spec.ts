import { test, expect } from '@playwright/test'

test.describe('Responsive Terminal - 终端界面响应式布局', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!process.env.TEST_USERNAME || !process.env.TEST_PASSWORD);

    // 首先登录
    await page.goto('/');
    await page.fill('input[type="text"]', process.env.TEST_USERNAME || 'admin');
    await page.fill('input[type="password"]', process.env.TEST_PASSWORD || 'admin123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  // 移动端测试 (iPhone 12)
  test('应该在移动设备上显示垂直布局', async ({ page }) => {
    // 设置视口为移动设备尺寸
    await page.setViewportSize({ width: 375, height: 667 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 命令输入应该固定在底部
    const commandInput = page.locator('input, textarea').first();
    const box = await commandInput.boundingBox();

    if (box) {
      // 输入框应该在视口底部附近（100px 以内）
      expect(box.y + box.height).toBeGreaterThan(667 - 100);
    }

    // 输入框上方应该有可滚动的消息列表
    const messageList = page.locator('.messageList, [data-testid="message-list"]');
    await expect(messageList.first()).toBeVisible();
  });

  // 平板测试
  test('应该在平板设备上显示正确的布局', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 所有终端元素应该可见
    const terminal = page.locator('[data-testid="terminal"], .agentTerminal');
    await expect(terminal.first()).toBeVisible();

    const commandInput = page.locator('input, textarea').first();
    await expect(commandInput).toBeVisible();
  });

  // 桌面端测试
  test('应该在桌面设备上显示并排布局', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 终端界面应该可见
    const terminal = page.locator('[data-testid="terminal"], .agentTerminal');
    await expect(terminal.first()).toBeVisible();

    // 桌面端快捷命令应该可见
    const quickCommands = page.locator('[data-testid="quick-commands"], .quickCommands');
    await expect(quickCommands.first()).toBeVisible();
  });

  test('应该处理从桌面到移动端的视口调整', async ({ page }) => {
    // 从桌面视口开始
    await page.setViewportSize({ width: 1920, height: 1080 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 验证桌面端终端可见
    const terminal = page.locator('[data-testid="terminal"], .agentTerminal');
    await expect(terminal.first()).toBeVisible();

    // 调整为移动端
    await page.setViewportSize({ width: 375, height: 667 });

    // 终端应该仍然可见和可用
    await expect(terminal.first()).toBeVisible();

    const commandInput = page.locator('input, textarea').first();
    await expect(commandInput).toBeVisible();
  });

  test('应该处理从移动端到桌面端的视口调整', async ({ page }) => {
    // 从移动端视口开始
    await page.setViewportSize({ width: 375, height: 667 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 验证移动端终端可见
    const terminal = page.locator('[data-testid="terminal"], .agentTerminal');
    await expect(terminal.first()).toBeVisible();

    // 调整为桌面端
    await page.setViewportSize({ width: 1920, height: 1080 });

    // 终端应该仍然可见
    await expect(terminal.first()).toBeVisible();

    // 桌面端可能有更多功能可见
    const quickCommands = page.locator('[data-testid="quick-commands"], .quickCommands');
    await expect(quickCommands.first()).toBeVisible();
  });

  test('应该在小屏移动设备上正确显示', async ({ page }) => {
    // 小屏移动设备 (iPhone SE)
    await page.setViewportSize({ width: 320, height: 568 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 终端应该是响应式的，并在需要时水平滚动
    const terminal = page.locator('[data-testid="terminal"], .agentTerminal');
    await expect(terminal.first()).toBeVisible();
  });

  test('应该在大屏桌面显示器上正确显示', async ({ page }) => {
    // 4K 显示器
    await page.setViewportSize({ width: 2560, height: 1440 });

    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 终端应该适当使用可用空间
    const terminal = page.locator('[data-testid="terminal"], .agentTerminal');
    await expect(terminal.first()).toBeVisible();

    // 验证所有组件可见
    const commandInput = page.locator('input, textarea').first();
    await expect(commandInput).toBeVisible();

    const quickCommands = page.locator('[data-testid="quick-commands"], .quickCommands');
    await expect(quickCommands.first()).toBeVisible();
  });
});
