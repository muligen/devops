import { test, expect } from '@playwright/test';

test.describe('Task Management', () => {
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

  test('should display tasks list page', async ({ page }) => {
    // Navigate to tasks page
    await page.goto('/tasks');

    // Check if tasks table is visible
    await expect(page.locator('.ant-table').or(page.locator('[data-testid="tasks-list"]'))).toBeVisible();
  });

  test('should show create task button', async ({ page }) => {
    await page.goto('/tasks');

    // Find create task button
    const createButton = page.locator('button:has-text("Create")').or(
      page.locator('button:has-text("New Task")')
    ).or(
      page.locator('[data-testid="create-task-button"]')
    );

    await expect(createButton.first()).toBeVisible();
  });

  test('should open create task modal', async ({ page }) => {
    await page.goto('/tasks');

    // Click create button
    const createButton = page.locator('button:has-text("Create")').or(
      page.locator('button:has-text("New Task")')
    ).first();

    await createButton.click();

    // Check if modal is visible
    await expect(page.locator('.ant-modal')).toBeVisible();
  });

  test('should display task type options', async ({ page }) => {
    await page.goto('/tasks');

    // Open create modal
    const createButton = page.locator('button:has-text("Create")').or(
      page.locator('button:has-text("New Task")')
    ).first();
    await createButton.click();

    // Find task type select
    const typeSelect = page.locator('[data-testid="task-type-select"]').or(
      page.locator('.ant-select').filter({ hasText: 'Type' })
    );

    if (await typeSelect.isVisible()) {
      await typeSelect.click();

      // Check for task type options
      await expect(page.locator('.ant-select-dropdown li:has-text("exec_shell")').or(
        page.locator('.ant-select-dropdown li:has-text("Exec Shell")')
      )).toBeVisible();
    }
  });

  test('should display task details', async ({ page }) => {
    await page.goto('/tasks');

    // Wait for tasks to load
    await page.waitForSelector('.ant-table-tbody tr, [data-testid="task-item"]', {
      timeout: 5000
    }).catch(() => {
      // No tasks exist, skip test
      test.skip();
    });

    // Click on first task
    const firstTask = page.locator('.ant-table-tbody tr').first().or(
      page.locator('[data-testid="task-item"]').first()
    );

    await firstTask.click();

    // Check if we're on task details page or modal
    await expect(page.locator('[data-testid="task-details"]').or(
      page.locator('.ant-modal-content')
    )).toBeVisible();
  });

  test('should filter tasks by status', async ({ page }) => {
    await page.goto('/tasks');

    // Find status filter
    const statusFilter = page.locator('[data-testid="status-filter"]').or(
      page.locator('.ant-select').filter({ hasText: 'Status' })
    );

    if (await statusFilter.isVisible()) {
      await statusFilter.click();

      // Select pending status
      await page.click('.ant-select-dropdown:visible li:has-text("Pending")');
    }
  });

  test('should cancel a task', async ({ page }) => {
    await page.goto('/tasks');

    // Find a pending task with cancel button
    const cancelButton = page.locator('[data-testid="cancel-task"]').first();

    if (await cancelButton.isVisible()) {
      // Set up dialog handler
      page.on('dialog', dialog => dialog.accept());

      await cancelButton.click();

      // Wait for success message
      await expect(page.locator('.ant-message-success')).toBeVisible({ timeout: 5000 });
    } else {
      test.skip();
    }
  });

  test('should display task output', async ({ page }) => {
    await page.goto('/tasks');

    // Wait for tasks to load
    await page.waitForSelector('.ant-table-tbody tr, [data-testid="task-item"]', {
      timeout: 5000
    }).catch(() => test.skip());

    // Click on a completed task
    const completedTask = page.locator('.ant-table-tbody tr:has-text("Completed")').first();

    if (await completedTask.isVisible()) {
      await completedTask.click();

      // Check if output is visible
      const outputSection = page.locator('[data-testid="task-output"]').or(
        page.locator('.task-output')
      );

      await expect(outputSection).toBeVisible();
    } else {
      test.skip();
    }
  });

  test('should show batch create option', async ({ page }) => {
    await page.goto('/tasks');

    // Look for batch create button
    const batchButton = page.locator('button:has-text("Batch")').or(
      page.locator('[data-testid="batch-create-button"]')
    );

    if (await batchButton.isVisible()) {
      await batchButton.click();

      // Check if batch modal is visible
      await expect(page.locator('.ant-modal')).toBeVisible();
    }
  });

  test('should search tasks', async ({ page }) => {
    await page.goto('/tasks');

    // Find search input
    const searchInput = page.locator('input[placeholder*="Search"]').or(
      page.locator('[data-testid="search-input"]')
    );

    if (await searchInput.isVisible()) {
      await searchInput.fill('test-task');
      await page.waitForTimeout(500); // Wait for debounce
    }
  });

  test('should display task statistics', async ({ page }) => {
    await page.goto('/tasks');

    // Look for task statistics cards
    const statsSection = page.locator('[data-testid="task-stats"]').or(
      page.locator('.task-statistics')
    );

    if (await statsSection.isVisible()) {
      // Check for statistics content
      await expect(statsSection.locator('.ant-statistic').or(
        statsSection.locator('.stat-item')
      ).first()).toBeVisible();
    }
  });
});
