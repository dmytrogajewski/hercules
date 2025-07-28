import { test, expect } from '@playwright/test';

test.describe('UAST Development Service', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should load the main page with correct title', async ({ page }) => {
    await expect(page).toHaveTitle('UAST Mapping Development Service');
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('UAST Mapping Development Service');
    
    // Check subtitle
    await expect(page.locator('p')).toContainText('Interactive development environment for UAST mappings');
  });

  test('should display all three tabs', async ({ page }) => {
    // Check tab buttons exist
    await expect(page.locator('[role="tab"]')).toHaveCount(3);
    await expect(page.locator('text=Code Editor')).toBeVisible();
    await expect(page.locator('text=Query')).toBeVisible();
    await expect(page.locator('text=Examples')).toBeVisible();
  });

  test('should start with Code Editor tab active', async ({ page }) => {
    // Code Editor tab should be active by default
    await expect(page.locator('[role="tab"][aria-selected="true"]')).toContainText('Code Editor');
    
    // Code Editor content should be visible
    await expect(page.locator('text=Code Input')).toBeVisible();
    await expect(page.locator('text=UAST Output')).toBeVisible();
  });

  test('should have language selector with Go as default', async ({ page }) => {
    const languageSelect = page.locator('[role="combobox"]');
    await expect(languageSelect).toBeVisible();
    
    // Check that Go is selected by default
    await expect(languageSelect).toContainText('Go');
  });

  test('should have code editor textarea', async ({ page }) => {
    const codeTextarea = page.locator('textarea[id="code"]');
    await expect(codeTextarea).toBeVisible();
    
    // Check that it contains the default Go code
    await expect(codeTextarea).toContainText('package main');
    await expect(codeTextarea).toContainText('func main()');
  });

  test('should have Parse Code button', async ({ page }) => {
    const parseButton = page.locator('button:has-text("Parse Code")');
    await expect(parseButton).toBeVisible();
    await expect(parseButton).toBeEnabled();
  });

  test('should switch to Query tab when clicked', async ({ page }) => {
    // Click on Query tab
    await page.locator('text=Query').click();
    
    // Query tab should be active
    await expect(page.locator('[role="tab"][aria-selected="true"]')).toContainText('Query');
    
    // Query content should be visible
    await expect(page.locator('text=UAST Query')).toBeVisible();
    await expect(page.locator('input[id="query"]')).toBeVisible();
    await expect(page.locator('button:has-text("Execute Query")')).toBeVisible();
  });

  test('should switch to Examples tab when clicked', async ({ page }) => {
    // Click on Examples tab
    await page.locator('text=Examples').click();
    
    // Examples tab should be active
    await expect(page.locator('[role="tab"][aria-selected="true"]')).toContainText('Examples');
    
    // Examples content should be visible
    await expect(page.locator('text=Example Queries')).toBeVisible();
    
    // Check for example query cards
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
    const codeTextarea = page.locator('textarea[id="code"]');
    await expect(codeTextarea).toContainText('def main():');
    await expect(codeTextarea).toContainText('print("Hello, World!")');
  });

  test('should show loading state when parsing code', async ({ page }) => {
    // Click Parse Code button
    await page.locator('button:has-text("Parse Code")').click();
    
    // Should show loading state
    await expect(page.locator('button:has-text("Parsing...")')).toBeVisible();
    await expect(page.locator('button:has-text("Parsing...")')).toBeDisabled();
  });

  test('should show loading state when executing query', async ({ page }) => {
    // Switch to Query tab
    await page.locator('text=Query').click();
    
    // Enter a query
    await page.locator('input[id="query"]').fill('rfilter(.type == "Test")');
    
    // Click Execute Query button
    await page.locator('button:has-text("Execute Query")').click();
    
    // Should show loading state
    await expect(page.locator('button:has-text("Executing...")')).toBeVisible();
    await expect(page.locator('button:has-text("Executing...")')).toBeDisabled();
  });

  test('should copy example query when clicked', async ({ page }) => {
    // Switch to Examples tab
    await page.locator('text=Examples').click();
    
    // Click on first example query
    await page.locator('text=Find all Import nodes').click();
    
    // Switch back to Query tab
    await page.locator('text=Query').click();
    
    // Check that query was copied to input
    await expect(page.locator('input[id="query"]')).toHaveValue('rfilter(.type == "Import")');
  });

  test('should execute query on Enter key', async ({ page }) => {
    // Switch to Query tab
    await page.locator('text=Query').click();
    
    // Enter a query and press Enter
    await page.locator('input[id="query"]').fill('rfilter(.type == "Test")');
    await page.locator('input[id="query"]').press('Enter');
    
    // Should show loading state
    await expect(page.locator('button:has-text("Executing...")')).toBeVisible();
  });

  test('should be responsive on mobile viewport', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Check that tabs are still accessible
    await expect(page.locator('text=Code Editor')).toBeVisible();
    await expect(page.locator('text=Query')).toBeVisible();
    await expect(page.locator('text=Examples')).toBeVisible();
    
    // Check that main content is visible
    await expect(page.locator('text=Code Input')).toBeVisible();
  });
}); 