import { test, expect } from '@playwright/test';

test.describe('API Integration - Simple Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should parse Go code successfully', async ({ page }) => {
    // Wait for the page to load
    await page.waitForLoadState('networkidle');
    
    // Click Parse Code button
    await page.getByRole('button', { name: 'Parse Code' }).click();
    
    // Wait for parsing to complete (button should return to normal state)
    await expect(page.getByRole('button', { name: 'Parse Code' })).toBeVisible({ timeout: 10000 });
    
    // Check that some output appears (any pre element)
    const preElements = page.locator('pre');
    await expect(preElements.first()).toBeVisible();
    
    // The output should contain some JSON structure
    const outputText = await preElements.first().textContent();
    expect(outputText).toContain('"');
  });

  test('should show error when parsing empty code', async ({ page }) => {
    // Clear the code editor
    await page.locator('textarea[id="code"]').clear();
    
    // Click Parse Code button
    await page.getByRole('button', { name: 'Parse Code' }).click();
    
    // Should show some error indication (toast or alert)
    // Wait a bit for error to appear
    await page.waitForTimeout(2000);
    
    // Check if any error message appears
    const errorElements = page.locator('[role="alert"], .error, [aria-live="assertive"]');
    await expect(errorElements.first()).toBeVisible({ timeout: 5000 });
  });

  test('should execute query successfully', async ({ page }) => {
    // First parse some code
    await page.getByRole('button', { name: 'Parse Code' }).click();
    await expect(page.getByRole('button', { name: 'Parse Code' })).toBeVisible({ timeout: 10000 });
    
    // Switch to Query tab
    await page.getByRole('tab', { name: 'Query' }).click();
    
    // Enter a simple query
    await page.locator('input[id="query"]').fill('rfilter(.type == "Package")');
    
    // Click Execute Query button
    await page.getByRole('button', { name: 'Execute Query' }).click();
    
    // Wait for query to complete
    await expect(page.getByRole('button', { name: 'Execute Query' })).toBeVisible({ timeout: 10000 });
    
    // Check that some output appears
    const preElements = page.locator('pre');
    await expect(preElements.first()).toBeVisible();
  });

  test('should change language and parse different code', async ({ page }) => {
    // Change language to Python
    await page.locator('[role="combobox"]').click();
    await page.locator('text=Python').click();
    
    // Wait for code to change
    await expect(page.locator('textarea[id="code"]')).toContainText('def main():');
    
    // Parse the Python code
    await page.getByRole('button', { name: 'Parse Code' }).click();
    
    // Wait for parsing to complete
    await expect(page.getByRole('button', { name: 'Parse Code' })).toBeVisible({ timeout: 10000 });
    
    // Check that some output appears
    const preElements = page.locator('pre');
    await expect(preElements.first()).toBeVisible();
  });
}); 