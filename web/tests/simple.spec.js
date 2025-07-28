import { test, expect } from '@playwright/test';

  test('should load the main page', async ({ page }) => {
    await page.goto('/');

    // Check title
    await expect(page).toHaveTitle('UAST Mapping Development Service');

    // Check main heading
    await expect(page.locator('h1')).toContainText('UAST Mapping Development Service');

    // Check that main sections exist
    await expect(page.locator('text=Code Input')).toBeVisible();
    await expect(page.locator('text=UAST Output & Query Results')).toBeVisible();
  });

  test('should have code editor and query input visible', async ({ page }) => {
    await page.goto('/');

    // Check code editor is visible
    await expect(page.locator('textarea')).toBeVisible();

    // Check query input is visible
    await expect(page.locator('input[placeholder*="Enter UAST query"]')).toBeVisible();
  });

  test('should show examples panel when clicking floating button', async ({ page }) => {
    await page.goto('/');

    // Click on floating examples button
    await page.locator('button[class*="rounded-full"]').click();

    // Check examples panel is visible
    await expect(page.locator('text=Example Queries')).toBeVisible();
  }); 