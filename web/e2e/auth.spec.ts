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
  const testEmail = `test_${Date.now()}@example.com`;
  const testPassword = 'Test123456';

  test('should register a new user', async ({ page }) => {
    await page.goto('/register');

    // 填写注册表单
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByLabel(/password/i).first().fill(testPassword);

    // 提交注册
    await page.getByRole('button', { name: /register|sign up/i }).click();

    // 验证注册成功后跳转
    await expect(page).toHaveURL(/.*cabinet|.*dashboard/i, { timeout: 10000 });
  });

  test('should login with valid credentials', async ({ page }) => {
    // 先注册一个用户
    await page.goto('/register');
    const email = `login_${Date.now()}@example.com`;
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/password/i).first().fill(testPassword);
    await page.getByRole('button', { name: /register|sign up/i }).click();
    await expect(page).toHaveURL(/.*cabinet|.*dashboard/i, { timeout: 10000 });

    // 退出登录（如果有退出按钮）
    const logoutButton = page.getByRole('button', { name: /logout|sign out/i });
    if (await logoutButton.isVisible()) {
      await logoutButton.click();
    } else {
      // 直接访问登录页
      await page.goto('/login');
    }

    // 登录
    await page.goto('/login');
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/password/i).fill(testPassword);
    await page.getByRole('button', { name: /login|sign in/i }).click();

    // 验证登录成功
    await expect(page).toHaveURL(/.*cabinet|.*dashboard/i, { timeout: 10000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');

    // 使用无效凭据
    await page.getByLabel(/email/i).fill('invalid@example.com');
    await page.getByLabel(/password/i).fill('wrongpassword');
    await page.getByRole('button', { name: /login|sign in/i }).click();

    // 验证错误提示
    await expect(page.getByText(/invalid|error|failed/i)).toBeVisible({ timeout: 5000 });
  });
});
