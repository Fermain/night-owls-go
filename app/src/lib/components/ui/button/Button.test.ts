import { render, screen, fireEvent } from '@testing-library/svelte';
import { expect, test, describe, vi } from 'vitest';
import Button from './button.svelte';

describe('Button Component', () => {
  test('renders with default properties', () => {
    render(Button, { children: 'Click me' });
    
    const button = screen.getByRole('button', { name: 'Click me' });
    expect(button).toBeInTheDocument();
    expect(button).toHaveClass('inline-flex'); // Based on shadcn/ui button classes
  });

  test('applies variant styles correctly', () => {
    const { rerender } = render(Button, { 
      variant: 'destructive',
      children: 'Delete' 
    });
    
    let button = screen.getByRole('button');
    expect(button).toHaveClass('bg-destructive');
    
    // Test secondary variant
    rerender({ variant: 'secondary', children: 'Cancel' });
    button = screen.getByRole('button');
    expect(button).toHaveClass('bg-secondary');
  });

  test('applies size styles correctly', () => {
    const { rerender } = render(Button, { 
      size: 'sm',
      children: 'Small' 
    });
    
    let button = screen.getByRole('button');
    expect(button).toHaveClass('h-9');
    
    // Test large size
    rerender({ size: 'lg', children: 'Large' });
    button = screen.getByRole('button');
    expect(button).toHaveClass('h-11');
  });

  test('handles disabled state', () => {
    render(Button, { 
      disabled: true,
      children: 'Disabled' 
    });
    
    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
    expect(button).toHaveClass('disabled:pointer-events-none');
  });

  test('handles click events', async () => {
    const handleClick = vi.fn();
    
    render(Button, { 
      onclick: handleClick,
      children: 'Click me' 
    });
    
    const button = screen.getByRole('button');
    await fireEvent.click(button);
    
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  test('does not fire click when disabled', async () => {
    const handleClick = vi.fn();
    
    render(Button, { 
      disabled: true,
      onclick: handleClick,
      children: 'Disabled' 
    });
    
    const button = screen.getByRole('button');
    await fireEvent.click(button);
    
    expect(handleClick).not.toHaveBeenCalled();
  });

  test('renders as different HTML elements', () => {
    const { rerender } = render(Button, { 
      href: '/test',
      children: 'Link button' 
    });
    
    // Should render as anchor when href is provided
    let element = screen.getByRole('link');
    expect(element).toHaveAttribute('href', '/test');
    
    // Should render as button by default
    rerender({ children: 'Regular button' });
    element = screen.getByRole('button');
    expect(element.tagName).toBe('BUTTON');
  });

  test('forwards additional props', () => {
    render(Button, { 
      'data-testid': 'custom-button',
      'aria-label': 'Custom action',
      children: 'Test' 
    });
    
    const button = screen.getByTestId('custom-button');
    expect(button).toHaveAttribute('aria-label', 'Custom action');
  });

  test('handles loading state (if implemented)', () => {
    // This test assumes a loading prop exists
    render(Button, { 
      loading: true,
      children: 'Loading...' 
    });
    
    const button = screen.getByRole('button');
    // Check if loading indicator is shown
    // expect(button).toHaveClass('loading'); // Adjust based on implementation
    expect(button).toBeDisabled(); // Loading buttons should be disabled
  });
}); 