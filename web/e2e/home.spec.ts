/**
 * @file home.spec.ts
 * @description 首页 E2E 测试
 *
 * 测试场景：
 *  - 页面正确渲染
 *  - 取件码输入与验证
 *  - 导航到登录/注册页面
 *
 * @author AhaVault Team
 * @created 2026-02-06
 */

import { test, expect } from '@playwright/test';

test.describe('Home Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should render landing page correctly', async ({ page }) => {
    // 检查主标题
    await expect(page.getByText('Your Data,')).toBeVisible();
    await expect(page.getByText('Truly Secure.')).toBeVisible();

    // 检查取件码输入框
    await expect(page.getByPlaceholder('AHA-XXXX-XXXX')).toBeVisible();

    // 检查取件按钮
    await expect(page.getByRole('button', { name: /Retrieve File/i })).toBeVisible();
  });

  test('should show validation error for short pickup code', async ({ page }) => {
    // 输入过短的取件码
    await page.getByPlaceholder('AHA-XXXX-XXXX').fill('SHORT');

    // 点击取件按钮
    await page.getByRole('button', { name: /Retrieve File/i }).click();

    // 检查错误提示
    await expect(page.getByText('Please enter a valid 8-digit code')).toBeVisible();
  });

  test('should navigate to login page', async ({ page }) => {
    // 点击登录链接
    await page.getByRole('link', { name: /Login/i }).click();

    // 验证跳转到登录页
    await expect(page).toHaveURL(/.*login/);
  });

  test('should navigate to register page', async ({ page }) => {
    // 点击注册链接
    await page.getByRole('link', { name: /Register/i }).click();

    // 验证跳转到注册页
    await expect(page).toHaveURL(/.*register/);
  });
});
