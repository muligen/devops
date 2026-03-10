const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const errors = [];
  page.on('console', msg => {
    if (msg.type() === 'error') {
      errors.push(msg.text());
    }
  });

  // Login first
  await page.goto('http://localhost:3000/login');
  await page.fill('input[placeholder="请输入用户名"]', 'admin');
  await page.fill('input[placeholder="请输入密码"]', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard', { timeout: 5000 });

  // Go to agent detail page
  await page.goto('http://localhost:3000/agents/b0381bda-fb22-474f-964e-137ce03b9b34');
  await page.waitForTimeout(3000);

  // Take screenshot
  await page.screenshot({ path: '/tmp/agent-detail.png', fullPage: true });

  // Check for chart element
  const chartDiv = await page.$('div[_echarts_instance_]');
  console.log('Chart found:', chartDiv !== null);

  // Check for history card
  const historyCard = await page.$('text=历史指标');
  console.log('History card found:', historyCard !== null);

  // Check metrics state
  const metricsData = await page.evaluate(() => {
    // Try to find any canvas elements
    const canvases = document.querySelectorAll('canvas');
    return {
      canvasCount: canvases.length,
      chartContainers: document.querySelectorAll('[class*="echarts"]').length
    };
  });
  console.log('Metrics data:', JSON.stringify(metricsData));

  if (errors.length > 0) {
    console.log('Console errors:', errors);
  }

  await browser.close();
})();
