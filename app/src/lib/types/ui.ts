/**
 * UI-specific types for Night Owls application
 * These types are for component props, state management, and UI patterns
 */

import type { ComponentType } from 'svelte';
import type { HTMLAttributes } from 'svelte/elements';

// === COMPONENT PROPS ===

export interface BaseComponentProps {
	className?: string;
	id?: string;
	'data-testid'?: string;
}

export interface LoadingProps extends BaseComponentProps {
	isLoading: boolean;
	loadingText?: string;
	size?: 'sm' | 'md' | 'lg';
}

export interface ErrorProps extends BaseComponentProps {
	error: Error | string | null;
	title?: string;
	showRetry?: boolean;
	onRetry?: () => void;
}

export interface EmptyStateProps extends BaseComponentProps {
	title: string;
	description?: string;
	icon?: ComponentType;
	action?: {
		label: string;
		onClick: () => void;
	};
}

// === FORM TYPES ===

export interface FormFieldProps extends BaseComponentProps {
	label: string;
	name: string;
	required?: boolean;
	error?: string | null;
	disabled?: boolean;
	helperText?: string;
}

export interface SelectOption<T = string> {
	label: string;
	value: T;
	disabled?: boolean;
	description?: string;
}

export interface MultiSelectProps<T = string> extends FormFieldProps {
	options: SelectOption<T>[];
	selected: T[];
	placeholder?: string;
	searchable?: boolean;
	maxSelected?: number;
	onSelectionChange: (selected: T[]) => void;
}

export interface DateRangePickerProps extends FormFieldProps {
	value: { from: string | null; to: string | null };
	onChange: (range: { from: string | null; to: string | null }) => void;
	minDate?: string;
	maxDate?: string;
	placeholder?: string;
}

// === TABLE TYPES ===

export interface TableColumn<T = unknown> {
	key: string;
	label: string;
	sortable?: boolean;
	width?: string | number;
	align?: 'left' | 'center' | 'right';
	render?: (value: unknown, row: T) => string | ComponentType;
	className?: string;
}

export interface TableProps<T = unknown> extends BaseComponentProps {
	columns: TableColumn<T>[];
	data: T[];
	loading?: boolean;
	error?: string | null;
	emptyMessage?: string;
	keyField?: keyof T;
	sortBy?: string;
	sortDirection?: 'asc' | 'desc';
	onSort?: (key: string, direction: 'asc' | 'desc') => void;
	onRowClick?: (row: T) => void;
}

export interface PaginationProps extends BaseComponentProps {
	currentPage: number;
	totalPages: number;
	totalItems: number;
	itemsPerPage: number;
	onPageChange: (page: number) => void;
	showInfo?: boolean;
}

// === CARD TYPES ===

export interface CardProps extends BaseComponentProps {
	title?: string;
	subtitle?: string;
	action?: {
		label: string;
		onClick: () => void;
	};
	loading?: boolean;
	error?: string | null;
}

export interface MetricCardProps extends CardProps {
	value: string | number;
	change?: {
		value: number;
		period: string;
		trend: 'up' | 'down' | 'neutral';
	};
	icon?: ComponentType;
	color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'gray';
}

export interface StatusCardProps extends CardProps {
	status: 'success' | 'warning' | 'error' | 'info';
	description?: string;
	actions?: Array<{
		label: string;
		onClick: () => void;
		variant?: 'primary' | 'secondary' | 'outline';
	}>;
}

// === DIALOG TYPES ===

export interface DialogProps extends BaseComponentProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title?: string;
	description?: string;
	size?: 'sm' | 'md' | 'lg' | 'xl';
	closeOnEscape?: boolean;
	closeOnOutsideClick?: boolean;
}

export interface ConfirmDialogProps extends DialogProps {
	message: string;
	confirmLabel?: string;
	cancelLabel?: string;
	variant?: 'default' | 'destructive';
	onConfirm: () => void;
	onCancel?: () => void;
}

// === NAVIGATION TYPES ===

export interface NavigationItem {
	label: string;
	href?: string;
	icon?: ComponentType;
	badge?: string | number;
	children?: NavigationItem[];
	onClick?: () => void;
	disabled?: boolean;
	visible?: boolean;
}

export interface BreadcrumbItem {
	label: string;
	href?: string;
	current?: boolean;
}

export interface TabItem {
	key: string;
	label: string;
	icon?: ComponentType;
	disabled?: boolean;
	badge?: string | number;
}

export interface TabsProps extends BaseComponentProps {
	items: TabItem[];
	activeTab: string;
	onTabChange: (key: string) => void;
	variant?: 'default' | 'pills' | 'underline';
	size?: 'sm' | 'md' | 'lg';
}

// === NOTIFICATION TYPES ===

export interface NotificationItem {
	id: string;
	title: string;
	message?: string;
	type: 'info' | 'success' | 'warning' | 'error';
	timestamp: string;
	read: boolean;
	actions?: Array<{
		label: string;
		onClick: () => void;
	}>;
}

export interface ToastProps {
	id: string;
	title?: string;
	message: string;
	type: 'info' | 'success' | 'warning' | 'error';
	duration?: number;
	persistent?: boolean;
	action?: {
		label: string;
		onClick: () => void;
	};
	onDismiss: (id: string) => void;
}

// === LAYOUT TYPES ===

export interface PageHeaderProps extends BaseComponentProps {
	title: string;
	subtitle?: string;
	icon?: ComponentType;
	breadcrumbs?: BreadcrumbItem[];
	actions?: Array<{
		label: string;
		onClick: () => void;
		variant?: 'primary' | 'secondary' | 'outline';
		icon?: ComponentType;
	}>;
}

export interface SidebarProps extends BaseComponentProps {
	collapsed?: boolean;
	onToggle?: () => void;
	navigation: NavigationItem[];
	footer?: ComponentType;
}

// === FILTER TYPES ===

export interface FilterOption<T = string> {
	label: string;
	value: T;
	count?: number;
}

export interface FilterGroupProps<T = string> extends BaseComponentProps {
	title: string;
	options: FilterOption<T>[];
	selected: T[];
	multiple?: boolean;
	searchable?: boolean;
	collapsible?: boolean;
	defaultCollapsed?: boolean;
	onSelectionChange: (selected: T[]) => void;
}

// === TIMELINE TYPES ===

export interface TimelineItem {
	id: string;
	title: string;
	description?: string;
	timestamp: string;
	type?: 'default' | 'success' | 'warning' | 'error' | 'info';
	icon?: ComponentType;
	details?: Record<string, unknown>;
	metadata?: {
		user?: string;
		location?: string;
		ip?: string;
	};
}

export interface TimelineProps extends BaseComponentProps {
	items: TimelineItem[];
	loading?: boolean;
	error?: string | null;
	emptyMessage?: string;
	onItemClick?: (item: TimelineItem) => void;
	groupByDate?: boolean;
}

// === CHART TYPES ===

export interface ChartDataPoint {
	label: string;
	value: number;
	color?: string;
	metadata?: Record<string, unknown>;
}

export interface ChartProps extends BaseComponentProps {
	data: ChartDataPoint[];
	type: 'bar' | 'line' | 'pie' | 'doughnut';
	title?: string;
	height?: number;
	loading?: boolean;
	error?: string | null;
	showLegend?: boolean;
	responsive?: boolean;
}

// === SEARCH TYPES ===

export interface SearchProps extends BaseComponentProps {
	value: string;
	placeholder?: string;
	disabled?: boolean;
	loading?: boolean;
	debounceMs?: number;
	onSearch: (query: string) => void;
	onClear?: () => void;
}

export interface SearchResultItem<T = unknown> {
	id: string;
	title: string;
	subtitle?: string;
	description?: string;
	type?: string;
	data: T;
}

export interface SearchResultsProps<T = unknown> extends BaseComponentProps {
	query: string;
	results: SearchResultItem<T>[];
	loading?: boolean;
	error?: string | null;
	emptyMessage?: string;
	onResultClick: (result: SearchResultItem<T>) => void;
	groupBy?: (result: SearchResultItem<T>) => string;
}

// === ASYNC STATE TYPES ===

export interface AsyncState<T> {
	data: T | null;
	loading: boolean;
	error: Error | null;
	lastFetch?: Date;
}

export interface AsyncListState<T> extends AsyncState<T[]> {
	hasMore: boolean;
	loadingMore: boolean;
}

// === FORM VALIDATION TYPES ===

export interface ValidationRule<T = unknown> {
	required?: boolean | string;
	minLength?: { value: number; message: string };
	maxLength?: { value: number; message: string };
	pattern?: { value: RegExp; message: string };
	min?: { value: number; message: string };
	max?: { value: number; message: string };
	custom?: (value: T) => string | null;
}

export interface FieldState<T = unknown> {
	value: T;
	error: string | null;
	touched: boolean;
	dirty: boolean;
}

export interface FormState<T extends Record<string, unknown> = Record<string, unknown>> {
	values: T;
	errors: Partial<Record<keyof T, string>>;
	touched: Partial<Record<keyof T, boolean>>;
	dirty: boolean;
	valid: boolean;
	submitting: boolean;
}

// === THEME TYPES ===

export type ThemeMode = 'light' | 'dark' | 'system';

export interface ThemeContextType {
	mode: ThemeMode;
	setMode: (mode: ThemeMode) => void;
	isDark: boolean;
	colors: Record<string, string>;
}

// === RESPONSIVE TYPES ===

export type Breakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';

export interface ResponsiveValue<T> {
	base?: T;
	xs?: T;
	sm?: T;
	md?: T;
	lg?: T;
	xl?: T;
	'2xl'?: T;
}

// === COMPONENT VARIANTS ===

export type ButtonVariant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
export type ButtonSize = 'default' | 'sm' | 'lg' | 'icon';

export type InputVariant = 'default' | 'error' | 'success';
export type InputSize = 'sm' | 'md' | 'lg';

export type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';

// === UTILITY TYPES ===

export type ElementRef<T extends keyof HTMLElementTagNameMap> = HTMLElementTagNameMap[T] | null;

export interface WithClassName {
	className?: string;
}

export interface WithChildren {
	children?: import('svelte').Snippet;
}

export interface WithElementRef<T extends HTMLElement = HTMLElement> {
	ref?: T | null;
}

// Re-export common HTML attributes for convenience
export type ButtonAttributes = HTMLAttributes<HTMLButtonElement>;
export type InputAttributes = HTMLAttributes<HTMLInputElement>;
export type DivAttributes = HTMLAttributes<HTMLDivElement>;
export type FormAttributes = HTMLAttributes<HTMLFormElement>;
