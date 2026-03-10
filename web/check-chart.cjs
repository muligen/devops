const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // Login
    await page.goto('http://localhost:3000/login');
    await page.waitForTimeout(1000);
    await page.fill('input[placeholder="用户名"]', 'admin');
    await page.fill('input[placeholder="密码"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(5000); // Wait longer for WebSocket

    // Go to agent detail
    await page.goto('http://localhost:3000/agents/b0381bda-fb22-474f-964e-137ce03b9b34');
    await page.waitForTimeout(5000); // Wait for data load

    // Take screenshot
    await page.screenshot({ path: 'F:/demo/devops/agent-detail-final.png', fullPage: true });

    // Check all cards
    const cardInfo = await page.evaluate(() => {
      const cards = document.querySelectorAll('.ant-card');
      const result = [];
      cards.forEach(card => {
        const title = card.querySelector('.ant-card-head-title');
        const svg = card.querySelector('svg');
        const canvas = card.querySelector('canvas');
        result.push({
          title: title ? title.textContent : 'No title',
          hasSvg: svg !== null,
          hasCanvas: canvas !== null,
          svgPaths: svg ? svg.querySelectorAll('path').length : 0
        });
      });
      return result;
    });

    console.log('Cards:', JSON.stringify(cardInfo, null, 2));

    // Check WebSocket status
    const wsStatus = await page.evaluate(() => {
      const notification = document.querySelector('.ant-notification');
      if (!notification) return { hasNotification: false };
      return {
        hasNotification: true,
        text: notification.textContent
      };
    });
    console.log('WebSocket status:', JSON.stringify(wsStatus, null, 2));

  } catch (e) {
    console.log('Error:', e.message);
  }

  await browser.close();
})();
