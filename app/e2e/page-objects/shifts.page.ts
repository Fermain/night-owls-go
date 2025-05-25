import { type Page, type Locator, expect } from '@playwright/test';

export class ShiftsPage {
	readonly page: Page;

	// Locators
	readonly heading: Locator;
	readonly availableShiftsSection: Locator;
	readonly shiftCards: Locator;
	readonly bookShiftButton: Locator;
	readonly buddyNameInput: Locator;
	readonly confirmBookingButton: Locator;
	readonly successMessage: Locator;
	readonly errorMessage: Locator;
	readonly myBookingsLink: Locator;
	readonly cancelBookingButton: Locator;

	constructor(page: Page) {
		this.page = page;

		this.heading = page.getByRole('heading', { name: /available.*shifts|shifts/i });
		this.availableShiftsSection = page
			.getByTestId('available-shifts')
			.or(page.locator('.shifts-container'));
		this.shiftCards = page.locator('.shift-card, [data-testid^="shift-"]');
		this.bookShiftButton = page.getByRole('button', { name: /book.*shift|book/i });
		this.buddyNameInput = page.getByLabel(/buddy.*name|partner/i);
		this.confirmBookingButton = page.getByRole('button', { name: /confirm.*booking|book.*now/i });
		this.myBookingsLink = page.getByRole('link', { name: /my.*bookings|bookings/i });
		this.cancelBookingButton = page.getByRole('button', { name: /cancel.*booking|cancel/i });

		// Messages
		this.successMessage = page.getByText(/shift.*booked|booking.*successful/i);
		this.errorMessage = page.getByText(/booking.*failed|error/i);
	}

	async goto() {
		await this.page.goto('/shifts');
		await expect(this.heading).toBeVisible();
	}

	async gotoMyBookings() {
		await this.page.goto('/bookings/my');
	}

	async expectShiftsVisible() {
		await expect(this.availableShiftsSection).toBeVisible();
	}

	async getAvailableShiftsCount(): Promise<number> {
		await this.page.waitForLoadState('networkidle');
		return await this.shiftCards.count();
	}

	async clickFirstAvailableShift() {
		const firstShift = this.shiftCards.first();
		await expect(firstShift).toBeVisible();
		await firstShift.click();
	}

	async bookShift(shiftSelector?: string, buddyName?: string) {
		// If no specific shift selector, book the first available shift
		if (shiftSelector) {
			await this.page.locator(shiftSelector).click();
		} else {
			await this.clickFirstAvailableShift();
		}

		// Click book button
		await this.bookShiftButton.click();

		// Fill buddy name if provided
		if (buddyName) {
			await expect(this.buddyNameInput).toBeVisible();
			await this.buddyNameInput.fill(buddyName);
		}

		// Confirm booking
		await this.confirmBookingButton.click();
		await expect(this.successMessage).toBeVisible();
	}

	async expectBookingSuccess() {
		await expect(this.successMessage).toBeVisible();
	}

	async expectBookingError(errorText?: string) {
		if (errorText) {
			await expect(this.page.getByText(errorText)).toBeVisible();
		} else {
			await expect(this.errorMessage).toBeVisible();
		}
	}

	async expectShiftNotAvailable(shiftName: string) {
		await expect(this.page.getByText(shiftName)).not.toBeVisible();
	}

	async expectShiftDetails(details: {
		scheduleName?: string;
		startTime?: string;
		duration?: string;
		positionsAvailable?: number;
	}) {
		if (details.scheduleName) {
			await expect(this.page.getByText(details.scheduleName)).toBeVisible();
		}
		if (details.startTime) {
			await expect(this.page.getByText(details.startTime)).toBeVisible();
		}
		if (details.duration) {
			await expect(this.page.getByText(details.duration)).toBeVisible();
		}
		if (details.positionsAvailable) {
			await expect(this.page.getByText(`${details.positionsAvailable} position`)).toBeVisible();
		}
	}

	async cancelBooking(bookingId?: string) {
		if (bookingId) {
			const bookingRow = this.page.locator(`[data-booking-id="${bookingId}"]`);
			const cancelButton = bookingRow.getByRole('button', { name: /cancel/i });
			await cancelButton.click();
		} else {
			await this.cancelBookingButton.first().click();
		}

		// Confirm cancellation
		const confirmButton = this.page.getByRole('button', { name: /confirm|yes/i });
		await confirmButton.click();
	}

	async expectBookingInMyBookings(shiftName: string) {
		await this.gotoMyBookings();
		await expect(this.page.getByText(shiftName)).toBeVisible();
	}

	async expectNoBookings() {
		await this.gotoMyBookings();
		await expect(this.page.getByText(/no.*bookings|no.*shifts/i)).toBeVisible();
	}

	async expectBuddyDisplayed(buddyName: string) {
		await expect(this.page.getByText(buddyName)).toBeVisible();
	}

	async filterShiftsByDate(startDate: string, endDate?: string) {
		const dateFilter = this.page.getByLabel(/date.*filter|from.*date/i);
		await dateFilter.fill(startDate);

		if (endDate) {
			const endDateFilter = this.page.getByLabel(/to.*date|end.*date/i);
			await endDateFilter.fill(endDate);
		}

		const applyButton = this.page.getByRole('button', { name: /apply|filter/i });
		await applyButton.click();
	}

	async searchShifts(searchTerm: string) {
		const searchInput = this.page.getByPlaceholder(/search.*shifts/i);
		await searchInput.fill(searchTerm);
	}
}
