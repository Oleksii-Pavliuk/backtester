import Form from '@/app/ui/invoices/create-form';
import Breadcrumbs from '@/app/ui/invoices/breadcrumbs';

export default async function Page() {

  return (
    <main>
      <Breadcrumbs
        breadcrumbs={[
          { label: 'Strategies', href: '/dashboard/strategies' },
          {
            label: 'Add Strategy',
            href: '/dashboard/strategies/create',
            active: true,
          },
        ]}
      />
      <Form/>
    </main>
  );
}