import { test, expect } from '@playwright/test'

test('dashboard loads with navigation', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Agents', exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Sources', exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Tasks', exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Webhooks', exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Settings', exact: true })).toBeVisible()
})

test('can navigate to agents page', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('link', { name: 'Agents', exact: true }).click()
  await expect(page).toHaveURL(/\/agents/)
})

test('can navigate to settings page', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('link', { name: 'Settings' }).click()
  await expect(page).toHaveURL(/\/settings/)
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
})
