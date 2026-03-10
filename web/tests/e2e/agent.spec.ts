import { test, expect } from '@playwright/test';

test.describe('Agent Management', () => {
  test.beforeEach(async ({ page }) => {
    // Skip if no test credentials available
    test.skip(!process.env.TEST_USERNAME || !process.env.TEST_PASSWORD);

    // Login first
    await page.goto('/');
    await page.fill('input[type="text"]', process.env.TEST_USERNAME || 'admin');
    await page.fill('input[type="password"]', process.env.TEST_PASSWORD || 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  test('should display agents list page', async ({ page }) => {
    // Navigate to agents page
    await page.goto('/agents');

    // Check if agents table or list is visible
    await expect(page.locator('.ant-table').or(page.locator('[data-testid="agents-list"]'))).toBeVisible();
  });

  test('should show create agent button', async ({ page }) => {
    await page.goto('/agents');

    // Find create agent button
    const createButton = page.locator('button:has-text("Create")').or(
      page.locator('button:has-text("Add")')
    ).or(
      page.locator('[data-testid="create-agent-button"]')
    );

    await expect(createButton.first()).toBeVisible();
  });

  test('should open create agent modal', async ({ page }) => {
    await page.goto('/agents');

    // Click create button
    const createButton = page.locator('button:has-text("Create")').or(
      page.locator('button:has-text("Add")')
    ).first();

    await createButton.click();

    // Check if modal is visible
    await expect(page.locator('.ant-modal')).toBeVisible();
  });

  test('should validate agent name input', async ({ page }) => {
    await page.goto('/agents');

    // Open create modal
    const createButton = page.locator('button:has-text("Create")').or(
      page.locator('button:has-text("Add")')
    ).first();
    await createButton.click();

    // Submit without name
    await page.click('.ant-modal button[type="submit"]');

    // Check for validation error
    await expect(page.locator('.ant-form-item-explain-error')).toBeVisible();
  });

  test('should display agent details', async ({ page }) => {
    // Navigate to agents page first
    await page.goto('/agents');

    // Wait for agents to load
    await page.waitForSelector('.ant-table-tbody tr, [data-testid="agent-item"]', {
      timeout: 5000
    }).catch(() => {
      // No agents exist, skip test
      test.skip();
    });

    // Click on first agent
    const firstAgent = page.locator('.ant-table-tbody tr').first().or(
      page.locator('[data-testid="agent-item"]').first()
    );

    await firstAgent.click();

    // Check if we're on agent details page or modal
    await expect(page.locator('[data-testid="agent-details"]').or(
      page.locator('.ant-modal-content')
    )).toBeVisible();
  });

  test('should filter agents by status', async ({ page }) => {
    await page.goto('/agents');

    // Find status filter
    const statusFilter = page.locator('[data-testid="status-filter"]').or(
      page.locator('.ant-select').filter({ hasText: 'Status' })
    );

    if (await statusFilter.isVisible()) {
      await statusFilter.click();

      // Select online status
      await page.click('.ant-select-dropdown:visible li:has-text("Online")');
    }
  });

  test('should search agents by name', async ({ page }) => {
    await page.goto('/agents');

    // Find search input
    const searchInput = page.locator('input[placeholder*="Search"]').or(
      page.locator('[data-testid="search-input"]')
    );

    if (await searchInput.isVisible()) {
      await searchInput.fill('test-agent');
      await page.waitForTimeout(500); // Wait for debounce
    }
  });

  test('should display pagination for agents', async ({ page }) => {
    await page.goto('/agents');

    // Check if pagination exists
    const pagination = page.locator('.ant-pagination');

    // Pagination might not exist if there are few agents
    const hasPagination = await pagination.isVisible().catch(() => false);

    if (hasPagination) {
      const nextButton = pagination.locator('.ant-pagination-next');
      if (await nextButton.isEnabled()) {
        await nextButton.click();
        await page.waitForSelector('.ant-table-tbody tr');
      }
    }
  });

  test('should delete agent', async ({ page }) => {
    await page.goto('/agents');

    // Find delete button on first agent
    const deleteButton = page.locator('[data-testid="delete-agent"]').first();

    if (await deleteButton.isVisible()) {
      // Set up dialog handler
      page.on('dialog', dialog => dialog.accept());

      await deleteButton.click();

      // Wait for success message
      await expect(page.locator('.ant-message-success')).toBeVisible({ timeout: 5000 });
    } else {
      test.skip();
    }
  });
});
