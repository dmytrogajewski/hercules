import { test, expect } from '@playwright/test';

test.describe('API Integration Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should parse Go code successfully', async ({ page }) => {
    // Wait for the page to load
    await page.waitForLoadState('networkidle');
    
    // Click Parse Code button
    await page.locator('button:has-text("Parse Code")').click();
    
    // Wait for parsing to complete (button should return to normal state)
    await expect(page.locator('button:has-text("Parse Code")')).toBeVisible({ timeout: 10000 });
    
    // Check that UAST output contains some content
    const uastOutput = page.locator('pre:has-text("UAST")');
    await expect(uastOutput).toBeVisible();
    
    // The output should contain UAST structure
    const outputText = await page.locator('pre').first().textContent();
    expect(outputText).toContain('"type"');
    expect(outputText).toContain('"children"');
  });

  test('should show error when parsing empty code', async ({ page }) => {
    // Clear the code editor
    await page.locator('textarea[id="code"]').clear();
    
    // Click Parse Code button
    await page.locator('button:has-text("Parse Code")').click();
    
    // Should show error toast
    await expect(page.locator('text=Error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Please enter some code to parse')).toBeVisible();
  });

  test('should execute query successfully', async ({ page }) => {
    // First parse some code
    await page.locator('button:has-text("Parse Code")').click();
    await expect(page.locator('button:has-text("Parse Code")')).toBeVisible({ timeout: 10000 });
    
    // Switch to Query tab
    await page.locator('text=Query').click();
    
    // Enter a simple query
    await page.locator('input[id="query"]').fill('rfilter(.type == "Package")');
    
    // Click Execute Query button
    await page.locator('button:has-text("Execute Query")').click();
    
    // Wait for query to complete
    await expect(page.locator('button:has-text("Execute Query")')).toBeVisible({ timeout: 10000 });
    
    // Check that query results contain some content
    const queryOutput = page.locator('pre:has-text("results")');
    await expect(queryOutput).toBeVisible();
  });

  test('should show error when executing query without parsing first', async ({ page }) => {
    // Switch to Query tab
    await page.locator('text=Query').click();
    
    // Enter a query
    await page.locator('input[id="query"]').fill('rfilter(.type == "Test")');
    
    // Click Execute Query button
    await page.locator('button:has-text("Execute Query")').click();
    
    // Should show error toast
    await expect(page.locator('text=Error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Please parse some code first')).toBeVisible();
  });

  test('should show error when executing empty query', async ({ page }) => {
    // First parse some code
    await page.locator('button:has-text("Parse Code")').click();
    await expect(page.locator('button:has-text("Parse Code")')).toBeVisible({ timeout: 10000 });
    
    // Switch to Query tab
    await page.locator('text=Query').click();
    
    // Click Execute Query button without entering query
    await page.locator('button:has-text("Execute Query")').click();
    
    // Should show error toast
    await expect(page.locator('text=Error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Please enter a query')).toBeVisible();
  });

  test('should change language and parse different code', async ({ page }) => {
    // Change language to Python
    await page.locator('[role="combobox"]').click();
    await page.locator('text=Python').click();
    
    // Wait for code to change
    await expect(page.locator('textarea[id="code"]')).toContainText('def main():');
    
    // Parse the Python code
    await page.locator('button:has-text("Parse Code")').click();
    
    // Wait for parsing to complete
    await expect(page.locator('button:has-text("Parse Code")')).toBeVisible({ timeout: 10000 });
    
    // Check that UAST output contains content
    const uastOutput = page.locator('pre:has-text("UAST")');
    await expect(uastOutput).toBeVisible();
  });

  test('should copy example query and execute it', async ({ page }) => {
    // First parse some code
    await page.locator('button:has-text("Parse Code")').click();
    await expect(page.locator('button:has-text("Parse Code")')).toBeVisible({ timeout: 10000 });
    
    // Switch to Examples tab
    await page.locator('text=Examples').click();
    
    // Click on an example query
    await page.locator('text=Find all Import nodes').click();
    
    // Switch to Query tab
    await page.locator('text=Query').click();
    
    // Verify query was copied
    await expect(page.locator('input[id="query"]')).toHaveValue('rfilter(.type == "Import")');
    
    // Execute the query
    await page.locator('button:has-text("Execute Query")').click();
    
    // Wait for query to complete
    await expect(page.locator('button:has-text("Execute Query")')).toBeVisible({ timeout: 10000 });
    
    // Check that query results are visible
    const queryOutput = page.locator('pre:has-text("results")');
    await expect(queryOutput).toBeVisible();
  });
}); 