import { test, expect } from '@playwright/test';

test.describe('UAST Development Service - Basic Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should load the main page with correct title', async ({ page }) => {
    await expect(page).toHaveTitle('UAST Mapping Development Service');

    // Check main heading
    await expect(page.locator('h1')).toContainText('UAST Mapping Development Service');

    // Check subtitle - use more specific selector
    await expect(page.locator('p.text-sm')).toContainText('Interactive development environment for UAST mappings');
  });

  test('should display main layout components', async ({ page }) => {
    // Check that main sections exist
    await expect(page.locator('text=Code Input')).toBeVisible();
    await expect(page.locator('text=UAST Output & Query Results')).toBeVisible();
  });

  test('should have language selector with Go as default', async ({ page }) => {
    const languageSelect = page.locator('[role="combobox"]');
    await expect(languageSelect).toBeVisible();
    
    // Check that Go is selected by default
    await expect(languageSelect).toContainText('Go');
  });

  test('should have code editor textarea', async ({ page }) => {
    const codeTextarea = page.locator('textarea');
    await expect(codeTextarea).toBeVisible();
    
    // Check that it contains the default Go code
    await expect(codeTextarea).toContainText('package main');
    await expect(codeTextarea).toContainText('func main()');
  });

  test('should have query input field', async ({ page }) => {
    const queryInput = page.locator('input[placeholder*="Enter UAST query"]');
    await expect(queryInput).toBeVisible();
    await expect(queryInput).toBeEnabled();
  });

  test('should display floating examples button', async ({ page }) => {
    // Check that floating examples button exists
    await expect(page.locator('button[class*="rounded-full"]')).toBeVisible();
  });

  test('should show examples panel when clicking floating button', async ({ page }) => {
    // Click on floating examples button
    await page.locator('button[class*="rounded-full"]').click();

    // Check examples panel is visible
    await expect(page.locator('text=Example Queries')).toBeVisible();
    
    // Check for example query buttons
    await expect(page.locator('text=Find all Import nodes')).toBeVisible();
    await expect(page.locator('text=Find all Function declarations')).toBeVisible();
  });

  test('should change language when selected', async ({ page }) => {
    // Click on language selector
    await page.locator('[role="combobox"]').click();
    
    // Select Python
    await page.locator('text=Python').click();
    
    // Check that Python is selected
    await expect(page.locator('[role="combobox"]')).toContainText('Python');
    
    // Check that code editor content changed to Python
    const codeTextarea = page.locator('textarea');
    await expect(codeTextarea).toContainText('def main():');
    await expect(codeTextarea).toContainText('print("Hello, World!")');
  });

  test('should be responsive on mobile viewport', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Check that main sections are still visible
    await expect(page.locator('text=Code Input')).toBeVisible();
    await expect(page.locator('text=UAST Output & Query Results')).toBeVisible();
    
    // Check that floating button is still visible
    await expect(page.locator('button[class*="rounded-full"]')).toBeVisible();
  });
}); 