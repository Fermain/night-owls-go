import { render, screen, fireEvent } from '@testing-library/svelte';
import { expect, test, describe, vi } from 'vitest';
import ButtonTestWrapper from './ButtonTestWrapper.svelte';

describe('Button Component', () => {
	test('renders with default properties', () => {
		render(ButtonTestWrapper, { text: 'Click me' });

		const button = screen.getByRole('button', { name: 'Click me' });
		expect(button).toBeInTheDocument();
		expect(button).toHaveClass('inline-flex');
	});

	test('applies variant styles correctly', () => {
		const { rerender } = render(ButtonTestWrapper, {
			variant: 'destructive',
			text: 'Delete'
		});

		let button = screen.getByRole('button');
		expect(button).toHaveClass('bg-destructive');

		rerender({ variant: 'secondary', text: 'Cancel' });
		button = screen.getByRole('button');
		expect(button).toHaveClass('bg-secondary');
	});

	test('applies size styles correctly', () => {
		const { rerender } = render(ButtonTestWrapper, {
			size: 'sm',
			text: 'Small'
		});

		let button = screen.getByRole('button');
		expect(button).toHaveClass('h-8');

		rerender({ size: 'lg', text: 'Large' });
		button = screen.getByRole('button');
		expect(button).toHaveClass('h-10');
	});

	test('handles disabled state', () => {
		render(ButtonTestWrapper, {
			disabled: true,
			text: 'Disabled'
		});

		const button = screen.getByRole('button');
		expect(button).toBeDisabled();
		expect(button).toHaveClass('disabled:pointer-events-none');
	});

	test('handles click events', async () => {
		const handleClick = vi.fn();

		render(ButtonTestWrapper, {
			onclick: handleClick,
			text: 'Click me'
		});

		const button = screen.getByRole('button');
		await fireEvent.click(button);

		expect(handleClick).toHaveBeenCalledTimes(1);
	});

	test('does not fire click when disabled', async () => {
		const handleClick = vi.fn();

		render(ButtonTestWrapper, {
			disabled: true,
			onclick: handleClick,
			text: 'Disabled'
		});

		const button = screen.getByRole('button');
		await fireEvent.click(button);

		expect(handleClick).not.toHaveBeenCalled();
	});

	test('renders as anchor when href provided', () => {
		render(ButtonTestWrapper, {
			href: '/test',
			text: 'Link button'
		});

		const element = screen.getByRole('link');
		expect(element).toHaveAttribute('href', '/test');
	});

	test('renders as button by default', () => {
		render(ButtonTestWrapper, {
			text: 'Regular button'
		});

		const element = screen.getByRole('button');
		expect(element.tagName).toBe('BUTTON');
	});

	test('forwards additional props', () => {
		render(ButtonTestWrapper, {
			'data-testid': 'custom-button',
			'aria-label': 'Custom action',
			text: 'Test'
		});

		const button = screen.getByTestId('custom-button');
		expect(button).toHaveAttribute('aria-label', 'Custom action');
	});
});
