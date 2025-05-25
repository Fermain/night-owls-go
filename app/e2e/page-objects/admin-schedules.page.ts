import { type Page, type Locator, expect } from '@playwright/test';

export class AdminSchedulesPage {
  readonly page: Page;

  // Locators
  readonly heading: Locator;
  readonly createScheduleButton: Locator;
  readonly searchInput: Locator;
  readonly schedulesTable: Locator;
  readonly scheduleNameInput: Locator;
  readonly scheduleDescriptionInput: Locator;
  readonly cronExpressionInput: Locator;
  readonly durationInput: Locator;
  readonly positionsInput: Locator;
  readonly timezoneSelect: Locator;
  readonly saveButton: Locator;
  readonly cancelButton: Locator;
  readonly successMessage: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    
    this.heading = page.getByRole('heading', { name: /schedules/i });
    this.createScheduleButton = page.getByRole('button', { name: /create.*schedule/i });
    this.searchInput = page.getByPlaceholder(/search.*schedule/i);
    this.schedulesTable = page.getByRole('table');
    
    // Form inputs
    this.scheduleNameInput = page.getByLabel(/name/i);
    this.scheduleDescriptionInput = page.getByLabel(/description/i);
    this.cronExpressionInput = page.getByLabel(/cron.*expression/i);
    this.durationInput = page.getByLabel(/duration/i);
    this.positionsInput = page.getByLabel(/positions/i);
    this.timezoneSelect = page.getByLabel(/timezone/i);
    
    // Actions
    this.saveButton = page.getByRole('button', { name: /save|create/i });
    this.cancelButton = page.getByRole('button', { name: /cancel/i });
    
    // Messages
    this.successMessage = page.getByText(/schedule.*created|schedule.*updated/i);
    this.errorMessage = page.getByText(/error|failed/i);
  }

  async goto() {
    await this.page.goto('/admin/schedules');
    await expect(this.heading).toBeVisible();
  }

  async clickCreateSchedule() {
    await this.createScheduleButton.click();
  }

  async fillScheduleForm(schedule: {
    name: string;
    description: string;
    cronExpression: string;
    duration: number;
    positions: number;
    timezone?: string;
  }) {
    await this.scheduleNameInput.fill(schedule.name);
    await this.scheduleDescriptionInput.fill(schedule.description);
    await this.cronExpressionInput.fill(schedule.cronExpression);
    await this.durationInput.fill(schedule.duration.toString());
    await this.positionsInput.fill(schedule.positions.toString());
    
    if (schedule.timezone) {
      await this.timezoneSelect.selectOption(schedule.timezone);
    }
  }

  async saveSchedule() {
    await this.saveButton.click();
  }

  async createSchedule(schedule: {
    name: string;
    description: string;
    cronExpression: string;
    duration: number;
    positions: number;
    timezone?: string;
  }) {
    await this.clickCreateSchedule();
    await this.fillScheduleForm(schedule);
    await this.saveSchedule();
    await expect(this.successMessage).toBeVisible();
  }

  async searchSchedules(searchTerm: string) {
    await this.searchInput.fill(searchTerm);
  }

  async expectScheduleInList(scheduleName: string) {
    await expect(this.page.getByText(scheduleName)).toBeVisible();
  }

  async expectScheduleNotInList(scheduleName: string) {
    await expect(this.page.getByText(scheduleName)).not.toBeVisible();
  }

  async clickSchedule(scheduleName: string) {
    await this.page.getByText(scheduleName).click();
  }

  async deleteSchedule(scheduleName: string) {
    const scheduleRow = this.page.getByRole('row').filter({ hasText: scheduleName });
    const deleteButton = scheduleRow.getByRole('button', { name: /delete/i });
    await deleteButton.click();
    
    // Confirm deletion
    const confirmButton = this.page.getByRole('button', { name: /confirm|yes|delete/i });
    await confirmButton.click();
  }

  async editSchedule(scheduleName: string, updates: {
    name?: string;
    description?: string;
    cronExpression?: string;
    duration?: number;
    positions?: number;
  }) {
    const scheduleRow = this.page.getByRole('row').filter({ hasText: scheduleName });
    const editButton = scheduleRow.getByRole('button', { name: /edit/i });
    await editButton.click();

    if (updates.name) {
      await this.scheduleNameInput.clear();
      await this.scheduleNameInput.fill(updates.name);
    }
    if (updates.description) {
      await this.scheduleDescriptionInput.clear();
      await this.scheduleDescriptionInput.fill(updates.description);
    }
    if (updates.cronExpression) {
      await this.cronExpressionInput.clear();
      await this.cronExpressionInput.fill(updates.cronExpression);
    }
    if (updates.duration) {
      await this.durationInput.clear();
      await this.durationInput.fill(updates.duration.toString());
    }
    if (updates.positions) {
      await this.positionsInput.clear();
      await this.positionsInput.fill(updates.positions.toString());
    }

    await this.saveButton.click();
    await expect(this.successMessage).toBeVisible();
  }

  async expectValidationError(field: string) {
    const errorSelector = `[data-field="${field}"] .error, .field-error, .form-error`;
    await expect(this.page.locator(errorSelector)).toBeVisible();
  }

  async expectFormError(message: string) {
    await expect(this.page.getByText(message)).toBeVisible();
  }

  async getScheduleCount(): Promise<number> {
    const rows = await this.page.getByRole('row').count();
    return Math.max(0, rows - 1); // Subtract header row
  }
} 