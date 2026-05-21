import { describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { GuitarForm } from './GuitarForm';

describe('<GuitarForm />', () => {
  it('fills the form and calls onSubmit with API-shaped values (price in minor units)', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<GuitarForm onSubmit={onSubmit} submitLabel="Create" />);

    await user.type(screen.getByLabelText(/brand/i), 'Fender');
    await user.type(screen.getByLabelText(/^type$/i), 'Stratocaster');

    const buildYear = screen.getByLabelText(/build year/i);
    await user.clear(buildYear);
    await user.type(buildYear, '1996');

    await user.type(screen.getByLabelText(/serial number/i), 'SN-1234');
    await user.type(screen.getByLabelText(/^price$/i), '1999.00');

    await user.type(
      screen.getByLabelText(/add picture url/i),
      'https://example.com/strat.jpg',
    );
    await user.click(screen.getByRole('button', { name: /add picture url/i }));

    await user.type(screen.getByLabelText(/description/i), 'A fine guitar.');

    await user.click(screen.getByRole('button', { name: /create/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      brand: 'Fender',
      typeName: 'Stratocaster',
      buildYear: 1996,
      priceAmount: 199900,
      priceCurrency: 'EUR',
      pictures: ['https://example.com/strat.jpg'],
      serialNumber: 'SN-1234',
      description: 'A fine guitar.',
    });
  });

  it('does not submit when required fields are missing', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<GuitarForm onSubmit={onSubmit} submitLabel="Create" />);
    await user.click(screen.getByRole('button', { name: /create/i }));

    await waitFor(() => {
      expect(screen.getByText(/brand is required/i)).toBeInTheDocument();
    });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('surfaces server-side validation errors passed in as serverError', () => {
    render(
      <GuitarForm
        onSubmit={() => {}}
        serverError="buildYear must be between 1800 and 2027"
      />,
    );
    expect(
      screen.getByText(/buildYear must be between 1800 and 2027/i),
    ).toBeInTheDocument();
  });
});
