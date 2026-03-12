import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display login page', async ({ page }) => {
    // Check if login form is visible
    await expect(page.locator('form')).toBeVisible();
    await expect(page.locator('input[type="text"]').first()).toBeVisible();
    await expect(page.locator('input[type="password"]').first()).toBeVisible();
  });

  test('should show validation error for empty fields', async ({ page }) => {
    // Click login button without entering credentials
    await page.click('button[type="submit"]');

    // Check for validation message
    await expect(page.locator('.ant-form-item-explain-error')).toBeVisible();
  });

  test('should show error for invalid credentials', async ({ page }) => {
    // Fill in invalid credentials
    await page.fill('input[type="text"]', 'invalid-user');
    await page.fill('input[type="password"]', 'wrong-password');

    // Submit form
    await page.click('button[type="submit"]');

    // Wait for error message
    await expect(page.locator('.ant-message-error')).toBeVisible({ timeout: 10000 });
  });

  test('should redirect to dashboard on successful login', async ({ page }) => {
    // Skip if no test credentials available
    test.skip(!process.env.TEST_USERNAME || !process.env.TEST_PASSWORD);

    // Fill in credentials
    await page.fill('input[type="text"]', process.env.TEST_USERNAME || 'admin');
    await page.fill('input[type="password"]', process.env.TEST_PASSWORD || 'password');

    // Submit form
    await page.click('button[type="submit"]');

    // Wait for redirect to dashboard
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  test('should persist login state', async ({ page }) => {
    // Skip if no test credentials available
    test.skip(!process.env.TEST_USERNAME || !process.env.TEST_PASSWORD);

    // Login
    await page.fill('input[type="text"]', process.env.TEST_USERNAME || 'admin');
    await page.fill('input[type="password"]', process.env.TEST_PASSWORD || 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // Reload page
    await page.reload();

    // Should still be on dashboard (not redirected to login)
    await expect(page).toHaveURL(/.*dashboard/);
  });

  test('should logout successfully', async ({ page }) => {
    // Skip if no test credentials available
    test.skip(!process.env.TEST_USERNAME || !process.env.TEST_PASSWORD);

    // Login first
    await page.fill('input[type="text"]', process.env.TEST_USERNAME || 'admin');
    await page.fill('input[type="password"]', process.env.TEST_PASSWORD || 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // Find and click logout button
    const logoutButton = page.locator('button:has-text("Logout")').or(
      page.locator('[data-testid="logout-button"]')
    );

    if (await logoutButton.isVisible()) {
      await logoutButton.click();

      // Should be redirected to login page
      await expect(page).toHaveURL(/.*login/, { timeout: 5000 });
    }
  });
});
