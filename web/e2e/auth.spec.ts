/**
 * @file auth.spec.ts
 * @description 认证流程 E2E 测试
 *
 * 测试场景：
 *  - 用户注册流程
 *  - 用户登录流程
 *  - 登录后跳转
 *
 * @author AhaVault Team
 * @created 2026-02-06
 */

import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  const testPassword = 'Test123456';

  test('should register a new user', async ({ page }) => {
    const testEmail = `test_${Date.now()}@example.com`;
    await page.goto('/register');

    // 填写注册表单 - 使用 placeholder 定位
    await page.getByPlaceholder('m@example.com').fill(testEmail);
    await page.getByPlaceholder('At least 8 characters').fill(testPassword);

    // 提交注册
    await page.getByRole('button', { name: /Create Account/i }).click();

    // 验证注册成功后跳转
    await expect(page).toHaveURL(/.*cabinet/i, { timeout: 10000 });
  });

  test('should login with valid credentials', async ({ page }) => {
    // 先注册一个用户
    const email = `login_${Date.now()}@example.com`;
    await page.goto('/register');
    await page.getByPlaceholder('m@example.com').fill(email);
    await page.getByPlaceholder('At least 8 characters').fill(testPassword);
    await page.getByRole('button', { name: /Create Account/i }).click();
    await expect(page).toHaveURL(/.*cabinet/i, { timeout: 10000 });

    // 退出登录 - 直接访问登录页
    await page.goto('/login');

    // 登录
    await page.getByPlaceholder('m@example.com').fill(email);
    await page.locator('input[type="password"]').fill(testPassword);
    // 使用表单内的提交按钮（type="submit"）
    await page.locator('button[type="submit"]').click();

    // 验证登录成功
    await expect(page).toHaveURL(/.*cabinet/i, { timeout: 10000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');

    // 使用无效凭据
    await page.getByPlaceholder('m@example.com').fill('invalid@example.com');
    await page.locator('input[type="password"]').fill('wrongpassword');
    // 使用表单内的提交按钮
    await page.locator('button[type="submit"]').click();

    // 验证错误提示
    await expect(page.locator('.bg-destructive\\/10')).toBeVisible({ timeout: 5000 });
  });
});
