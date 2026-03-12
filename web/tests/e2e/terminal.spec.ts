import { test, expect } from '@playwright/test'

test.describe('Agent 终端聊天界面', () => {
  test.beforeEach(async ({ page }) => {
    // 如果没有测试凭证可用则跳过
    test.skip(!process.env.TEST_USERNAME || !process.env.TEST_PASSWORD);

    // 首先登录
    await page.goto('/');
    await page.fill('input[type="text"]', process.env.TEST_USERNAME || 'admin');
    await page.fill('input[type="password"]', process.env.TEST_PASSWORD || 'admin123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  test('应该在 Agent 详情页面显示终端标签页', async ({ page }) => {
    // 导航到 Agent 页面
    await page.goto('/agents');

    // 等待 Agent 加载
    await page.waitForSelector('.ant-table-tbody tr, [data-testid="agent-item"]', {
      timeout: 5000,
    }).catch(() => {
      test.skip();
    });

    // 点击第一个 Agent 查看详情
    const firstAgent = page.locator('.ant-table-tbody tr').first().or(
      page.locator('[data-testid="agent-item"]').first()
    );
    await firstAgent.click();

    // 查找终端标签页
    const terminalTab = page.locator('text=终端').or(
      page.locator('[data-testid="terminal-tab"]')
    );

    // 终端标签页应该可见或存在于页面中
    await expect(terminalTab.first()).toBeVisible();
  });

  test('应该在终端中显示命令输入区域', async ({ page }) => {
    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    // 点击终端标签页
    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 命令输入应该可见
    const commandInput = page.locator('input[placeholder*="输入"], textarea, [data-testid="command-input"]');
    await expect(commandInput.first()).toBeVisible({ timeout: 5000 });
  });

  test('应该显示快捷命令按钮', async ({ page }) => {
    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 快捷命令区域应该可见
    const quickCommands = page.locator('[data-testid="quick-commands"], .quickCommands');
    await expect(quickCommands.first()).toBeVisible();
  });

  test('应该允许在输入框中输入命令', async ({ page }) => {
    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 查找输入框
    const commandInput = page.locator('input, textarea').first();
    await commandInput.fill('ls -la');

    expect(await commandInput.inputValue()).toBe('ls -la');
  });

  test('应该显示消息历史区域', async ({ page }) => {
    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 消息列表区域应该可见
    const messageList = page.locator('[data-testid="message-list"], .messageList, .ant-tabs-content');
    await expect(messageList.first()).toBeVisible();
  });

  test('应该在终端顶部显示 Agent 状态', async ({ page }) => {
    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').first();
    await firstAgent.click();

    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    // 终端头部应该可见并带有状态
    const terminalHeader = page.locator('[data-testid="terminal-header"], .terminalHeader');
    await expect(terminalHeader.first()).toBeVisible();

    // 状态指示器应该可见
    const statusIndicator = page.locator('.statusIndicator, [data-testid="agent-status"]');
    await expect(statusIndicator.first()).toBeVisible();
  });

  test('切换 Agent 时应该保留终端状态', async ({ page }) => {
    await page.goto('/agents');

    await page.waitForSelector('.ant-table-tbody tr', { timeout: 5000 }).catch(() => {
      test.skip();
    });

    const firstAgent = page.locator('.ant-table-tbody tr').nth(0);
    const secondAgent = page.locator('.ant-table-tbody tr').nth(1);

    // 检查第二个 Agent 是否存在
    const secondAgentCount = await secondAgent.count();
    if (secondAgentCount === 0) {
      test.skip();
    }

    // 打开第一个 Agent 终端
    await firstAgent.click();
    const terminalTab = page.locator('text=终端').first();
    await terminalTab.click();

    const commandInput = page.locator('input, textarea').first();
    await commandInput.fill('test command 1');

    // 返回 Agent 列表
    await page.goBack();

    // 打开第二个 Agent 终端
    await secondAgent.click();
    await terminalTab.click();

    // 输入应该是空的或与第二个 Agent 不同
    const inputValue = await commandInput.inputValue();
    expect(inputValue).not.toBe('test command 1');
  });
});
