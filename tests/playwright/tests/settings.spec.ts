import { test, expect } from '@playwright/test'

test('settings page shows tabs', async ({ page }) => {
  await page.goto('/settings')

  await expect(page.getByRole('button', { name: 'API Tokens' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Service Accounts' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Organization', exact: true })).toBeVisible()
})

test('can switch to organization tab', async ({ page }) => {
  await page.goto('/settings')
  await page.getByRole('button', { name: 'Organization', exact: true }).click()

  await expect(page.getByRole('heading', { name: 'Details' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Members' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Teams' })).toBeVisible()
})

test('can create an API token', async ({ page }) => {
  await page.goto('/settings')

  const tokenName = `pw-token-${Date.now()}`
  await page.getByPlaceholder('Token name').fill(tokenName)
  await page.getByRole('button', { name: 'Create' }).click()

  // Should show the token value.
  await expect(page.getByText('Token created')).toBeVisible({ timeout: 5_000 })
  await expect(page.getByText('plt_')).toBeVisible()
})
