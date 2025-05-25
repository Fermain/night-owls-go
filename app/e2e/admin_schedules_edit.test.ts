import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

test.describe('Admin › Schedule Slots – DateRangePicker', () => {
  const visitSlots = async (page) => {
    await page.goto('/admin/schedules/slots');
    // page-ready signal instead of network-idle
    await expect(
      page.getByRole('heading', { name: /schedule slots/i })
    ).toBeVisible();
  };

  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await visitSlots(page);
  });

  test('allows picking and re-picking a date range', async ({ page }) => {

    const rangeButton = page.getByRole('button', { name: /date range/i }).first();
    await expect(rangeButton).toBeVisible();

    // Helper to pick any two distinct visible days in the current calendar view
    const pickRange = async (startIdx: number, endIdx: number) => {
      await rangeButton.click();

      const dialog = page.getByRole('dialog').filter({
        has: page.getByRole('grid') // ensures it’s the calendar, not another dialog
      });
      await expect(dialog).toBeVisible();

      const days = dialog.getByRole('gridcell').filter({
        hasNot: page.locator('[aria-disabled="true"], .outside') // ignore disabled/outside-month
      });

      await days.nth(startIdx).click();
      await days.nth(endIdx).click();

      // Calendar closes after second click
      await expect(dialog).toBeHidden();
    };

    // --- First selection ---
    await pickRange(0, 5);                 // first + sixth visible day
    const firstValue = (await rangeButton.textContent())!.trim();
    expect(firstValue).not.toEqual('');

    // --- Second selection (re-select) ---
    await pickRange(10, 15);               // later in the month
    const secondValue = (await rangeButton.textContent())!.trim();

    expect(secondValue).not.toEqual('');
    expect(secondValue).not.toEqual(firstValue);
  });
});
